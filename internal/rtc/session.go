package rtc

import (
	"github.com/pion/webrtc/v2"
	"github.com/rriverak/gogo/internal/gst"
)

//Session is a GroupVideo Call
type Session struct {
	ID            string      `json:"ID"`
	API           *webrtc.API `json:"-"`
	Codec         string
	VideoTrack    *webrtc.Track `json:"-"`
	AudioTrack    *webrtc.Track `json:"-"`
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

//AddUser to Session and restart Pipeline
func (s *Session) AddUser(newUser User) {
	// Register Users RemoteTrack with Session
	newUser.Peer.OnTrack(newUser.RemoteTrackHandler(s))

	// Add user to Collection
	if s.Users == nil {
		s.Users = make([]User, 0)
	}
	s.Users = append(s.Users, newUser)

	// Restart Session Pipeline
	s.Restart()
}

//RemoveUser from Session and restart Pipeline
func (s *Session) RemoveUser(userName string) {
	if s.Users == nil {
		s.Users = make([]User, 0)
	}
	tmpUsers := []User{}
	for _, usr := range s.Users {
		if usr.Name != userName {
			tmpUsers = append(tmpUsers, usr)
		}
	}
	s.Users = tmpUsers

	if len(s.Users) != 0 {
		s.Restart()
	} else {
		s.Stop()
	}
}
