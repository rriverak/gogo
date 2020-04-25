package admin

import (
	"net/http"

	"github.com/rriverak/gogo/internal/rtc"
	"github.com/rriverak/gogo/internal/utils"
)

//NewDashboardController Controller
func NewDashboardController(sessionManager *rtc.SessionManager, navBuilder utils.NavBuilder) DashboardController {
	return DashboardController{SessionManager: sessionManager, NavBuilder: navBuilder, ViewBasePath: "./web/templates/"}
}

//DashboardController for Admin Dashboard
type DashboardController struct {
	ViewBasePath   string
	NavBuilder     utils.NavBuilder
	SessionManager *rtc.SessionManager
}

//Get Handles the GetRequest in this Controller
func (d *DashboardController) Get(w http.ResponseWriter, r *http.Request) {
	activeUsers := 0
	activeSessions := 0
	for _, sess := range d.SessionManager.GetAllSessions() {
		activeUsers += len(sess.Participants)
		activeSessions++
	}
	data := utils.GetViewData(r)
	data["SideNav"] = d.NavBuilder.GetNavigation("Dashboard")
	data["ActiveUsers"] = activeUsers
	data["ActiveSessions"] = activeSessions

	utils.GetPageTemplate(d.ViewBasePath, "views/admin/dashboard.page.html").Execute(w, data)
}
