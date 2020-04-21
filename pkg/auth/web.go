package auth

import (
	"net/http"

	"github.com/rriverak/gogo/internal/mgt"
	"github.com/rriverak/gogo/internal/utils"
	"github.com/wader/gormstore"
)

//NewUsersController Controller
func newWebLoginController(userRepo mgt.Repository, sessionStore *gormstore.Store) loginController {
	return loginController{
		ViewBasePath:   "./web/templates/",
		UserRepository: userRepo,
		SessionStore:   sessionStore,
	}
}

//UsersController for Admin Dashboard
type loginController struct {
	ViewBasePath   string
	UserRepository mgt.Repository
	SessionStore   *gormstore.Store
}

type loginCredentials struct {
	Username string
	Password string
}

//GetLogin Handles the GetRequest in this Controller
func (l *loginController) GetLogin(w http.ResponseWriter, r *http.Request) {
	data := utils.GetViewData(r)
	utils.GetPlainTemplate(l.ViewBasePath, "views/auth/login.html").Execute(w, data)
}

//PostLogin Handles the GetRequest in this Controller
func (l *loginController) PostLogin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	allUsers, _ := l.UserRepository.List()
	for _, iusr := range allUsers {
		usr := iusr.(mgt.User)
		if usr.UserName == username {
			if isPasswordOk, _ := usr.CheckPassword(password); isPasswordOk {
				sess, _ := l.SessionStore.Get(r, "session")
				sess.Values["user_id"] = usr.ID
				err := sess.Save(r, w)
				if err == nil {
					//Redirect to Index Page
					w.Header().Set("Location", "/")
					w.WriteHeader(http.StatusSeeOther)
					return
				}
			}
		}
	}
	//Redirect to Index Page
	w.Header().Set("Location", "/login?error=true")
	w.WriteHeader(http.StatusSeeOther)
}

//GetLogin Handles the GetRequest in this Controller
func (l *loginController) GetLogout(w http.ResponseWriter, r *http.Request) {
	sess, _ := l.SessionStore.Get(r, "session")
	sess.Options.MaxAge = -1
	sess.Save(r, w)
	//Redirect to Index Page
	w.Header().Set("Location", "/login")
	w.WriteHeader(http.StatusSeeOther)
}
