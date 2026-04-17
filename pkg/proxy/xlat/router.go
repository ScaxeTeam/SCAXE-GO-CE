package xlat

import (
	"github.com/scaxe/scaxe-go/pkg/protocol"
	"github.com/scaxe/scaxe-go/pkg/proxy/net"
)

type Session interface {
	GetPE() interface{}
	GetJE() *net.Conn
	SetJE(jeConn *net.Conn)
	SetState(state int)
	GetState() int

	SetEntityPos(eid int32, x, y, z float32)
	GetEntityPos(eid int32) (float32, float32, float32, bool)
	RemoveEntity(eid int32)

	MarkEntityAsPlayer(eid int32)
	IsPlayerEntity(eid int32) bool

	MarkTeleport()
	IsInTeleportCooldown() bool
}

type PeToJeTranslator interface {
	Translate(session Session, pePkt protocol.DataPacket) error
}

type JeToPeTranslator interface {
	Translate(session Session, packetID int32, payload []byte) error
}

type Router struct {
	jeTranslators	map[int32]JeToPeTranslator
	peTranslators	map[byte]PeToJeTranslator
}

func NewRouter() *Router {
	return &Router{
		jeTranslators:	make(map[int32]JeToPeTranslator),
		peTranslators:	make(map[byte]PeToJeTranslator),
	}
}

func (r *Router) RegisterJE(packetID int32, translator JeToPeTranslator) {
	r.jeTranslators[packetID] = translator
}

func (r *Router) RegisterPE(packetID byte, translator PeToJeTranslator) {
	r.peTranslators[packetID] = translator
}

func (r *Router) HandleJE(session Session, packetID int32, payload []byte) error {
	t, ok := r.jeTranslators[packetID]
	if !ok {
		return nil
	}
	return t.Translate(session, packetID, payload)
}

func (r *Router) HandlePE(session Session, pePkt protocol.DataPacket) error {
	t, ok := r.peTranslators[pePkt.ID()]
	if !ok {
		return nil
	}
	return t.Translate(session, pePkt)
}
