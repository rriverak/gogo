package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rriverak/gogo/internal/mgt"
	"github.com/rriverak/gogo/internal/rtc"
	"github.com/rriverak/gogo/internal/utils"
)

//NewUsersController Controller
func NewUsersController(sessionManager *rtc.SessionManager, navBuilder utils.NavBuilder, userRepo mgt.Repository) UsersController {
	return UsersController{SessionManager: sessionManager, NavBuilder: navBuilder, ViewBasePath: "./web/templates/", UserRepository: userRepo}
}

//UsersController for Admin Dashboard
type UsersController struct {
	ViewBasePath   string
	NavBuilder     utils.NavBuilder
	SessionManager *rtc.SessionManager
	UserRepository mgt.Repository
}

//GetList Handles the GetRequest in this Controller
func (d *UsersController) GetList(w http.ResponseWriter, r *http.Request) {
	rUsers, _ := d.UserRepository.List()

	data := utils.GetViewData(r)
	data["Users"] = rUsers
	data["SideNav"] = d.NavBuilder.GetNavigation("Users")

	utils.GetPageTemplate(d.ViewBasePath, "views/admin/users.page.html").Execute(w, data)
}

//GetUser Handles the GetRequest in this Controller
func (d *UsersController) GetUser(w http.ResponseWriter, r *http.Request) {
	var user *mgt.User = d.getUser(r)

	data := utils.GetViewData(r)
	data["User"] = user
	data["SideNav"] = d.NavBuilder.GetNavigation("Users")

	utils.GetPageTemplate(d.ViewBasePath, "views/admin/users-detail.page.html").Execute(w, data)
}

//PostUser Handles the PostRequest in this Controller
func (d *UsersController) PostUser(w http.ResponseWriter, r *http.Request) {
	var user *mgt.User = d.getUser(r)

	r.ParseForm()
	name := r.Form.Get("name")
	if len(name) < 3 {
		utils.WriteError(w, fmt.Errorf("Name to short... %v", name))
		return
	}
	password := r.Form.Get("password")
	if len(password) < 5 {
		utils.WriteError(w, fmt.Errorf("Password to short... %v", name))
		return
	}

	if user != nil {
		user.UserName = name
		user.ChangePassword(password)
		err := d.UserRepository.Update(&user)
		if err != nil {
			utils.WriteError(w, err)
			return
		}
	} else {
		newUser := mgt.NewUser(name, password)
		err := d.UserRepository.Insert(&newUser)
		if err != nil {
			utils.WriteError(w, err)
			return
		}
	}

	//Redirect to Index Page
	w.Header().Set("Location", "/admin/users")
	w.WriteHeader(http.StatusSeeOther)
}

func (d *UsersController) getID(r *http.Request) (int64, error) {
	id := mux.Vars(r)["id"]
	return strconv.ParseInt(id, 10, 64)
}

func (d *UsersController) getUser(r *http.Request) *mgt.User {
	var user *mgt.User
	id, err := d.getID(r)
	if err == nil {
		//id found
		res, err := d.UserRepository.ByID(id)
		if err == nil {
			//user found
			usr := res.(mgt.User)
			user = &usr
		}
	}
	return user
}
