package api

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/pion/rtcp"
	"github.com/pion/sdp/v2"
	"github.com/pion/webrtc/v2"
	"github.com/rriverak/gogo/internal/rtc"
	"github.com/rriverak/gogo/internal/signal"
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
	var session *rtc.Session
	sessKey := s.GetSessionKey(sessID)
	sess, sessionFound := s.SessionRegister.Get(sessKey)
	if sessionFound {
		Logger.Infof("Join Session with Key => '%v' \n", sessKey)
		session = sess.(*rtc.Session)
	} else {
		Logger.Infof("Create new Session with Key => '%v' \n", sessKey)
		session = &rtc.Session{ID: sessID}
	}

	// Get the offer
	offer := webrtc.SessionDescription{}
	body, _ := ioutil.ReadAll(r.Body)
	signal.Decode(string(body), &offer)

	Logger.Info(offer)
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
	newUser := rtc.NewUser(userName)
	newUser.Peer = peerConnection
	session.AddUser(newUser)
	Logger.Infof("Users in Session => %v", session.Users)

	// Allow the Peer to send a Video Stream
	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}

	// Allow the Peer to send a Audio Stream
	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
		panic(err)
	}

	// Create a new Mixed Video Track if not exists
	if session.VideoTrack == nil {
		mixedVideoTrack, newTrackErr := peerConnection.NewTrack(webrtc.DefaultPayloadTypeVP8, rand.Uint32(), "video", "video-mixed")
		if newTrackErr != nil {
			panic(newTrackErr)
		}
		session.VideoTrack = mixedVideoTrack
	}

	// Create a new Mixed Audio Track if not exists
	if session.AudioTrack == nil {
		mixedAudioTrack, newTrackErr := peerConnection.NewTrack(webrtc.DefaultPayloadTypeOpus, rand.Uint32(), "audio", "audio-mixed")
		if newTrackErr != nil {
			panic(newTrackErr)
		}
		session.AudioTrack = mixedAudioTrack
	}

	// Add VideoMixed Track to Peer
	if _, err = peerConnection.AddTrack(session.VideoTrack); err != nil {
		panic(err)
	}
	// Add AudioMixed Track to Peer
	if _, err = peerConnection.AddTrack(session.AudioTrack); err != nil {
		panic(err)
	}

	session.Restart()

	// On Peer Conncetion Timeout or Disconnected
	peerConnection.OnConnectionStateChange(func(f webrtc.PeerConnectionState) {
		Logger.Infof("User '%v' State Changed => %v", userName, f.String())
		if f == webrtc.PeerConnectionStateDisconnected {
			Logger.Infof("User => %v Timeout or Disconnected", userName)
			session.RemoveUser(userName)
		}
	})

	// DataChannel
	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		Logger.Infof("New DataChannel %s %d", d.Label(), d.ID())
		// Register channel opening handling
		d.OnOpen(func() {
			// Send Open Messages
			sendErr := d.SendText("open")
			if sendErr != nil {
				panic(sendErr)
			}
		})
		// Register text message handling
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			message := string(msg.Data)
			Logger.Infof("Message from DataChannel '%s': '%s'", d.Label(), message)
			if d.Label() == "data" && message == "close" {
				Logger.Infof("User => %v close the Session", userName)
				session.RemoveUser(userName)
				err := peerConnection.Close()
				if err != nil {
					Logger.Error(err)
				}
			}
		})
	})

	// Set a handler for when a new remote track starts by our Peer
	peerConnection.OnTrack(func(remoteTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {
		if remoteTrack.PayloadType() == webrtc.DefaultPayloadTypeVP8 || remoteTrack.PayloadType() == webrtc.DefaultPayloadTypeH264 {
			// Video Track
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
				session.VideoPipeline.WriteSampleToInputSource(rtpBuf[:i], newUser.ID)
			}
		} else if remoteTrack.PayloadType() == webrtc.DefaultPayloadTypeOpus {
			// Audio Track
			// Create a Buffer Loop
			rtpBuf := make([]byte, 1400)
			for {
				// Read remote Buffer
				i, readErr := remoteTrack.Read(rtpBuf)
				if readErr != nil {
					panic(readErr)
				}
				// Push RTP Samples to GStreamer Pipeline with specific appsrc (user_id)
				session.AudioPipeline.WriteSampleToInputSource(rtpBuf[:i], newUser.ID)
			}
		} else {
			Logger.Infof("NoTrack => %v", remoteTrack.PayloadType())
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

// firstCodecOfType returns the first codec of a chosen type from a session description
func firstCodecOfType(sd webrtc.SessionDescription, codecName string) (*sdp.Codec, error) {
	sdpsd := sdp.SessionDescription{}
	err := sdpsd.Unmarshal([]byte(sd.SDP))
	if err != nil {
		return nil, err
	}
	for _, md := range sdpsd.MediaDescriptions {
		for _, format := range md.MediaName.Formats {
			pt, err := strconv.Atoi(format)
			if err != nil {
				return nil, fmt.Errorf("format parse error")
			}
			payloadType := uint8(pt)
			payloadCodec, err := sdpsd.GetCodecForPayloadType(payloadType)
			if err != nil {
				return nil, fmt.Errorf("could not find codec for payload type %d", payloadType)
			}
			if payloadCodec.Name == codecName {
				return &payloadCodec, nil
			}
		}
	}
	return nil, fmt.Errorf("no codec of type %s found in SDP", codecName)
}
