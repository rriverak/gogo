package app

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pion/webrtc/v2"
	"github.com/rriverak/gogo/internal/rtc"
	"github.com/rriverak/gogo/internal/signal"
	"github.com/rriverak/gogo/internal/utils"
)

//NewSessionController Controller
func NewSessionController(sessionManager *rtc.SessionManager) SessionController {
	return SessionController{SessionManager: sessionManager, TemplateBasePath: "./web/templates/"}
}

//SessionController for Admin Dashboard
type SessionController struct {
	TemplateBasePath string
	SessionManager   *rtc.SessionManager
}

//PostNewSession Handles the GetRequest in this Controller
func (d *SessionController) PostNewSession(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	// Create Session if Name gte 3
	name := r.Form.Get("name")
	if len(name) >= 3 {
		d.SessionManager.NewSession(name)
	}
	//Redirect to Index Page
	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusSeeOther)
}

//GetSession Handles the GetRequest in this Controller
func (d *SessionController) GetSession(w http.ResponseWriter, r *http.Request) {
	// userID := r.Context().Value(auth.SessionContextKey).(uint)

	// Get Parameter
	sessionID := mux.Vars(r)["id"]

	// Declare PageData
	session := d.SessionManager.GetSession(sessionID)

	// Getting From Values
	// Setup Page Rendering
	data := map[string]interface{}{
		"Session": session,
	}
	utils.GetPageTemplate(d.TemplateBasePath, "views/app/session.page.html").Execute(w, data)
}

//PostSDPSession Handles the GetRequest in this Controller
func (d *SessionController) PostSDPSession(w http.ResponseWriter, r *http.Request) {
	// userID := r.Context().Value(auth.SessionContextKey).(uint)

	// Get Parameter
	sessionID := mux.Vars(r)["id"]

	// Post Payload
	offer := webrtc.SessionDescription{}
	body, _ := ioutil.ReadAll(r.Body)
	signal.Decode(string(body), &offer)

	// Get Session
	session := d.SessionManager.GetSession(sessionID)

	// Proceed Joining
	if session != nil {
		// Create a Session User
		randUserName := fmt.Sprintf("User-%v", utils.RandSeq(3))
		user, _ := session.CreateParticipant(randUserName, offer)

		// Getting the Answer
		answer := signal.Encode(user.Anwser(offer))

		// Save the current Session in Cache.
		d.SessionManager.SaveSession(session)

		// Write Answer
		utils.WriteText(w, answer)
	} else {
		// Session not found
		utils.WriteStatusNotFound(w)
	}

}
