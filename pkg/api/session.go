package api

import (
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pion/webrtc/v2"
	"github.com/rriverak/gogo/internal/rtc"
	"github.com/rriverak/gogo/internal/signal"
	"github.com/rriverak/gogo/internal/utils"
)

//SessionHandler handles API Requests for Sessions
type SessionHandler struct {
	SessionManager *rtc.SessionManager
}

//RegisterSessionRoutes apply all Routes to the Router
func (s *SessionHandler) RegisterSessionRoutes(r *mux.Router) {
	r.HandleFunc("/", s.ListSessionsHandler).Methods("GET")
	r.HandleFunc("/{id}", s.DeleteSessionHandler).Methods("DELETE")
	r.HandleFunc("/{user}", s.CreateSessionHandler).Methods("POST")
	r.HandleFunc("/{id}/join/{user}", s.JoinSessionHandler).Methods("POST")
}

//ListSessionsHandler Handles a HTTP Get to List all Sessions with Users
func (s *SessionHandler) ListSessionsHandler(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, s.SessionManager.GetAllSessions())
}

//CreateSessionHandler Create a new Session if not exists
func (s *SessionHandler) CreateSessionHandler(w http.ResponseWriter, r *http.Request) {
	// HTTP VARs
	sessionName := mux.Vars(r)["user"]

	// Create Session
	session := s.SessionManager.NewSession(sessionName)

	Logger.Infof("Create Session with ID => '%v'", session.ID)
	utils.WriteStatusOK(w)
}

//JoinSessionHandler Handles the Joining Offer
func (s *SessionHandler) JoinSessionHandler(w http.ResponseWriter, r *http.Request) {
	// HTTP VARs
	sessionID := mux.Vars(r)["id"]
	userName := mux.Vars(r)["user"]

	// Get a Session
	var session *rtc.Session = s.SessionManager.GetSession(sessionID)

	if session == nil {
		utils.WriteStatusConfict(w)
		return
	}
	// Get the offer from Body
	offer := webrtc.SessionDescription{}
	body, _ := ioutil.ReadAll(r.Body)
	signal.Decode(string(body), &offer)

	// Create Participant from Session
	newPart, err := session.CreateParticipant(userName, offer)
	answer := newPart.Anwser(offer)

	// Save the current Session in Cache.
	err = s.SessionManager.SaveSession(session)
	if err != nil {
		panic(err)
	}

	// Write Awnser to Client
	utils.WriteJSON(w, answer)
}

//DeleteSessionHandler Handles a HTTP DELETE to Delete a Sessions and drop there Users
func (s *SessionHandler) DeleteSessionHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := mux.Vars(r)["id"]
	if s.SessionManager.RemoveSession(sessionID) {
		Logger.Infof("Delete Session with ID => '%v'", sessionID)
		utils.WriteStatusOK(w)
	} else {
		utils.WriteStatusNotFound(w)
	}
}
