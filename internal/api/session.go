package api

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v2"
	"github.com/rriverak/gogo/internal/gst"
	"github.com/rriverak/gogo/internal/signal"
	"github.com/rriverak/gogo/internal/utils"
)

var peerConnectionConfig webrtc.Configuration = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	},
}

const (
	rtcpPLIInterval = time.Second * 1
)

//Session is a GroupVideo Call
type Session struct {
	ID            string        `json:"ID"`
	API           *webrtc.API   `json:"-"`
	VideoTrack    *webrtc.Track `json:"-"`
	VideoPipeline *gst.Pipeline `json:"-"`
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
	s.VideoPipeline = gst.CreateVideoMixerPipeline(webrtc.VP8, chans)

	// Set Pipeline output
	s.VideoPipeline.SetOutputTrack(s.VideoTrack)

	// Start Pipeline output
	s.VideoPipeline.Start()
}

//Restart a Session with new Parameters
func (s *Session) Restart() {
	// Stop Running Pipeline
	if s.VideoPipeline != nil {
		s.VideoPipeline.Stop()
	}
	s.Start()
}

//NewUser creates a new User
func NewUser(name string) User {
	return User{ID: utils.RandSeq(5), Name: name}
}

//User can Connect to a Session
type User struct {
	ID         string
	Name       string
	AudioTrack *webrtc.Track
	Peer       *webrtc.PeerConnection
}

//SessionHandler handles API Requests for Sessions
type SessionHandler struct {
	SessionRegister *cache.Cache
	MediaEngine     *webrtc.MediaEngine
}

//RegisterSessionRoutes apply all Routes to the Router
func (s *SessionHandler) RegisterSessionRoutes(r *mux.Router) {
	s.SessionRegister = cache.New(cache.NoExpiration, cache.NoExpiration)
	sub := r.PathPrefix("/api/sessions").Subrouter()
	sub.HandleFunc("/", s.ListSessions).Methods("GET")
	sub.HandleFunc("/{id}/{user}", s.JoiningSessions).Methods("POST")
}

//ListSessions Handles a HTTP Get to List all Sessions with Users
func (s *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, s.SessionRegister.Items())
}

//GetSessionKey for Cache Repository
func (s *SessionHandler) GetSessionKey(sessID string) string {
	return fmt.Sprintf("session-%v", sessID)
}

//JoiningSessions Handles the Joining Offer
func (s *SessionHandler) JoiningSessions(w http.ResponseWriter, r *http.Request) {
	// HTTP VARs
	sessID := mux.Vars(r)["id"]
	userName := mux.Vars(r)["user"]

	// Get or Create a Session
	var session *Session
	sessKey := s.GetSessionKey(sessID)
	sess, sessionFound := s.SessionRegister.Get(sessKey)
	if sessionFound {
		Logger.Infof("Join Session with Key => '%v' \n", sessKey)
		session = sess.(*Session)
	} else {
		Logger.Infof("Create new Session with Key => '%v' \n", sessKey)
		session = &Session{ID: sessID}
	}

	// Get the offer
	offer := webrtc.SessionDescription{}
	body, _ := ioutil.ReadAll(r.Body)
	signal.Decode(string(body), &offer)

	// Prepare Engine
	err := s.MediaEngine.PopulateFromSDP(offer)
	if err != nil {
		panic(err)
	}

	// Create a API
	if session.API == nil {
		session.API = webrtc.NewAPI(webrtc.WithMediaEngine(*s.MediaEngine))
	}

	// Create a Peer
	peerConnection, err := session.API.NewPeerConnection(peerConnectionConfig)
	if err != nil {
		panic(err)
	}

	// Add User to List
	if session.Users == nil {
		session.Users = make([]User, 0)
	}
	newUser := NewUser(userName)
	newUser.Peer = peerConnection
	session.Users = append(session.Users, newUser)
	Logger.Infof("Users in Session => %v", session.Users)

	// Allow the Peer to send a Video Stream
	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}

	// Create a new Mixed Track if not exists
	if session.VideoTrack == nil {
		mixedTrack, newTrackErr := peerConnection.NewTrack(webrtc.DefaultPayloadTypeVP8, rand.Uint32(), "video", "mixed")
		if newTrackErr != nil {
			panic(newTrackErr)
		}
		session.VideoTrack = mixedTrack
	}

	// Add VideoMixed Track to Peer
	if _, err = peerConnection.AddTrack(session.VideoTrack); err != nil {
		panic(err)
	}
	session.Restart()

	// On Peer Conncetion Timeout or Disconnected
	peerConnection.OnConnectionStateChange(func(f webrtc.PeerConnectionState) {
		Logger.Infof("User '%v' State Changed => %v", userName, f.String())
		if f == webrtc.PeerConnectionStateDisconnected {
			Logger.Infof("User => %v Timeout or Disconnected", userName)
			users := []User{}
			for _, usr := range session.Users {
				if usr.Name != userName {
					users = append(users, usr)
				}
			}
			session.Users = users
			session.Restart()
		}
	})
	// Set a handler for when a new remote track starts by our Peer
	peerConnection.OnTrack(func(remoteTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {
		// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
		// This can be less wasteful by processing incoming RTCP events, then we would emit a NACK/PLI when a viewer requests it
		go func() {
			ticker := time.NewTicker(rtcpPLIInterval)
			for range ticker.C {
				if rtcpSendErr := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: remoteTrack.SSRC()}}); rtcpSendErr != nil {
					Logger.Error(rtcpSendErr)
				}
			}
		}()
		// Create a Buffer Loop
		rtpBuf := make([]byte, 1400)
		for {
			// Read remote Buffer
			i, readErr := remoteTrack.Read(rtpBuf)
			if readErr != nil {
				panic(readErr)
			}
			// Push RTP Samples to GStreamer Pipeline with specific appsrc (user_id)
			session.VideoPipeline.WriteSampleToInputSource(rtpBuf[:i], fmt.Sprintf("src-%v", userName))
		}
	})

	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}

	// Create answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	// Add the Session in Cache or Replace it.
	if _, found := s.SessionRegister.Get(s.GetSessionKey(sessID)); !found {
		err = s.SessionRegister.Add(s.GetSessionKey(sessID), session, cache.NoExpiration)
	} else {
		err = s.SessionRegister.Replace(s.GetSessionKey(sessID), session, cache.NoExpiration)
	}
	if err != nil {
		panic(err)
	}

	// Write Awnser to Client
	WriteJSON(w, answer)
}
