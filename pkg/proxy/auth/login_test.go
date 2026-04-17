package auth

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"

	"github.com/scaxe/scaxe-go/pkg/protocol"
	pnet "github.com/scaxe/scaxe-go/pkg/proxy/net"
	"github.com/scaxe/scaxe-go/pkg/proxy/xlat"
)

type mockAuthSession struct {
	xlat.Session
	mockSender	*mockPESender
	state		int
	jeConn		*pnet.Conn
}

func (s *mockAuthSession) SetState(st int)		{ s.state = st }
func (s *mockAuthSession) GetState() int		{ return s.state }
func (s *mockAuthSession) GetPE() interface{}		{ return s.mockSender }
func (s *mockAuthSession) SetJE(conn *pnet.Conn)	{ s.jeConn = conn }
func (s *mockAuthSession) GetJE() *pnet.Conn		{ return s.jeConn }

type mockPESender struct {
	received []protocol.DataPacket
}

func (m *mockPESender) SendPacket(packet protocol.DataPacket) {
	m.received = append(m.received, packet)
}

func TestLoginHandshake(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
	}()

	translator := &PELoginTranslator{
		JEAddress:	"127.0.0.1",
		JEPort:		uint16(port),
	}

	session := &mockAuthSession{
		mockSender: &mockPESender{},
	}

	loginPkt := &protocol.LoginPacket{}
	loginPkt.Username = "TestPlayer"

	err = translator.Translate(session, loginPkt)
	if err != nil {
		t.Fatalf("Translate failed: %v", err)
	}

	if session.GetState() != 1 {
		t.Errorf("Expected state to be 1, got %d", session.GetState())
	}
}

func TestJELoginSuccess(t *testing.T) {
	session := &mockAuthSession{
		mockSender:	&mockPESender{},
		state:		1,
	}

	translator := &JELoginTranslator{}

	payload := new(bytes.Buffer)
	pnet.WriteString(payload, "dummy-uuid")
	pnet.WriteString(payload, "TestPlayer")

	err := translator.Translate(session, 0x02, payload.Bytes())
	if err != nil {
		t.Fatalf("S02 handling failed: %v", err)
	}

	if session.GetState() != 2 {
		t.Errorf("Expected state to advance to 2, got %d", session.GetState())
	}

	if len(session.mockSender.received) == 0 {
		t.Fatalf("Expected PlayStatusPacket to be sent to PE")
	}
}

func TestJEJoinGame(t *testing.T) {
	session := &mockAuthSession{
		mockSender:	&mockPESender{},
		state:		2,
	}

	translator := &JELoginTranslator{}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, int32(1337))
	buf.WriteByte(1)
	buf.WriteByte(0)
	buf.WriteByte(1)
	buf.WriteByte(64)
	pnet.WriteString(buf, "default")

	err := translator.Translate(session, 0x01, buf.Bytes())
	if err != nil {
		t.Fatalf("Join handling failed: %v", err)
	}

	if session.GetState() != 3 {
		t.Errorf("Expected state to advance to 3, got %d", session.GetState())
	}

	if len(session.mockSender.received) == 0 {
		t.Fatalf("Expected StartGamePacket to be sent to PE")
	}

	startGamePkt, ok := session.mockSender.received[0].(*protocol.StartGamePacket)
	if !ok || startGamePkt.EntityID != 1337 || startGamePkt.Dimension != 0 {
		t.Errorf("StartGamePacket values incorrectly translated")
	}
}
