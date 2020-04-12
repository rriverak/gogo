package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
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

	// Add User to List
	newUser, err := rtc.NewUser(userName, peerConnectionConfig, offer)

	for _, cdec := range newUser.MediaEngine.GetCodecsByKind(webrtc.RTPCodecTypeVideo) {
		Logger.Infof("User: %v Video Codec: %v PayloadType: %v Clock: %v", userName, cdec.Name, cdec.PayloadType, cdec.ClockRate)
	}
	for _, cdec := range newUser.MediaEngine.GetCodecsByKind(webrtc.RTPCodecTypeAudio) {
		Logger.Infof("User: %v Audio Codec: %v PayloadType: %v Clock: %v", userName, cdec.Name, cdec.PayloadType, cdec.ClockRate)
	}

	// Create a new Mixed Video Track if not exists
	vOut := newUser.VideOutput()
	Logger.Infof("Create Output VideoMix: Code: %v Payload: %v", vOut.Codec().Name, vOut.Codec().PayloadType)

	// Create a new Mixed Audio Track if not exists
	aOut := newUser.AudioOutput()
	Logger.Infof("Create Output AudioMix: Code: %v Payload: %v", aOut.Codec().Name, aOut.Codec().PayloadType)

	// On Peer Conncetion Timeout or Disconnected
	newUser.Peer.OnConnectionStateChange(func(f webrtc.PeerConnectionState) {
		Logger.Infof("User '%v' State Changed => %v", userName, f.String())
		if f == webrtc.PeerConnectionStateDisconnected {
			Logger.Infof("User => %v Timeout or Disconnected", userName)
			session.RemoveUser(userName)
		}
	})

	// DataChannel
	newUser.Peer.OnDataChannel(func(d *webrtc.DataChannel) {
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
				err := newUser.Peer.Close()
				if err != nil {
					Logger.Error(err)
				}
			}
		})
	})

	// Add User to Session and Restart Pipeline
	session.Codec = newUser.Codec
	session.AddUser(*newUser)

	// Set the remote SessionDescription
	err = newUser.Peer.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}

	// Create answer
	answer, err := newUser.Peer.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	err = newUser.Peer.SetLocalDescription(answer)
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
