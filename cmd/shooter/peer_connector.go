package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/pion/webrtc/v4"
)

const signalPollInterval = time.Second

var errPeerSessionExpired = errors.New("peer session expired")

type PeerConnector interface {
	MenuLabel() string
	StatusLine(counter uint64) string
	ServerURL() string
	RoomID() string
	SetServerURL(string)
	SetRoomID(string)
	Action() error
}

type peerConnector struct {
	mutex sync.Mutex

	peerConnection *webrtc.PeerConnection
	dataChannel    *webrtc.DataChannel

	serverBaseURL         string
	roomIdentifier        string
	participantIdentifier string
	roomRole              string
	sessionNumber         uint64

	lastServerBaseURL  string
	lastRoomIdentifier string
	statusLine         string
	statusPending      bool
}

type peerJoinRoomRequest struct {
	RoomIdentifier string `json:"room_identifier"`
}

type peerJoinRoomResponse struct {
	ParticipantIdentifier string `json:"participant_identifier"`
	Role                  string `json:"role"`
}

type peerSignalingRequest struct {
	RoomIdentifier        string                     `json:"room_identifier"`
	ParticipantIdentifier string                     `json:"participant_identifier"`
	SessionDescription    *webrtc.SessionDescription `json:"session_description,omitempty"`
}

type peerSignalingResponse struct {
	SessionDescription webrtc.SessionDescription `json:"session_description"`
}

func newPeerConnector() PeerConnector {
	return &peerConnector{
		lastServerBaseURL: "http://localhost:8500",
		statusLine:        "Peer: idle",
	}
}

func (connector *peerConnector) MenuLabel() string {
	connector.mutex.Lock()
	defer connector.mutex.Unlock()

	if connector.serverBaseURL != "" || connector.peerConnection != nil {
		return "Disconnect peer"
	}

	return "Connect to peer"
}

func (connector *peerConnector) StatusLine(counter uint64) string {
	connector.mutex.Lock()
	defer connector.mutex.Unlock()

	if !connector.statusPending {
		return connector.statusLine
	}

	frames := []string{"|", "/", "-", "\\"}
	return connector.statusLine + " " + frames[(counter/12)%uint64(len(frames))]
}

func (connector *peerConnector) ServerURL() string {
	connector.mutex.Lock()
	defer connector.mutex.Unlock()
	return connector.lastServerBaseURL
}

func (connector *peerConnector) RoomID() string {
	connector.mutex.Lock()
	defer connector.mutex.Unlock()
	return connector.lastRoomIdentifier
}

func (connector *peerConnector) SetServerURL(serverBaseURL string) {
	connector.mutex.Lock()
	defer connector.mutex.Unlock()
	connector.lastServerBaseURL = normalizeServerBaseURL(serverBaseURL)
}

func (connector *peerConnector) SetRoomID(roomIdentifier string) {
	connector.mutex.Lock()
	defer connector.mutex.Unlock()
	connector.lastRoomIdentifier = strings.TrimSpace(roomIdentifier)
}

func (connector *peerConnector) Action() error {
	if connector.hasActiveSession() {
		go connector.finishSession("Peer: disconnected")
		return nil
	}

	serverBaseURL := normalizeServerBaseURL(connector.ServerURL())
	if serverBaseURL == "" {
		connector.setStatus("Peer: signaling server URL is required")
		return nil
	}

	roomIdentifier := strings.TrimSpace(connector.RoomID())
	if roomIdentifier == "" {
		connector.setStatus("Peer: room ID is required")
		return nil
	}

	if err := connector.finishSession(""); err != nil {
		connector.setStatus("Peer: " + err.Error())
		return nil
	}

	connector.rememberDefaults(serverBaseURL, roomIdentifier)
	sessionNumber := connector.startSession(serverBaseURL, roomIdentifier)
	connector.setPendingStatus(fmt.Sprintf("Peer: joining room %q", roomIdentifier))

	go connector.runConnect(sessionNumber, serverBaseURL, roomIdentifier)
	return nil
}

func (connector *peerConnector) runConnect(sessionNumber uint64, serverBaseURL string, roomIdentifier string) {
	joinResponse, err := connector.joinRoom(serverBaseURL, roomIdentifier)
	if err != nil {
		_ = connector.finishSession("")
		connector.setStatus("Peer: " + err.Error())
		return
	}

	if !connector.isCurrentSession(sessionNumber) {
		_ = connector.leaveRoom(serverBaseURL, roomIdentifier, joinResponse.ParticipantIdentifier)
		return
	}

	connector.setRoomMembership(joinResponse.ParticipantIdentifier, joinResponse.Role)
	connector.setPendingStatus(fmt.Sprintf("Peer: joined as %s", joinResponse.Role))

	if joinResponse.Role == "offerer" {
		err = connector.connectAsOfferer(sessionNumber)
	} else {
		err = connector.connectAsAnswerer(sessionNumber)
	}

	if errors.Is(err, errPeerSessionExpired) {
		return
	}
	if err != nil {
		_ = connector.finishSession("")
		connector.setStatus("Peer: " + err.Error())
	}
}

func (connector *peerConnector) connectAsOfferer(sessionNumber uint64) error {
	connector.setPendingStatus("Peer: creating offer")

	peerConnection, err := connector.newPeerConnection()
	if err != nil {
		return err
	}

	if !connector.isCurrentSession(sessionNumber) {
		_ = peerConnection.Close()
		return errPeerSessionExpired
	}

	dataChannel, err := peerConnection.CreateDataChannel("game", nil)
	if err != nil {
		return err
	}
	connector.attachDataChannel(peerConnection, dataChannel)

	offerDescription, err := peerConnection.CreateOffer(nil)
	if err != nil {
		return err
	}
	if err := peerConnection.SetLocalDescription(offerDescription); err != nil {
		return err
	}

	connector.setPendingStatus("Peer: gathering offer candidates")
	<-webrtc.GatheringCompletePromise(peerConnection)

	if !connector.isCurrentSession(sessionNumber) {
		return errPeerSessionExpired
	}

	localDescription := peerConnection.LocalDescription()
	if localDescription == nil {
		return errors.New("local offer was not generated")
	}

	if err := connector.publishSessionDescription("offer", localDescription); err != nil {
		return err
	}

	connector.setPendingStatus("Peer: offer sent, waiting for answer")

	answerDescription, err := connector.waitForSessionDescription(sessionNumber, "answer")
	if err != nil {
		return err
	}
	if answerDescription.Type != webrtc.SDPTypeAnswer {
		return fmt.Errorf("expected answer, got %s", answerDescription.Type.String())
	}
	if err := peerConnection.SetRemoteDescription(answerDescription); err != nil {
		return err
	}

	connector.setPendingStatus("Peer: answer received, opening data channel")
	return nil
}

func (connector *peerConnector) connectAsAnswerer(sessionNumber uint64) error {
	connector.setPendingStatus("Peer: waiting for offer")

	offerDescription, err := connector.waitForSessionDescription(sessionNumber, "offer")
	if err != nil {
		return err
	}
	if offerDescription.Type != webrtc.SDPTypeOffer {
		return fmt.Errorf("expected offer, got %s", offerDescription.Type.String())
	}

	connector.setPendingStatus("Peer: offer received, creating answer")
	peerConnection, err := connector.newPeerConnection()
	if err != nil {
		return err
	}

	if !connector.isCurrentSession(sessionNumber) {
		_ = peerConnection.Close()
		return errPeerSessionExpired
	}

	if err := peerConnection.SetRemoteDescription(offerDescription); err != nil {
		return err
	}

	answerDescription, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return err
	}
	if err := peerConnection.SetLocalDescription(answerDescription); err != nil {
		return err
	}

	connector.setPendingStatus("Peer: gathering answer candidates")
	<-webrtc.GatheringCompletePromise(peerConnection)

	if !connector.isCurrentSession(sessionNumber) {
		return errPeerSessionExpired
	}

	localDescription := peerConnection.LocalDescription()
	if localDescription == nil {
		return errors.New("local answer was not generated")
	}

	if err := connector.publishSessionDescription("answer", localDescription); err != nil {
		return err
	}

	connector.setPendingStatus("Peer: answer sent, waiting for peer")
	return nil
}

func (connector *peerConnector) waitForSessionDescription(sessionNumber uint64, signalKind string) (webrtc.SessionDescription, error) {
	for {
		if !connector.isCurrentSession(sessionNumber) {
			return webrtc.SessionDescription{}, errPeerSessionExpired
		}

		sessionDescription, found, err := connector.fetchSessionDescription(signalKind)
		if err != nil {
			return webrtc.SessionDescription{}, err
		}
		if found {
			return sessionDescription, nil
		}

		time.Sleep(signalPollInterval)
	}
}

func (connector *peerConnector) newPeerConnection() (*webrtc.PeerConnection, error) {
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
			{URLs: []string{"stun:stun1.l.google.com:19302"}},
		},
	})
	if err != nil {
		return nil, err
	}

	connector.mutex.Lock()
	connector.peerConnection = peerConnection
	connector.dataChannel = nil
	connector.mutex.Unlock()

	peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		if !connector.isCurrentPeerConnection(peerConnection) {
			return
		}

		switch state {
		case webrtc.PeerConnectionStateConnected:
			connector.setStatus(fmt.Sprintf("Peer: connected as %s", connector.currentRole()))
		case webrtc.PeerConnectionStateConnecting:
			connector.setPendingStatus("Peer: connecting")
		case webrtc.PeerConnectionStateDisconnected:
			connector.setStatus("Peer: disconnected")
		case webrtc.PeerConnectionStateFailed:
			connector.setStatus("Peer: connection failed")
		case webrtc.PeerConnectionStateClosed:
			connector.setStatus("Peer: connection closed")
		default:
			connector.setStatus("Peer: " + strings.ToLower(state.String()))
		}
	})

	peerConnection.OnDataChannel(func(dataChannel *webrtc.DataChannel) {
		if !connector.isCurrentPeerConnection(peerConnection) {
			return
		}
		connector.attachDataChannel(peerConnection, dataChannel)
	})

	return peerConnection, nil
}

func (connector *peerConnector) attachDataChannel(peerConnection *webrtc.PeerConnection, dataChannel *webrtc.DataChannel) {
	connector.mutex.Lock()
	if connector.peerConnection != peerConnection {
		connector.mutex.Unlock()
		return
	}
	connector.dataChannel = dataChannel
	connector.mutex.Unlock()

	dataChannel.OnOpen(func() {
		if !connector.isCurrentDataChannel(dataChannel) {
			return
		}
		connector.setStatus(fmt.Sprintf("Peer: connected as %s", connector.currentRole()))
	})

	dataChannel.OnClose(func() {
		if !connector.isCurrentDataChannel(dataChannel) {
			return
		}
		connector.clearDataChannel(dataChannel)
		connector.setStatus("Peer: data channel closed")
	})

	dataChannel.OnError(func(err error) {
		if !connector.isCurrentDataChannel(dataChannel) {
			return
		}
		connector.setStatus("Peer: " + err.Error())
	})

	dataChannel.OnMessage(func(message webrtc.DataChannelMessage) {
		_ = message
	})

	connector.setPendingStatus("Peer: data channel attached, waiting to open")
}

func (connector *peerConnector) finishSession(note string) error {
	connector.mutex.Lock()
	currentPeerConnection := connector.peerConnection
	currentServerBaseURL := connector.serverBaseURL
	currentRoomIdentifier := connector.roomIdentifier
	currentParticipantIdentifier := connector.participantIdentifier
	connector.peerConnection = nil
	connector.dataChannel = nil
	connector.serverBaseURL = ""
	connector.roomIdentifier = ""
	connector.participantIdentifier = ""
	connector.roomRole = ""
	connector.sessionNumber++
	if note != "" {
		connector.statusLine = note
	} else {
		connector.statusLine = "Peer: idle"
	}
	connector.statusPending = false
	connector.mutex.Unlock()

	if currentParticipantIdentifier != "" {
		_ = connector.leaveRoom(currentServerBaseURL, currentRoomIdentifier, currentParticipantIdentifier)
	}

	if currentPeerConnection != nil {
		if err := currentPeerConnection.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (connector *peerConnector) rememberDefaults(serverBaseURL string, roomIdentifier string) {
	connector.mutex.Lock()
	defer connector.mutex.Unlock()
	connector.lastServerBaseURL = serverBaseURL
	connector.lastRoomIdentifier = roomIdentifier
}

func (connector *peerConnector) startSession(serverBaseURL string, roomIdentifier string) uint64 {
	connector.mutex.Lock()
	defer connector.mutex.Unlock()

	connector.sessionNumber++
	connector.serverBaseURL = serverBaseURL
	connector.roomIdentifier = roomIdentifier
	connector.participantIdentifier = ""
	connector.roomRole = ""
	return connector.sessionNumber
}

func (connector *peerConnector) setRoomMembership(participantIdentifier string, roomRole string) {
	connector.mutex.Lock()
	defer connector.mutex.Unlock()
	connector.participantIdentifier = participantIdentifier
	connector.roomRole = roomRole
}

func (connector *peerConnector) joinRoom(serverBaseURL string, roomIdentifier string) (peerJoinRoomResponse, error) {
	requestBody := peerJoinRoomRequest{RoomIdentifier: roomIdentifier}
	var responseBody peerJoinRoomResponse

	_, err := connector.performJSONRequest(
		http.MethodPost,
		serverBaseURL+"/api/rooms/join",
		requestBody,
		&responseBody,
	)
	if err != nil {
		return peerJoinRoomResponse{}, err
	}

	return responseBody, nil
}

func (connector *peerConnector) leaveRoom(serverBaseURL string, roomIdentifier string, participantIdentifier string) error {
	if serverBaseURL == "" || roomIdentifier == "" || participantIdentifier == "" {
		return nil
	}

	_, err := connector.performJSONRequest(
		http.MethodPost,
		serverBaseURL+"/api/rooms/leave",
		peerSignalingRequest{
			RoomIdentifier:        roomIdentifier,
			ParticipantIdentifier: participantIdentifier,
		},
		nil,
	)
	return err
}

func (connector *peerConnector) publishSessionDescription(signalKind string, sessionDescription *webrtc.SessionDescription) error {
	serverBaseURL, roomIdentifier, participantIdentifier := connector.currentSignalingTarget()
	if serverBaseURL == "" || roomIdentifier == "" || participantIdentifier == "" {
		return errors.New("peer signaling session is not active")
	}

	_, err := connector.performJSONRequest(
		http.MethodPut,
		serverBaseURL+"/api/rooms/"+signalKind,
		peerSignalingRequest{
			RoomIdentifier:        roomIdentifier,
			ParticipantIdentifier: participantIdentifier,
			SessionDescription:    sessionDescription,
		},
		nil,
	)
	return err
}

func (connector *peerConnector) fetchSessionDescription(signalKind string) (webrtc.SessionDescription, bool, error) {
	serverBaseURL, roomIdentifier, _ := connector.currentSignalingTarget()
	if serverBaseURL == "" || roomIdentifier == "" {
		return webrtc.SessionDescription{}, false, errPeerSessionExpired
	}

	requestURL := fmt.Sprintf(
		"%s/api/rooms/%s?room_identifier=%s",
		serverBaseURL,
		signalKind,
		url.QueryEscape(roomIdentifier),
	)

	var responseBody peerSignalingResponse
	statusCode, err := connector.performJSONRequest(http.MethodGet, requestURL, nil, &responseBody)
	if err != nil {
		if statusCode == http.StatusNotFound {
			return webrtc.SessionDescription{}, false, nil
		}
		return webrtc.SessionDescription{}, false, err
	}

	return responseBody.SessionDescription, true, nil
}

func (connector *peerConnector) performJSONRequest(method string, requestURL string, requestBody any, responseBody any) (int, error) {
	var requestReader io.Reader
	if requestBody != nil {
		payload, err := json.Marshal(requestBody)
		if err != nil {
			return 0, err
		}
		requestReader = bytes.NewReader(payload)
	}

	request, err := http.NewRequest(method, requestURL, requestReader)
	if err != nil {
		return 0, err
	}
	if requestBody != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		payload, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			return response.StatusCode, fmt.Errorf("request failed with status %d", response.StatusCode)
		}

		message := strings.TrimSpace(string(payload))
		if message == "" {
			message = http.StatusText(response.StatusCode)
		}

		return response.StatusCode, errors.New(message)
	}

	if responseBody != nil {
		if err := json.NewDecoder(response.Body).Decode(responseBody); err != nil {
			return response.StatusCode, err
		}
	}

	return response.StatusCode, nil
}

func (connector *peerConnector) hasActiveSession() bool {
	connector.mutex.Lock()
	defer connector.mutex.Unlock()
	return connector.serverBaseURL != "" || connector.peerConnection != nil
}

func (connector *peerConnector) currentSignalingTarget() (string, string, string) {
	connector.mutex.Lock()
	defer connector.mutex.Unlock()
	return connector.serverBaseURL, connector.roomIdentifier, connector.participantIdentifier
}

func (connector *peerConnector) currentRole() string {
	connector.mutex.Lock()
	defer connector.mutex.Unlock()
	if connector.roomRole == "" {
		return "peer"
	}
	return connector.roomRole
}

func (connector *peerConnector) isCurrentPeerConnection(peerConnection *webrtc.PeerConnection) bool {
	connector.mutex.Lock()
	defer connector.mutex.Unlock()
	return connector.peerConnection == peerConnection
}

func (connector *peerConnector) isCurrentDataChannel(dataChannel *webrtc.DataChannel) bool {
	connector.mutex.Lock()
	defer connector.mutex.Unlock()
	return connector.dataChannel == dataChannel
}

func (connector *peerConnector) isCurrentSession(sessionNumber uint64) bool {
	connector.mutex.Lock()
	defer connector.mutex.Unlock()
	return connector.sessionNumber == sessionNumber
}

func (connector *peerConnector) clearDataChannel(dataChannel *webrtc.DataChannel) {
	connector.mutex.Lock()
	defer connector.mutex.Unlock()
	if connector.dataChannel == dataChannel {
		connector.dataChannel = nil
	}
}

func (connector *peerConnector) setStatus(status string) {
	connector.mutex.Lock()
	defer connector.mutex.Unlock()
	connector.statusLine = status
	connector.statusPending = false
}

func (connector *peerConnector) setPendingStatus(status string) {
	connector.mutex.Lock()
	defer connector.mutex.Unlock()
	connector.statusLine = status
	connector.statusPending = true
}

func normalizeServerBaseURL(rawServerBaseURL string) string {
	serverBaseURL := strings.TrimSpace(rawServerBaseURL)
	if serverBaseURL == "" {
		return ""
	}
	if !strings.HasPrefix(serverBaseURL, "http://") && !strings.HasPrefix(serverBaseURL, "https://") {
		serverBaseURL = "http://" + serverBaseURL
	}
	return strings.TrimRight(serverBaseURL, "/")
}
