package auth

import (
	"context"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/rriverak/gogo/internal/mgt"
	"github.com/rriverak/gogo/internal/utils"
	"github.com/wader/gormstore"
)

//Middleware is the Type
type Middleware struct {
	SessionStore *gormstore.Store
	CsrfKey      []byte
	UserRepo     mgt.Repository
}

//CsfrMiddleware for Authentication and UserID in Context
func (m *Middleware) CsfrMiddleware(next http.Handler) http.Handler {
	csrf.Secure(false)
	return csrf.Protect(m.CsrfKey, csrf.CookieName("csrf"))(next)
}

//SessionMiddleware for Authentication and UserID in Context
func (m *Middleware) SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := m.SessionStore.Get(r, "session")
		//If Session is OK
		if err == nil && session != nil {
			//If a UserID in Session
			if userID, ok := session.Values["user_id"]; ok {
				//Append User in Request Context
				user, err := m.UserRepo.ByID(int64(userID.(uint)))
				if user != nil && err == nil {
					//Serve Next
					ctx := context.WithValue(r.Context(), utils.ContextKeyType{}, user)
					next.ServeHTTP(w, r.WithContext(ctx))
					//Skip rest of function
					return
				}
			}
		}
		// Default Denied
		// MaxAge<0 means delete cookie immediately.
		session.Options.MaxAge = -1
		// Save to Response
		session.Save(r, w)
		// Write Forbidden or Redirect
		utils.WriteRedirect(w, "/login")
	})
}
