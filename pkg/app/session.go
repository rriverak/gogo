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

//PostNewSession Handles the PostRequest to create a new Session
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

//GetSession Handles the GetRequest to Join a Session
func (d *SessionController) GetSession(w http.ResponseWriter, r *http.Request) {
	//Get ViewData
	data := utils.GetViewData(r)

	// Get Parameter
	sessionID := mux.Vars(r)["id"]

	// Set Session
	data["Session"] = d.SessionManager.GetSession(sessionID)
	// Render Page
	utils.GetPageTemplate(d.TemplateBasePath, "views/app/session.page.html").Execute(w, data)
}

//PostSDPSession Handles the SDP for the Session
func (d *SessionController) PostSDPSession(w http.ResponseWriter, r *http.Request) {
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
