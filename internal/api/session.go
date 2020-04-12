package api

import (
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/pion/webrtc/v2"
	"github.com/rriverak/gogo/internal/rtc"
	"github.com/rriverak/gogo/internal/signal"
)

var peerConnectionConfig webrtc.Configuration = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	},
}

//SessionHandler handles API Requests for Sessions
type SessionHandler struct {
	SessionRegister *cache.Cache
	MediaEngine     *webrtc.MediaEngine
}

//RegisterSessionRoutes apply all Routes to the Router
func (s *SessionHandler) RegisterSessionRoutes(r *mux.Router) {
	s.SessionRegister = cache.New(cache.NoExpiration, cache.NoExpiration)
	sub := r.PathPrefix("/api/sessions").Subrouter()
	sub.HandleFunc("/", s.ListSessions).Methods("GET")
	sub.HandleFunc("/{id}", s.DeleteSession).Methods("DELETE")
	sub.HandleFunc("/{id}/{user}", s.JoinOrCreateSessions).Methods("POST")
}

//getSession Get from Cache or Create a new Session
func (s *SessionHandler) getSession(sessionID string) *rtc.Session {
	var session *rtc.Session
	sess, sessionFound := s.SessionRegister.Get(sessionID)
	if sessionFound {
		session = sess.(*rtc.Session)
	} else {
		session = rtc.NewSession()
	}
	return session
}

//saveSession Add the Session in Cache or Replace it.
func (s *SessionHandler) saveSession(session *rtc.Session) error {
	var err error
	if _, found := s.SessionRegister.Get(session.ID); !found {
		err = s.SessionRegister.Add(session.ID, session, cache.NoExpiration)
	} else {
		err = s.SessionRegister.Replace(session.ID, session, cache.NoExpiration)
	}
	return err
}

//removeSession remove the Session from Cache if found. Returns true if found.
func (s *SessionHandler) removeSession(sessionID string) bool {
	if sess, found := s.SessionRegister.Get(sessionID); found {
		session := sess.(*rtc.Session)
		// Disconnect all Users
		for _, usr := range session.Users {
			session.DisconnectUser(&usr)
		}
		// Remove Session from Cache
		s.SessionRegister.Delete(sessionID)
		return true
	}
	return false
}

//ListSessions Handles a HTTP Get to List all Sessions with Users
func (s *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, s.SessionRegister.Items())
}

//DeleteSession Handles a HTTP DELETE to Delete a Sessions and drop there Users
func (s *SessionHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	sessKey := mux.Vars(r)["id"]
	if s.removeSession(sessKey) {
		WriteStatusOK(w)
	} else {
		Logger.Infof("Session not found! '%v'", sessKey)
		WriteStatusNotFound(w)
	}
}

//JoinOrCreateSessions Handles the Joining Offer
func (s *SessionHandler) JoinOrCreateSessions(w http.ResponseWriter, r *http.Request) {
	// HTTP VARs
	sessionID := mux.Vars(r)["id"]
	userName := mux.Vars(r)["user"]

	// Get or Create a Session
	var session *rtc.Session = s.getSession(sessionID)
	Logger.Infof("User => %v Create Session with ID => '%v' \n", userName, sessionID)

	// Get the offer from Body
	offer := webrtc.SessionDescription{}
	body, _ := ioutil.ReadAll(r.Body)
	signal.Decode(string(body), &offer)

	// Create User from Session
	newUser, err := session.CreateUser(userName, peerConnectionConfig, offer)
	answer := newUser.Anwser(offer)
	// Save the current Session in Cache.
	err = s.saveSession(session)
	if err != nil {
		panic(err)
	}

	// Write Awnser to Client
	WriteJSON(w, answer)
}
