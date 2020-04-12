package api

import (
	"fmt"
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
	sub.HandleFunc("/{id}/{user}", s.JoiningSessions).Methods("POST")
}

//getSessionKey formatted
func (s *SessionHandler) getSessionKey(sessID string) string {
	return fmt.Sprintf("session-%v", sessID)
}

//getSession Get from Cache or Create a new Session
func (s *SessionHandler) getSession(sessID string, userName string) *rtc.Session {
	var session *rtc.Session
	sessKey := s.getSessionKey(sessID)
	sess, sessionFound := s.SessionRegister.Get(sessKey)
	if sessionFound {
		Logger.Infof("User => %v join Session with Key => '%v' \n", userName, sessKey)
		session = sess.(*rtc.Session)
	} else {
		Logger.Infof("User => %v create a new Session with Key => '%v' \n", userName, sessKey)
		session = rtc.NewSession(sessKey)
	}
	return session
}

//saveSession Add the Session in Cache or Replace it.
func (s *SessionHandler) saveSession(sessID string, session *rtc.Session) error {
	var err error
	if _, found := s.SessionRegister.Get(s.getSessionKey(sessID)); !found {
		err = s.SessionRegister.Add(s.getSessionKey(sessID), session, cache.NoExpiration)
	} else {
		err = s.SessionRegister.Replace(s.getSessionKey(sessID), session, cache.NoExpiration)
	}
	if err != nil {
		panic(err)
	}
	return err
}

//ListSessions Handles a HTTP Get to List all Sessions with Users
func (s *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, s.SessionRegister.Items())
}

//JoiningSessions Handles the Joining Offer
func (s *SessionHandler) JoiningSessions(w http.ResponseWriter, r *http.Request) {
	// HTTP VARs
	sessID := mux.Vars(r)["id"]
	userName := mux.Vars(r)["user"]

	// Get or Create a Session
	var session *rtc.Session = s.getSession(sessID, userName)

	// Get the offer from Body
	offer := webrtc.SessionDescription{}
	body, _ := ioutil.ReadAll(r.Body)
	signal.Decode(string(body), &offer)

	// Create User from Session
	newUser, err := session.CreateUser(userName, peerConnectionConfig, offer)

	// Set the remote SessionDescription for User
	err = newUser.Peer.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}

	// Create answer for User
	answer, err := newUser.Peer.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners for User
	err = newUser.Peer.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	// Save the current Session in Cache.
	err = s.saveSession(sessID, session)
	if err != nil {
		panic(err)
	}

	// Write Awnser to Client
	WriteJSON(w, answer)
}
