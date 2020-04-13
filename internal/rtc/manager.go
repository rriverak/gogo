package rtc

import (
	"github.com/patrickmn/go-cache"
	"github.com/rriverak/gogo/internal/config"
)

//SessionManager manage Sessions
type SessionManager struct {
	Config          *config.Config
	SessionRegister *cache.Cache
}

//NewSession Get from Cache or Create a new Session
func (s *SessionManager) NewSession(name string) *Session {
	var session *Session = newSession(name)
	s.SaveSession(session)
	return session
}

//GetAllSessions from Manager
func (s *SessionManager) GetAllSessions() []*Session {
	all := []*Session{}
	for _, sess := range s.SessionRegister.Items() {
		all = append(all, sess.Object.(*Session))
	}
	return all
}

//GetSession Get from Cache or Create a new Session
func (s *SessionManager) GetSession(sessionID string) *Session {
	var session *Session
	sess, sessionFound := s.SessionRegister.Get(sessionID)
	if sessionFound {
		session = sess.(*Session)
	}
	return session
}

//SaveSession Add the Session in Cache or Replace it.
func (s *SessionManager) SaveSession(session *Session) error {
	var err error
	if _, found := s.SessionRegister.Get(session.ID); !found {
		err = s.SessionRegister.Add(session.ID, session, cache.NoExpiration)
	} else {
		err = s.SessionRegister.Replace(session.ID, session, cache.NoExpiration)
	}
	return err
}

//RemoveSession remove the Session from Cache if found. Returns true if found.
func (s *SessionManager) RemoveSession(sessionID string) bool {
	if sess, found := s.SessionRegister.Get(sessionID); found {
		session := sess.(*Session)
		// Disconnect all Users
		for _, usr := range session.Users {
			session.DisconnectUser(&usr)
		}
		// Remove Session from Cache
		s.SessionRegister.Delete(sessionID)
		return true
	}
	return false
}
