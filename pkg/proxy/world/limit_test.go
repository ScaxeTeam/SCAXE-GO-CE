package world

import (
	"testing"
	"github.com/scaxe/scaxe-go/pkg/protocol"
	"github.com/scaxe/scaxe-go/pkg/proxy/xlat"
)

type mockSession struct {
	xlat.Session
	mockSender	*mockPESender
}

func (s *mockSession) GetPE() interface{} {
	return s.mockSender
}

type mockPESender struct {
	lastPacket protocol.DataPacket
}

func (m *mockPESender) SendPacket(packet protocol.DataPacket) {
	m.lastPacket = packet
}

func TestMovePlayerLimit(t *testing.T) {
	translator := &MovePlayerTranslator{}

	sender := &mockPESender{}
	session := &mockSession{mockSender: sender}

	legalPkt := &protocol.MovePlayerPacket{Y: 100}
	err := translator.Translate(session, legalPkt)
	if err != nil {
		t.Fatalf("Unexpected translation error: %v", err)
	}

	if sender.lastPacket != nil {
		t.Errorf("Expected no correction for Y=100, got packet sent")
	}

	illegalPkt := &protocol.MovePlayerPacket{Y: 150}
	err = translator.Translate(session, illegalPkt)
	if err != nil {
		t.Fatalf("Unexpected translation error: %v", err)
	}

	if illegalPkt.Y != 127 {
		t.Errorf("Expected local packet to be mutated to Y=127, got %v", illegalPkt.Y)
	}

	if sender.lastPacket == nil {
		t.Fatalf("Expected correction packet to be sent back downwards to bypass client prediction, got nil")
	}

	correction, ok := sender.lastPacket.(*protocol.MovePlayerPacket)
	if !ok {
		t.Fatalf("Expected correction to be MovePlayerPacket")
	}

	if correction.Y != 127 {
		t.Errorf("Expected correction Y to be strictly clipped to 127, got %v", correction.Y)
	}
}
