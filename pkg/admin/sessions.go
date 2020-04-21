package admin

import (
	"net/http"

	"github.com/rriverak/gogo/internal/rtc"
	"github.com/rriverak/gogo/internal/utils"
)

//NewSessionsController Controller
func NewSessionsController(sessionManager *rtc.SessionManager, navBuilder utils.NavBuilder) SessionsController {
	return SessionsController{SessionManager: sessionManager, NavBuilder: navBuilder, ViewBasePath: "./web/templates/"}
}

//SessionsController for Admin Dashboard
type SessionsController struct {
	ViewBasePath   string
	NavBuilder     utils.NavBuilder
	SessionManager *rtc.SessionManager
}

//Get Handles the GetRequest in this Controller
func (d *SessionsController) Get(w http.ResponseWriter, r *http.Request) {

	data := utils.GetViewData(r)
	data["Sessions"] = d.SessionManager.GetAllSessions()
	data["SideNav"] = d.NavBuilder.GetNavigation("Sessions")
	utils.GetPageTemplate(d.ViewBasePath, "views/admin/sessions.page.html").Execute(w, data)
}
