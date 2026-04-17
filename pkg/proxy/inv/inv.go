package inv

import (
	"bytes"
	"encoding/binary"

	"github.com/scaxe/scaxe-go/pkg/protocol"
	"github.com/scaxe/scaxe-go/pkg/proxy/net"
	"github.com/scaxe/scaxe-go/pkg/proxy/xlat"
)

func MapItemID(id int16) int16 {
	return id
}

type InvTranslator struct{}

type PESender interface {
	SendPacket(packet protocol.DataPacket)
}

func (t *InvTranslator) Translate(session xlat.Session, packetID int32, payload []byte) error {
	switch packetID {
	case 0x2D:
		return handleOpenWindow(session, payload)
	case 0x2E:
		return handleCloseWindow(session, payload)
	case 0x2F:
		return handleSetSlot(session, payload)
	case 0x30:
		return handleWindowItems(session, payload)
	case 0x32:
		return handleConfirmTx(session, payload)
	}
	return nil
}

func handleOpenWindow(session xlat.Session, payload []byte) error {
	buf := bytes.NewReader(payload)
	windowID, _ := buf.ReadByte()
	invType, _ := buf.ReadByte()

	net.ReadString(buf)
	buf.ReadByte()

	var peType byte
	switch invType {
	case 0:
		peType = 1
	case 1:
		peType = 2
	case 2:
		peType = 3
	default:
		peType = 0
	}

	pkt := protocol.NewContainerOpenPacket()
	pkt.WindowID = windowID
	pkt.Type = peType

	if sender, ok := session.GetPE().(PESender); ok {
		sender.SendPacket(pkt)
	}
	return nil
}

func handleCloseWindow(session xlat.Session, payload []byte) error {
	buf := bytes.NewReader(payload)
	windowID, _ := buf.ReadByte()

	pkt := protocol.NewContainerClosePacket()
	pkt.WindowID = windowID
	if sender, ok := session.GetPE().(PESender); ok {
		sender.SendPacket(pkt)
	}
	return nil
}

func handleSetSlot(session xlat.Session, payload []byte) error {
	buf := bytes.NewReader(payload)
	windowID, _ := buf.ReadByte()

	var slot int16
	binary.Read(buf, binary.BigEndian, &slot)

	var id int16
	binary.Read(buf, binary.BigEndian, &id)

	if id == -1 {
		return sendSetSlotToPE(session, windowID, slot, 0, 0, 0)
	}

	count, _ := buf.ReadByte()
	var damage int16
	binary.Read(buf, binary.BigEndian, &damage)

	return sendSetSlotToPE(session, windowID, slot, MapItemID(id), count, damage)
}

func handleWindowItems(session xlat.Session, payload []byte) error {

	return nil
}

func handleConfirmTx(session xlat.Session, payload []byte) error {

	return nil
}

func sendSetSlotToPE(session xlat.Session, winID byte, slot, id int16, count byte, damage int16) error {

	return nil
}
