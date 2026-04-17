package world

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/scaxe/scaxe-go/pkg/logger"
	"github.com/scaxe/scaxe-go/pkg/protocol"
	"github.com/scaxe/scaxe-go/pkg/proxy/xlat"
)

type MovePlayerTranslator struct{}

func (t *MovePlayerTranslator) Translate(session xlat.Session, pePkt protocol.DataPacket) error {
	movePkt, ok := pePkt.(*protocol.MovePlayerPacket)
	if !ok {
		return nil
	}

	logger.ProxyDebug("<<< PE MovePlayer received",
		"eid", movePkt.EntityID,
		"x", fmt.Sprintf("%.2f", movePkt.X),
		"eyeY", fmt.Sprintf("%.2f", movePkt.Y),
		"z", fmt.Sprintf("%.2f", movePkt.Z),
		"yaw", fmt.Sprintf("%.1f", movePkt.Yaw),
		"pitch", fmt.Sprintf("%.1f", movePkt.Pitch),
		"mode", movePkt.Mode,
		"onGround", movePkt.OnGround)

	if movePkt.Y > 127 {
		movePkt.Y = 127
		logger.ProxyDebug("--- PE MovePlayer Y clamped to 127")

		if sender, ok := session.GetPE().(PESender); ok {
			correction := &protocol.MovePlayerPacket{
				EntityID:	movePkt.EntityID,
				X:		movePkt.X,
				Y:		127,
				Z:		movePkt.Z,
				Yaw:		movePkt.Yaw,
				Pitch:		movePkt.Pitch,
				BodyYaw:	movePkt.BodyYaw,
				Mode:		movePkt.Mode,
				OnGround:	movePkt.OnGround,
			}
			sender.SendPacket(correction)
		}
	}

	jeConn := session.GetJE()
	if jeConn == nil {
		logger.ProxyDebug("--- PE MovePlayer: no JE conn, dropping")
		return nil
	}

	if session.IsInTeleportCooldown() {
		logger.ProxyDebug("--- PE MovePlayer SUPPRESSED (teleport cooldown active)")
		return nil
	}

	feetY := float64(movePkt.Y) - 1.62
	headY := float64(movePkt.Y)

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, float64(movePkt.X))
	binary.Write(buf, binary.BigEndian, feetY)
	binary.Write(buf, binary.BigEndian, headY)
	binary.Write(buf, binary.BigEndian, float64(movePkt.Z))
	binary.Write(buf, binary.BigEndian, movePkt.Yaw)
	binary.Write(buf, binary.BigEndian, movePkt.Pitch)

	onGround := byte(0)
	if movePkt.OnGround {
		onGround = 1
	}
	buf.WriteByte(onGround)

	err := jeConn.WritePacket(0x06, buf.Bytes())
	logger.ProxyDebug(">>> C06 movement forwarded to JE",
		"feetY", fmt.Sprintf("%.2f", feetY),
		"headY", fmt.Sprintf("%.2f", headY),
		"err", err)

	return err
}
