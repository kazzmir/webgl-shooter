package main

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestJoinRoomAssignsRoles(t *testing.T) {
	server := &signalingServer{rooms: map[string]*roomState{}}

	first, err := server.joinRoom("alpha")
	if err != nil {
		t.Fatalf("join first participant: %v", err)
	}
	second, err := server.joinRoom("alpha")
	if err != nil {
		t.Fatalf("join second participant: %v", err)
	}

	if first.Role != "offerer" {
		t.Fatalf("first role = %q, want offerer", first.Role)
	}
	if second.Role != "answerer" {
		t.Fatalf("second role = %q, want answerer", second.Role)
	}
}

func TestJoinRoomRejectsThirdParticipant(t *testing.T) {
	server := &signalingServer{rooms: map[string]*roomState{}}

	if _, err := server.joinRoom("alpha"); err != nil {
		t.Fatalf("join first participant: %v", err)
	}
	if _, err := server.joinRoom("alpha"); err != nil {
		t.Fatalf("join second participant: %v", err)
	}
	if _, err := server.joinRoom("alpha"); !errors.Is(err, errRoomFull) {
		t.Fatalf("join third participant err = %v, want %v", err, errRoomFull)
	}
}

func TestLeaveRoomKeepsOtherParticipantAndClearsSignals(t *testing.T) {
	server := &signalingServer{
		rooms: map[string]*roomState{
			"alpha": {
				participantRoles: map[string]string{
					"offerer":  "offerer",
					"answerer": "answerer",
				},
				offerPayload:  json.RawMessage(`{"type":"offer"}`),
				answerPayload: json.RawMessage(`{"type":"answer"}`),
			},
		},
	}

	server.leaveRoom("alpha", "answerer")

	room := server.rooms["alpha"]
	if room == nil {
		t.Fatalf("room was removed while one participant remained")
	}
	if len(room.participantRoles) != 1 {
		t.Fatalf("participant count = %d, want 1", len(room.participantRoles))
	}
	if len(room.offerPayload) != 0 || len(room.answerPayload) != 0 {
		t.Fatalf("signals were not cleared after leave")
	}
}

func TestSetSignalRequiresMatchingRole(t *testing.T) {
	server := &signalingServer{
		rooms: map[string]*roomState{
			"alpha": {
				participantRoles: map[string]string{
					"offerer":  "offerer",
					"answerer": "answerer",
				},
			},
		},
	}

	err := server.setSignal("answer", signalingRequest{
		RoomIdentifier:        "alpha",
		ParticipantIdentifier: "offerer",
		SessionDescription:    json.RawMessage(`{"type":"answer"}`),
	})
	if !errors.Is(err, errRoleMismatch) {
		t.Fatalf("set answer err = %v, want %v", err, errRoleMismatch)
	}
}

func TestValidateRoomIdentifierRejectsLongValue(t *testing.T) {
	roomIdentifier := strings.Repeat("a", maxRoomIdentifierLength+1)

	_, err := validateRoomIdentifier(roomIdentifier)
	if !errors.Is(err, errRoomIdentifierTooLong) {
		t.Fatalf("validate room id err = %v, want %v", err, errRoomIdentifierTooLong)
	}
}

func TestPingRoomUpdatesLastPingTime(t *testing.T) {
	server := &signalingServer{
		rooms: map[string]*roomState{
			"alpha": {
				participantRoles: map[string]string{},
			},
		},
	}

	now := time.Now()
	if err := server.pingRoom("alpha", now); err != nil {
		t.Fatalf("ping room: %v", err)
	}

	if got := server.rooms["alpha"].lastPingAt; !got.Equal(now) {
		t.Fatalf("lastPingAt = %v, want %v", got, now)
	}
}

func TestCleanupExpiredRoomsRemovesStaleEntries(t *testing.T) {
	now := time.Now()
	server := &signalingServer{
		rooms: map[string]*roomState{
			"fresh": {
				participantRoles: map[string]string{},
				lastPingAt:       now.Add(-30 * time.Second),
			},
			"stale": {
				participantRoles: map[string]string{},
				lastPingAt:       now.Add(-61 * time.Second),
			},
		},
	}

	server.cleanupExpiredRooms(now)

	if _, ok := server.rooms["stale"]; ok {
		t.Fatalf("stale room was not removed")
	}
	if _, ok := server.rooms["fresh"]; !ok {
		t.Fatalf("fresh room was removed")
	}
}
