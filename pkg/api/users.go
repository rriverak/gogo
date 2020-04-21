package api

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rriverak/gogo/internal/mgt"
	"github.com/rriverak/gogo/internal/utils"
)

//UsersHandler handles API Requests for Users
type UsersHandler struct {
	UserRepo mgt.Repository
}

//RegisterUsersRoutes apply all Routes to the Router
func (u *UsersHandler) RegisterUsersRoutes(r *mux.Router) {
	r.HandleFunc("/{id}", u.GetByHandler).Methods("GET")
	r.HandleFunc("/{id}", u.UpdateHandler).Methods("POST")
	r.HandleFunc("/{id}", u.DeleteHandler).Methods("DELETE")
	r.HandleFunc("/", u.CreateHandler).Methods("POST")
	r.HandleFunc("/", u.ListHandler).Methods("GET")
}

//ListHandler Handles a HTTP Get to List all Users
func (u *UsersHandler) ListHandler(w http.ResponseWriter, r *http.Request) {
	users, err := u.UserRepo.List()
	if err == nil {
		utils.WriteJSON(w, users)

	} else {
		utils.WriteError(w, err)
	}
}

//GetByHandler Handles a HTTP Get to a User
func (u *UsersHandler) GetByHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	id64, err := strconv.ParseInt(id, 10, 64)
	user, err := u.UserRepo.ByID(id64)
	utils.WriteResultOrError(w, user, err)
}

//CreateHandler Handles a HTTP Post to create a User
func (u *UsersHandler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	user := mgt.User{}
	if err := utils.DecodeBody(r, &user); err != nil {
		utils.WriteError(w, err)
	} else {
		user.PasswordHash = utils.GenerateHash(user.PasswordHash)
		err := u.UserRepo.Insert(&user)
		utils.WriteResultOrError(w, user, err)
	}
}

//UpdateHandler Handles a HTTP Post to update a User
func (u *UsersHandler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	user := mgt.User{}
	id := mux.Vars(r)["id"]
	id64, err := strconv.ParseInt(id, 10, 64)

	if err = utils.DecodeBody(r, &user); err != nil {
		utils.WriteError(w, err)
	} else {
		rUsr, err := u.UserRepo.ByID(id64)
		if err == nil && rUsr != nil {
			user.ID = rUsr.(mgt.User).ID
			err = u.UserRepo.Update(&user)
		}
		utils.WriteResultOrError(w, user, err)
	}

}

//DeleteHandler Handles a HTTP Delete to delete a User
func (u *UsersHandler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	user := mgt.User{}
	id := mux.Vars(r)["id"]
	id64, err := strconv.ParseInt(id, 10, 64)
	if err = utils.DecodeBody(r, &user); err != nil {
		utils.WriteError(w, err)
	} else {

		rUsr, err := u.UserRepo.ByID(id64)
		if err == nil && rUsr != nil {
			err = u.UserRepo.Delete(&rUsr)
		}
		utils.WriteResultOrError(w, user, err)
	}
}
