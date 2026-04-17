package proxy

import (
	"sync"
	"time"

	"github.com/scaxe/scaxe-go/pkg/proxy/net"
)

type ProxySession struct {
	state	int
	peLayer	interface{}
	jeLayer	*net.Conn

	mu		sync.RWMutex
	entityPos	map[int32][3]float32
	playerEntities	map[int32]bool

	lastTeleportTime	time.Time
}

func NewProxySession(pe interface{}) *ProxySession {
	return &ProxySession{
		state:		0,
		peLayer:	pe,
		entityPos:	make(map[int32][3]float32),
		playerEntities:	make(map[int32]bool),
	}
}

func (s *ProxySession) GetState() int {
	return s.state
}

func (s *ProxySession) SetState(state int) {
	s.state = state
}

func (s *ProxySession) GetPE() interface{} {
	return s.peLayer
}

func (s *ProxySession) GetJE() *net.Conn {
	return s.jeLayer
}

func (s *ProxySession) SetJE(je *net.Conn) {
	s.jeLayer = je
}

func (s *ProxySession) SetEntityPos(eid int32, x, y, z float32) {
	s.mu.Lock()
	s.entityPos[eid] = [3]float32{x, y, z}
	s.mu.Unlock()
}

func (s *ProxySession) GetEntityPos(eid int32) (float32, float32, float32, bool) {
	s.mu.RLock()
	pos, ok := s.entityPos[eid]
	s.mu.RUnlock()
	return pos[0], pos[1], pos[2], ok
}

func (s *ProxySession) RemoveEntity(eid int32) {
	s.mu.Lock()
	delete(s.entityPos, eid)
	delete(s.playerEntities, eid)
	s.mu.Unlock()
}

func (s *ProxySession) MarkEntityAsPlayer(eid int32) {
	s.mu.Lock()
	s.playerEntities[eid] = true
	s.mu.Unlock()
}

func (s *ProxySession) IsPlayerEntity(eid int32) bool {
	s.mu.RLock()
	isPlayer := s.playerEntities[eid]
	s.mu.RUnlock()
	return isPlayer
}

func (s *ProxySession) MarkTeleport() {
	s.lastTeleportTime = time.Now()
}

func (s *ProxySession) IsInTeleportCooldown() bool {
	return time.Since(s.lastTeleportTime) < 500*time.Millisecond
}
