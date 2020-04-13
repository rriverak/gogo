package rtc

import (
	"github.com/pion/webrtc/v2"
	"github.com/rriverak/gogo/internal/config"
	"github.com/rriverak/gogo/internal/gst"
	"github.com/rriverak/gogo/internal/utils"
)

//NewSession create a new Session
func newSession(name string) *Session {
	return &Session{ID: utils.RandSeq(5), Name: name}
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
	config        *config.Config
}

//Start a Session with new Parameters
func (s *Session) Start() {
	// Create Pipeline Channel
	chans := []string{}
	for _, usr := range s.Users {
		chans = append(chans, usr.ID)
	}

	if s.config.Media.Video.Enabled {
		// Create GStreamer Video Pipeline
		s.VideoPipeline = gst.CreateVideoMixerPipeline(s.Codec, chans)
	}

	if s.config.Media.Audio.Enabled {
		// Create GStreamer Audio Pipeline
		s.AudioPipeline = gst.CreateAudioMixerPipeline(webrtc.Opus, chans)
	}

	for _, usr := range s.Users {
		if s.VideoPipeline != nil {
			s.VideoPipeline.AddOutputTrack(usr.VideoOutput())
		}
		if s.AudioPipeline != nil {
			s.AudioPipeline.AddOutputTrack(usr.AudioOutput())
		}
	}

	// Start Pipeline output
	if s.VideoPipeline != nil {
		s.VideoPipeline.Start()
	}
	if s.AudioPipeline != nil {
		s.AudioPipeline.Start()
	}
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
func (s *Session) CreateUser(name string, offer webrtc.SessionDescription) (*User, error) {
	// Get MediaEngine
	customPayloadType, codec, media := GetMediaEngineForSDPOffer(offer, s.config.Media)

	//  Create New User with Peer
	var peerConnectionConfig webrtc.Configuration = webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: s.config.WebRTC.ICEServers,
			},
		},
	}
	newUser, err := NewUser(name, peerConnectionConfig, media, customPayloadType, codec)
	if err != nil {
		return nil, err
	}

	if s.config.Media.Video.Enabled {
		// Log Codec
		for _, cdec := range media.GetCodecsByKind(webrtc.RTPCodecTypeVideo) {
			Logger.Infof("User => %v offer => Video Codec: %v PayloadType: %v Clock: %v", name, cdec.Name, cdec.PayloadType, cdec.ClockRate)
		}
		// Allow the Peer to send a Video Stream
		if _, err = newUser.Peer.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
			panic(err)
		}
		// Add VideoMixed Track to Peer
		if _, err = newUser.Peer.AddTrack(newUser.VideoOutput()); err != nil {
			panic(err)
		}
	}

	if s.config.Media.Audio.Enabled {
		// Log Codec
		for _, cdec := range media.GetCodecsByKind(webrtc.RTPCodecTypeAudio) {
			Logger.Infof("User => %v offer => Audio Codec: %v PayloadType: %v Clock: %v", name, cdec.Name, cdec.PayloadType, cdec.ClockRate)
		}
		// Allow the Peer to send a Audio Stream
		if _, err = newUser.Peer.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
			panic(err)
		}
		// Add AudioMixed Track to Peer
		if _, err = newUser.Peer.AddTrack(newUser.AudioOutput()); err != nil {
			panic(err)
		}
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
