package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const maxRoomIdentifierLength = 100
const roomExpirationWindow = 60 * time.Second
const roomCleanupInterval = 5 * time.Second

type signalingServer struct {
	mutex sync.Mutex
	rooms map[string]*roomState
}

type roomState struct {
	participantRoles map[string]string
	offerPayload     json.RawMessage
	answerPayload    json.RawMessage
	lastPingAt       time.Time
}

type joinRoomRequest struct {
	RoomIdentifier string `json:"room_identifier"`
}

type joinRoomResponse struct {
	ParticipantIdentifier string `json:"participant_identifier"`
	Role                  string `json:"role"`
}

type signalingRequest struct {
	RoomIdentifier        string          `json:"room_identifier"`
	ParticipantIdentifier string          `json:"participant_identifier"`
	SessionDescription    json.RawMessage `json:"session_description,omitempty"`
}

type signalingResponse struct {
	SessionDescription json.RawMessage `json:"session_description"`
}

var (
	errRoomFull            = errors.New("room is full")
	errRoomNotFound        = errors.New("room was not found")
	errParticipantNotFound = errors.New("participant was not found")
	errRoleMismatch        = errors.New("participant role does not match this signal type")
	errRoomIdentifierTooLong = errors.New("room_identifier must be 100 characters or fewer")
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	address := flag.String("addr", ":8500", "server listen address")
	flag.Parse()

	server := &signalingServer{
		rooms: map[string]*roomState{},
	}

	multiplexer := http.NewServeMux()
	multiplexer.HandleFunc("/api/rooms/join", server.handleJoinRoom)
	multiplexer.HandleFunc("/api/rooms/leave", server.handleLeaveRoom)
	multiplexer.HandleFunc("/api/rooms/ping", server.handlePingRoom)
	multiplexer.HandleFunc("/api/rooms/offer", server.handleOffer)
	multiplexer.HandleFunc("/api/rooms/answer", server.handleAnswer)

	go server.cleanupExpiredRoomsLoop()

	log.Printf("signaling server listening on %s", *address)
	log.Fatal(http.ListenAndServe(*address, withCommonHeaders(multiplexer)))
}

func withCommonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.Header().Set("Access-Control-Allow-Origin", "*")
		responseWriter.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		responseWriter.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")

		if request.Method == http.MethodOptions {
			responseWriter.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(responseWriter, request)
	})
}

func (server *signalingServer) handleJoinRoom(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(responseWriter, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var joinRequest joinRoomRequest
	if err := json.NewDecoder(request.Body).Decode(&joinRequest); err != nil {
		http.Error(responseWriter, "invalid join request", http.StatusBadRequest)
		return
	}

	roomIdentifier, err := validateRoomIdentifier(joinRequest.RoomIdentifier)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	joinResponse, err := server.joinRoom(roomIdentifier)
	if err != nil {
		if errors.Is(err, errRoomFull) {
			http.Error(responseWriter, err.Error(), http.StatusConflict)
			return
		}
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(responseWriter, joinResponse)
}

func (server *signalingServer) handleLeaveRoom(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(responseWriter, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var leaveRequest signalingRequest
	if err := json.NewDecoder(request.Body).Decode(&leaveRequest); err != nil {
		http.Error(responseWriter, "invalid leave request", http.StatusBadRequest)
		return
	}

	server.leaveRoom(leaveRequest.RoomIdentifier, leaveRequest.ParticipantIdentifier)
	responseWriter.WriteHeader(http.StatusNoContent)
}

func (server *signalingServer) handlePingRoom(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(responseWriter, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var pingRequest joinRoomRequest
	if err := json.NewDecoder(request.Body).Decode(&pingRequest); err != nil {
		http.Error(responseWriter, "invalid ping request", http.StatusBadRequest)
		return
	}

	roomIdentifier, err := validateRoomIdentifier(pingRequest.RoomIdentifier)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	if err := server.pingRoom(roomIdentifier, time.Now()); err != nil {
		if errors.Is(err, errRoomNotFound) {
			http.Error(responseWriter, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	responseWriter.WriteHeader(http.StatusNoContent)
}

func (server *signalingServer) handleOffer(responseWriter http.ResponseWriter, request *http.Request) {
	server.handleSignal(responseWriter, request, "offer")
}

func (server *signalingServer) handleAnswer(responseWriter http.ResponseWriter, request *http.Request) {
	server.handleSignal(responseWriter, request, "answer")
}

func (server *signalingServer) handleSignal(responseWriter http.ResponseWriter, request *http.Request, signalKind string) {
	switch request.Method {
	case http.MethodGet:
		roomIdentifier, err := validateRoomIdentifier(request.URL.Query().Get("room_identifier"))
		if err != nil {
			http.Error(responseWriter, err.Error(), http.StatusBadRequest)
			return
		}

		sessionDescription, found := server.getSignal(roomIdentifier, signalKind)
		if !found {
			http.Error(responseWriter, "signal not available", http.StatusNotFound)
			return
		}

		writeJSON(responseWriter, signalingResponse{SessionDescription: sessionDescription})
	case http.MethodPut:
		var signalRequest signalingRequest
		if err := json.NewDecoder(request.Body).Decode(&signalRequest); err != nil {
			http.Error(responseWriter, "invalid signaling request", http.StatusBadRequest)
			return
		}

		if err := server.setSignal(signalKind, signalRequest); err != nil {
			switch {
			case errors.Is(err, errRoomNotFound):
				http.Error(responseWriter, err.Error(), http.StatusNotFound)
			case errors.Is(err, errParticipantNotFound):
				http.Error(responseWriter, err.Error(), http.StatusNotFound)
			case errors.Is(err, errRoleMismatch):
				http.Error(responseWriter, err.Error(), http.StatusForbidden)
			default:
				http.Error(responseWriter, err.Error(), http.StatusBadRequest)
			}
			return
		}

		responseWriter.WriteHeader(http.StatusNoContent)
	default:
		http.Error(responseWriter, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (server *signalingServer) joinRoom(roomIdentifier string) (joinRoomResponse, error) {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	room := server.rooms[roomIdentifier]
	if room == nil {
		room = &roomState{
			participantRoles: map[string]string{},
			lastPingAt:       time.Now(),
		}
		server.rooms[roomIdentifier] = room
		log.Printf("created signaling room %q", roomIdentifier)
	}

	if len(room.participantRoles) >= 2 {
		return joinRoomResponse{}, errRoomFull
	}

	role := "offerer"
	if len(room.participantRoles) == 1 {
		role = "answerer"
	}

	participantIdentifier, err := generateParticipantIdentifier()
	if err != nil {
		return joinRoomResponse{}, err
	}

	room.participantRoles[participantIdentifier] = role
	room.lastPingAt = time.Now()
	return joinRoomResponse{
		ParticipantIdentifier: participantIdentifier,
		Role:                  role,
	}, nil
}

func (server *signalingServer) leaveRoom(roomIdentifier string, participantIdentifier string) {
	roomIdentifier = strings.TrimSpace(roomIdentifier)
	participantIdentifier = strings.TrimSpace(participantIdentifier)
	if roomIdentifier == "" || participantIdentifier == "" {
		return
	}

	server.mutex.Lock()
	defer server.mutex.Unlock()

	room := server.rooms[roomIdentifier]
	if room == nil {
		return
	}

	delete(room.participantRoles, participantIdentifier)
	if len(room.participantRoles) == 0 {
		delete(server.rooms, roomIdentifier)
		return
	}

	room.offerPayload = nil
	room.answerPayload = nil
}

func (server *signalingServer) getSignal(roomIdentifier string, signalKind string) (json.RawMessage, bool) {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	room := server.rooms[roomIdentifier]
	if room == nil {
		return nil, false
	}

	if signalKind == "offer" && len(room.offerPayload) > 0 {
		return room.offerPayload, true
	}
	if signalKind == "answer" && len(room.answerPayload) > 0 {
		return room.answerPayload, true
	}

	return nil, false
}

func (server *signalingServer) setSignal(signalKind string, signalRequest signalingRequest) error {
	roomIdentifier, err := validateRoomIdentifier(signalRequest.RoomIdentifier)
	if err != nil {
		return err
	}
	participantIdentifier := strings.TrimSpace(signalRequest.ParticipantIdentifier)
	if participantIdentifier == "" {
		return errors.New("participant_identifier is required")
	}
	if len(signalRequest.SessionDescription) == 0 {
		return errors.New("session_description is required")
	}

	server.mutex.Lock()
	defer server.mutex.Unlock()

	room := server.rooms[roomIdentifier]
	if room == nil {
		return errRoomNotFound
	}

	role := room.participantRoles[participantIdentifier]
	if role == "" {
		return errParticipantNotFound
	}

	if signalKind == "offer" && role != "offerer" {
		return errRoleMismatch
	}
	if signalKind == "answer" && role != "answerer" {
		return errRoleMismatch
	}

	if signalKind == "offer" {
		room.offerPayload = append(json.RawMessage(nil), signalRequest.SessionDescription...)
		room.answerPayload = nil
		room.lastPingAt = time.Now()
		return nil
	}

	room.answerPayload = append(json.RawMessage(nil), signalRequest.SessionDescription...)
	room.lastPingAt = time.Now()
	return nil
}

func writeJSON(responseWriter http.ResponseWriter, payload any) {
	responseWriter.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(responseWriter).Encode(payload); err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
	}
}

func generateParticipantIdentifier() (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}

func validateRoomIdentifier(roomIdentifier string) (string, error) {
	roomIdentifier = strings.TrimSpace(roomIdentifier)
	if roomIdentifier == "" {
		return "", errors.New("room_identifier is required")
	}
	if len([]rune(roomIdentifier)) > maxRoomIdentifierLength {
		return "", errRoomIdentifierTooLong
	}
	return roomIdentifier, nil
}

func (server *signalingServer) pingRoom(roomIdentifier string, now time.Time) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	room := server.rooms[roomIdentifier]
	if room == nil {
		return errRoomNotFound
	}

	room.lastPingAt = now
	return nil
}

func (server *signalingServer) cleanupExpiredRoomsLoop() {
	ticker := time.NewTicker(roomCleanupInterval)
	defer ticker.Stop()

	for now := range ticker.C {
		server.cleanupExpiredRooms(now)
	}
}

func (server *signalingServer) cleanupExpiredRooms(now time.Time) {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	for roomIdentifier, room := range server.rooms {
		if room.lastPingAt.IsZero() {
			room.lastPingAt = now
			continue
		}
		if now.Sub(room.lastPingAt) >= roomExpirationWindow {
			delete(server.rooms, roomIdentifier)
			log.Printf("expired signaling room %q", roomIdentifier)
		}
	}
}
