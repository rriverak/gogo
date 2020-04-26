package rtc

import (
	"encoding/json"

	"github.com/pion/webrtc/v2"
	"github.com/rriverak/gogo/internal/config"
	"github.com/rriverak/gogo/internal/gst"
	"github.com/rriverak/gogo/internal/utils"
)

//NewSession create a new Session
func newSession(name string, cfg *config.Config) *Session {
	return &Session{ID: utils.RandSeq(5), Name: name, config: cfg}
}

//Session is a GroupVideo Call
type Session struct {
	ID            string        `json:"ID"`
	Name          string        `json:"Name"`
	API           *webrtc.API   `json:"-"`
	Codec         string        `json:"Codec"`
	VideoPipeline *gst.Pipeline `json:"-"`
	AudioPipeline *gst.Pipeline `json:"-"`
	Participants  []Participant `json:"Participants"`
	config        *config.Config
}

//Start a Session with new Parameters
func (s *Session) Start() {
	// Create Pipeline Channels
	channels := []*gst.Channel{}
	for _, usr := range s.Participants {
		channels = append(channels, gst.NewChannel(usr.ID, usr.Name))
	}

	if s.config.Media.Video.Enabled {
		// Create GStreamer Video Pipeline
		s.VideoPipeline = gst.CreateVideoMixerPipeline(s.Codec, channels, s.config)
	}

	if s.config.Media.Audio.Enabled {
		// Create GStreamer Audio Pipeline
		s.AudioPipeline = gst.CreateAudioMixerPipeline(webrtc.Opus, channels, s.config)
	}

	// Start Pipeline output
	if s.VideoPipeline != nil {
		s.VideoPipeline.Start()
	}
	if s.AudioPipeline != nil {
		s.AudioPipeline.Start()
	}

	for _, usr := range s.Participants {
		if s.VideoPipeline != nil {
			s.VideoPipeline.AddOutputTrack(usr.VideoOutput())
		}
		if s.AudioPipeline != nil {
			s.AudioPipeline.AddOutputTrack(usr.AudioOutput())
		}
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

// CreateParticipant in the Session
func (s *Session) CreateParticipant(name string, offer webrtc.SessionDescription) (*Participant, error) {
	// Get MediaEngine
	mediaCfg := s.config.Media
	customPayloadType, codec, media := GetMediaEngineForSDPOffer(offer, mediaCfg)

	//  Create New Participant with Peer
	var peerConnectionConfig webrtc.Configuration = webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: s.config.WebRTC.ICEServers,
			},
		},
	}
	newPart, err := NewParticipant(name, peerConnectionConfig, media, customPayloadType, codec)
	if err != nil {
		return nil, err
	}

	if s.config.Media.Video.Enabled {
		// Log Codec
		for _, cdec := range media.GetCodecsByKind(webrtc.RTPCodecTypeVideo) {
			Logger.Infof("Participant => %v offer => Video Codec: %v PayloadType: %v Clock: %v", name, cdec.Name, cdec.PayloadType, cdec.ClockRate)
		}
		// Allow the Peer to send a Video Stream
		if _, err = newPart.Peer.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
			panic(err)
		}
		// Add VideoMixed Track to Peer
		if _, err = newPart.Peer.AddTrack(newPart.VideoOutput()); err != nil {
			panic(err)
		}
	}

	if s.config.Media.Audio.Enabled {
		// Log Codec
		for _, cdec := range media.GetCodecsByKind(webrtc.RTPCodecTypeAudio) {
			Logger.Infof("Participant => %v offer Audio Codec: %v PayloadType: %v Clock: %v", name, cdec.Name, cdec.PayloadType, cdec.ClockRate)
		}
		// Allow the Peer to send a Audio Stream
		if _, err = newPart.Peer.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
			panic(err)
		}
		// Add AudioMixed Track to Peer
		if _, err = newPart.Peer.AddTrack(newPart.AudioOutput()); err != nil {
			panic(err)
		}
	}

	// Register Participants RemoteTrack with Session
	newPart.Peer.OnTrack(newPart.OnRemoteTrackHandler(s))
	// Register Session Auto-Leave on Timeout
	newPart.Peer.OnConnectionStateChange(newPart.OnParticipantConnectionStateChangedHandler(s))
	s.Codec = newPart.Codec

	// Add Participant to Session
	s.AddParticipant(*newPart)

	// Register Session DataChannel
	Logger.Info("Create Session DC")
	dc, err := newPart.Peer.CreateDataChannel("session", nil)
	if err != nil {
		Logger.Error(err)
	}
	dc.OnMessage(newPart.OnParticipantSessionMessage(s))
	dc.OnOpen(func() {
		s.BroadcastState()
	})
	dc.OnError(func(err error) {
		Logger.Error(err)
	})
	newPart.DataChannels["Session"] = dc
	return newPart, nil
}

//BroadcastState to the DataChannel
func (s *Session) BroadcastState() {
	for _, part := range s.Participants {
		if dc := part.DataChannels["Session"]; dc != nil {
			//get current state
			state := map[string]interface{}{}
			state["Users"] = s.Participants
			state["ID"] = s.ID
			state["Name"] = s.Name
			//Send state
			data, _ := json.Marshal(state)
			dc.SendText(string(data))
		}
	}
}

//AddParticipant to Session and restart Pipeline
func (s *Session) AddParticipant(newPart Participant) {
	// Add Participant to Collection
	if s.Participants == nil {
		s.Participants = make([]Participant, 0)
	}
	s.Participants = append(s.Participants, newPart)

	// Restart Session Pipeline
	s.Restart()
}

//RemoveParticipant from with ID from Session and restart Pipeline
func (s *Session) RemoveParticipant(usrID string) {
	if s.Participants == nil {
		s.Participants = make([]Participant, 0)
	}
	tmpParts := []Participant{}
	for _, usr := range s.Participants {
		if usr.ID != usrID {
			tmpParts = append(tmpParts, usr)
		}
	}
	s.Participants = tmpParts

	if len(s.Participants) != 0 {
		s.Restart()
	} else {
		s.Stop()
		Logger.Info("Session Stoped.")
	}
}

// DisconnectParticipant from Session
func (s *Session) DisconnectParticipant(part *Participant) {
	s.RemoveParticipant(part.ID) // Remove from Session
	part.Peer.Close()            // Close peer Connection
}
