package rtc

import (
	"github.com/pion/webrtc/v2"
	"github.com/rriverak/gogo/internal/gst"
	"github.com/rriverak/gogo/internal/utils"
)

//NewSession create a new Session
func NewSession() *Session {
	return &Session{ID: utils.RandSeq(5)}
}

//Session is a GroupVideo Call
type Session struct {
	ID            string        `json:"ID"`
	Name          string        `json:"Name"`
	API           *webrtc.API   `json:"-"`
	Codec         string        `json:"Codec"`
	VideoPipeline *gst.Pipeline `json:"-"`
	AudioPipeline *gst.Pipeline `json:"-"`
	Users         []User        `json:"Users"`
}

//Start a Session with new Parameters
func (s *Session) Start() {
	// Create Pipeline Channel
	chans := []string{}
	for _, usr := range s.Users {
		chans = append(chans, usr.ID)
	}

	// Create GStreamer Pipeline
	s.VideoPipeline = gst.CreateVideoMixerPipeline(s.Codec, chans)

	// Create GStreamer Pipeline
	s.AudioPipeline = gst.CreateAudioMixerPipeline(webrtc.Opus, chans)

	for _, usr := range s.Users {
		s.VideoPipeline.AddOutputTrack(usr.VideOutput())
		s.AudioPipeline.AddOutputTrack(usr.AudioOutput())
	}

	// Start Pipeline output
	s.VideoPipeline.Start()
	s.AudioPipeline.Start()
}

//Stop a Session
func (s *Session) Stop() {
	// Stop Running Pipeline
	if s.VideoPipeline != nil {
		// Set Locking
		s.VideoPipeline.Stop()
		s.VideoPipeline = nil
	}
	if s.AudioPipeline != nil {
		// Set Locking
		s.AudioPipeline.Stop()
		s.AudioPipeline = nil
	}

}

//Restart a Session with new Parameters
func (s *Session) Restart() {
	s.Stop()
	s.Start()
}

// CreateUser in the Session
func (s *Session) CreateUser(name string, peerConnectionConfig webrtc.Configuration, offer webrtc.SessionDescription) (*User, error) {
	// Create New User with Peer
	newUser, err := NewUser(name, peerConnectionConfig, offer)
	if err != nil {
		return nil, err
	}
	// Register Users RemoteTrack with Session
	newUser.Peer.OnTrack(newUser.OnRemoteTrackHandler(s))
	// Register Session Auto-Leave on Timeout
	newUser.Peer.OnConnectionStateChange(newUser.OnUserConnectionStateChangedHandler(s))
	s.Codec = newUser.Codec
	// Register Session DataChannel
	var id uint16 = 1
	negotiated := false
	opt := webrtc.DataChannelInit{Negotiated: &negotiated, ID: &id}
	dc, err := newUser.Peer.CreateDataChannel("session", &opt)
	if err != nil {
		Logger.Error(err)
	}
	dc.OnMessage(newUser.OnUserSessionMessage(s))

	// Add User to Session
	s.AddUser(*newUser)
	return newUser, nil
}

//AddUser to Session and restart Pipeline
func (s *Session) AddUser(newUser User) {
	// Add user to Collection
	if s.Users == nil {
		s.Users = make([]User, 0)
	}
	s.Users = append(s.Users, newUser)

	// Restart Session Pipeline
	s.Restart()
}

//RemoveUser from with ID from Session and restart Pipeline
func (s *Session) RemoveUser(usrID string) {
	if s.Users == nil {
		s.Users = make([]User, 0)
	}
	tmpUsers := []User{}
	for _, usr := range s.Users {
		if usr.ID != usrID {
			tmpUsers = append(tmpUsers, usr)
		}
	}
	s.Users = tmpUsers

	if len(s.Users) != 0 {
		s.Restart()
	} else {
		s.Stop()
		Logger.Info("Session Stoped.")
	}
}

// DisconnectUser a User from Session
func (s *Session) DisconnectUser(user *User) {
	s.RemoveUser(user.ID) // Remove from Session
	user.Peer.Close()     // Close peer Connection
}
