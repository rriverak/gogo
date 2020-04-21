package app

import (
	"net/http"

	"github.com/rriverak/gogo/internal/rtc"
	"github.com/rriverak/gogo/internal/utils"
)

//NewStartController Controller
func NewStartController(sessionManager *rtc.SessionManager) StartController {
	return StartController{SessionManager: sessionManager, TemplateBasePath: "./web/templates/"}
}

//StartController for Admin Dashboard
type StartController struct {
	TemplateBasePath string
	SessionManager   *rtc.SessionManager
}

//Get Handles the GetRequest in this Controller
func (d *StartController) Get(w http.ResponseWriter, r *http.Request) {
	data := utils.GetViewData(r)
	data["Sessions"] = d.SessionManager.GetAllSessions()
	utils.GetPageTemplate(d.TemplateBasePath, "views/app/start.page.html").Execute(w, data)
}
