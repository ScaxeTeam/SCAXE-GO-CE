package world

import (
	"bytes"
	"encoding/binary"

	"github.com/scaxe/scaxe-go/pkg/logger"
	"github.com/scaxe/scaxe-go/pkg/protocol"
	"github.com/scaxe/scaxe-go/pkg/proxy/xlat"
)

type ActionTranslator struct{}

func (t *ActionTranslator) Translate(session xlat.Session, pePkt protocol.DataPacket) error {
	switch pkt := pePkt.(type) {
	case *protocol.PlayerActionPacket:
		return handlePlayerAction(session, pkt)
	case *protocol.UseItemPacket:
		return handleUseItem(session, pkt)
	}
	return nil
}

func handlePlayerAction(session xlat.Session, pkt *protocol.PlayerActionPacket) error {
	jeConn := session.GetJE()
	if jeConn == nil {
		return nil
	}

	status := byte(pkt.Action)
	if status > 2 {
		return nil
	}

	buf := new(bytes.Buffer)
	buf.WriteByte(status)
	binary.Write(buf, binary.BigEndian, int32(pkt.X))
	buf.WriteByte(byte(pkt.Y))
	binary.Write(buf, binary.BigEndian, int32(pkt.Z))
	buf.WriteByte(byte(pkt.Face))

	logger.ProxyDebug(">>> C07 BlockDig", "status", status, "x", pkt.X, "y", pkt.Y, "z", pkt.Z)
	return jeConn.WritePacket(0x07, buf.Bytes())
}

func handleUseItem(session xlat.Session, pkt *protocol.UseItemPacket) error {
	jeConn := session.GetJE()
	if jeConn == nil {
		return nil
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, int32(pkt.X))
	buf.WriteByte(byte(pkt.Y))
	binary.Write(buf, binary.BigEndian, int32(pkt.Z))
	buf.WriteByte(byte(pkt.Face))

	if pkt.Item.ID <= 0 {
		binary.Write(buf, binary.BigEndian, int16(-1))
	} else {

		binary.Write(buf, binary.BigEndian, int16(pkt.Item.ID))
		buf.WriteByte(byte(pkt.Item.Count))
		binary.Write(buf, binary.BigEndian, int16(pkt.Item.Meta))
		binary.Write(buf, binary.BigEndian, int16(-1))
	}

	buf.WriteByte(byte(pkt.FX * 16.0))
	buf.WriteByte(byte(pkt.FY * 16.0))
	buf.WriteByte(byte(pkt.FZ * 16.0))

	logger.ProxyDebug(">>> C08 BlockPlace", "x", pkt.X, "y", pkt.Y, "z", pkt.Z, "face", pkt.Face)
	return jeConn.WritePacket(0x08, buf.Bytes())
}
