package ent

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/scaxe/scaxe-go/pkg/logger"
	"github.com/scaxe/scaxe-go/pkg/protocol"
	"github.com/scaxe/scaxe-go/pkg/proxy/net"
	"github.com/scaxe/scaxe-go/pkg/proxy/xlat"
	"github.com/scaxe/scaxe-go/pkg/item"
)

func MapEntityID(jeID byte) byte {
	switch jeID {
	case 50:
		return 33
	case 51:
		return 34
	case 52:
		return 35
	case 54:
		return 32
	case 55:
		return 37
	case 56:
		return 41
	case 57:
		return 36
	case 58:
		return 38
	case 59:
		return 40
	case 60:
		return 39
	case 61:
		return 43
	case 62:
		return 42
	case 65:
		return 19
	case 66:
		return 45
	case 90:
		return 12
	case 91:
		return 13
	case 92:
		return 11
	case 93:
		return 10
	case 94:
		return 17
	case 95:
		return 14
	case 96:
		return 16
	case 97:
		return 21
	case 98:
		return 22
	case 99:
		return 20
	case 100:
		return 23
	case 120:
		return 15
	}
	return 12
}

type EntityTranslator struct{}

type PESender interface {
	SendPacket(packet protocol.DataPacket)
}

func (t *EntityTranslator) Translate(session xlat.Session, packetID int32, payload []byte) error {
	switch packetID {
	case 0x04:
		return handleEntityEquipment(session, payload)
	case 0x0C:
		return handleSpawnPlayer(session, payload)
	case 0x0F:
		return handleSpawnMob(session, payload)
	case 0x15:
		return handleRelativeMove(session, payload)
	case 0x16:
		return handleEntityLook(session, payload)
	case 0x17:
		return handleLookAndRelativeMove(session, payload)
	case 0x18:
		return handleEntityTeleport(session, payload)
	case 0x13:
		return handleDestroyEntities(session, payload)
	case 0x14:
		return nil
	}
	return nil
}

func readJESlot(buf *bytes.Reader) item.Item {
	var id int16
	binary.Read(buf, binary.BigEndian, &id)
	if id == -1 || id == 0 {
		return item.Item{ID: 0}
	}
	count, _ := buf.ReadByte()
	var damage int16
	binary.Read(buf, binary.BigEndian, &damage)

	var nbtLen int16
	binary.Read(buf, binary.BigEndian, &nbtLen)
	if nbtLen > 0 && nbtLen != -1 {
		b := make([]byte, nbtLen)
		buf.Read(b)
	}

	return item.Item{ID: int(id), Count: int(count), Meta: int(damage)}
}

func handleEntityEquipment(session xlat.Session, payload []byte) error {
	buf := bytes.NewReader(payload)

	var eid int32
	binary.Read(buf, binary.BigEndian, &eid)

	var slot int16
	binary.Read(buf, binary.BigEndian, &slot)

	itm := readJESlot(buf)
	logger.ProxyDebug("<<< S04 EntityEquipment", "eid", eid, "slot", slot, "item", itm.ID)

	if slot == 0 {
		pkt := protocol.NewMobEquipmentPacket()
		pkt.EntityID = int64(eid)
		pkt.ItemID = int16(itm.ID)
		pkt.ItemCount = int8(itm.Count)
		pkt.ItemMeta = uint16(itm.Meta)
		pkt.Slot = 0
		pkt.SelectedSlot = 0
		if sender, ok := session.GetPE().(PESender); ok {
			sender.SendPacket(pkt)
		}
	} else if slot >= 1 && slot <= 4 {
		pkt := protocol.NewMobArmorEquipmentPacket()
		pkt.EntityID = int64(eid)
		aItm := protocol.ArmorItem{ID: int16(itm.ID), Count: int8(itm.Count), Meta: uint16(itm.Meta)}
		switch slot {
		case 1:
			pkt.Slots[3] = aItm
		case 2:
			pkt.Slots[2] = aItm
		case 3:
			pkt.Slots[1] = aItm
		case 4:
			pkt.Slots[0] = aItm
		}
		if sender, ok := session.GetPE().(PESender); ok {
			sender.SendPacket(pkt)
		}
	}
	return nil
}

func readJEString(buf *bytes.Reader) string {
	length, _ := net.ReadVarInt(buf)
	if length <= 0 || length > 32767 {
		return ""
	}
	b := make([]byte, length)
	buf.Read(b)
	return string(b)
}

func handleSpawnPlayer(session xlat.Session, payload []byte) error {
	buf := bytes.NewReader(payload)

	eid, _ := net.ReadVarInt(buf)

	uuidStr := readJEString(buf)

	name := readJEString(buf)

	dataCount, _ := net.ReadVarInt(buf)
	for i := int32(0); i < dataCount; i++ {
		readJEString(buf)
		readJEString(buf)
		readJEString(buf)
	}

	var x, y, z int32
	binary.Read(buf, binary.BigEndian, &x)
	binary.Read(buf, binary.BigEndian, &y)
	binary.Read(buf, binary.BigEndian, &z)

	yawByte, _ := buf.ReadByte()
	pitchByte, _ := buf.ReadByte()

	fx := float32(x) / 32.0
	fy := float32(y) / 32.0
	fz := float32(z) / 32.0

	session.MarkEntityAsPlayer(eid)

	fy += 1.62

	session.SetEntityPos(eid, fx, fy, fz)

	logger.ProxyDebug("<<< S0C SpawnPlayer",
		"eid", eid,
		"name", name,
		"uuid", uuidStr,
		"x", fmt.Sprintf("%.2f", fx),
		"y", fmt.Sprintf("%.2f", fy),
		"z", fmt.Sprintf("%.2f", fz))

	pkt := protocol.NewAddPlayerPacket()
	pkt.UUID = uuidStr
	pkt.Username = name
	pkt.EntityID = int64(eid)
	pkt.X = fx
	pkt.Y = fy
	pkt.Z = fz
	pkt.Yaw = float32(yawByte) * (360.0 / 256.0)
	pkt.Pitch = float32(pitchByte) * (360.0 / 256.0)

	if sender, ok := session.GetPE().(PESender); ok {
		sender.SendPacket(pkt)
	}

	return nil
}

func handleSpawnMob(session xlat.Session, payload []byte) error {
	buf := bytes.NewReader(payload)

	eid, _ := net.ReadVarInt(buf)
	entType, _ := buf.ReadByte()

	var x, y, z int32
	binary.Read(buf, binary.BigEndian, &x)
	binary.Read(buf, binary.BigEndian, &y)
	binary.Read(buf, binary.BigEndian, &z)

	pitch, _ := buf.ReadByte()
	yaw, _ := buf.ReadByte()

	fx := float32(x) / 32.0
	fy := float32(y) / 32.0
	fz := float32(z) / 32.0

	session.SetEntityPos(eid, fx, fy, fz)

	logger.ProxyDebug("<<< S0F SpawnMob",
		"eid", eid,
		"jeType", entType,
		"peType", MapEntityID(entType),
		"x", fmt.Sprintf("%.2f", fx),
		"y", fmt.Sprintf("%.2f", fy),
		"z", fmt.Sprintf("%.2f", fz))

	pkt := protocol.NewAddEntityPacket()
	pkt.EntityID = int64(eid)
	pkt.Type = int32(MapEntityID(entType))
	pkt.X = fx
	pkt.Y = fy
	pkt.Z = fz
	pkt.Pitch = float32(pitch) * (360.0 / 256.0)
	pkt.Yaw = float32(yaw) * (360.0 / 256.0)

	if sender, ok := session.GetPE().(PESender); ok {
		sender.SendPacket(pkt)
	}

	return nil
}

func readEntityInt(buf *bytes.Reader) int32 {
	var eid int32
	binary.Read(buf, binary.BigEndian, &eid)
	return eid
}

func sendMoveEntity(session xlat.Session, eid int32, x, y, z, yaw, pitch float32) {
	if session.IsPlayerEntity(eid) {
		pkt := protocol.NewMovePlayerPacket()
		pkt.EntityID = int64(eid)
		pkt.X = x
		pkt.Y = y
		pkt.Z = z
		pkt.Yaw = yaw
		pkt.BodyYaw = yaw
		pkt.Pitch = pitch
		pkt.Mode = protocol.MovePlayerModeNormal
		if sender, ok := session.GetPE().(PESender); ok {
			sender.SendPacket(pkt)
		}
		return
	}

	pkt := protocol.NewMoveEntityPacket()
	pkt.Entities = []protocol.MoveEntityEntry{{
		EntityID:	int64(eid),
		X:		x, Y: y, Z: z,
		Yaw:	yaw, HeadYaw: yaw, Pitch: pitch,
	}}
	if sender, ok := session.GetPE().(PESender); ok {
		sender.SendPacket(pkt)
	}
}

func handleRelativeMove(session xlat.Session, payload []byte) error {
	buf := bytes.NewReader(payload)
	eid := readEntityInt(buf)
	dx, _ := buf.ReadByte()
	dy, _ := buf.ReadByte()
	dz, _ := buf.ReadByte()

	ox, oy, oz, ok := session.GetEntityPos(eid)
	if !ok {
		return nil
	}

	nx := ox + float32(int8(dx))/32.0
	ny := oy + float32(int8(dy))/32.0
	nz := oz + float32(int8(dz))/32.0
	session.SetEntityPos(eid, nx, ny, nz)

	sendMoveEntity(session, eid, nx, ny, nz, 0, 0)
	return nil
}

func handleEntityLook(session xlat.Session, payload []byte) error {
	buf := bytes.NewReader(payload)
	eid := readEntityInt(buf)
	yawByte, _ := buf.ReadByte()
	pitchByte, _ := buf.ReadByte()

	ox, oy, oz, ok := session.GetEntityPos(eid)
	if !ok {
		return nil
	}

	yaw := float32(yawByte) * (360.0 / 256.0)
	pitch := float32(pitchByte) * (360.0 / 256.0)
	sendMoveEntity(session, eid, ox, oy, oz, yaw, pitch)
	return nil
}

func handleLookAndRelativeMove(session xlat.Session, payload []byte) error {
	buf := bytes.NewReader(payload)
	eid := readEntityInt(buf)
	dx, _ := buf.ReadByte()
	dy, _ := buf.ReadByte()
	dz, _ := buf.ReadByte()
	yawByte, _ := buf.ReadByte()
	pitchByte, _ := buf.ReadByte()

	ox, oy, oz, ok := session.GetEntityPos(eid)
	if !ok {
		return nil
	}

	nx := ox + float32(int8(dx))/32.0
	ny := oy + float32(int8(dy))/32.0
	nz := oz + float32(int8(dz))/32.0
	session.SetEntityPos(eid, nx, ny, nz)

	yaw := float32(yawByte) * (360.0 / 256.0)
	pitch := float32(pitchByte) * (360.0 / 256.0)
	sendMoveEntity(session, eid, nx, ny, nz, yaw, pitch)
	return nil
}

func handleEntityTeleport(session xlat.Session, payload []byte) error {
	buf := bytes.NewReader(payload)
	eid := readEntityInt(buf)

	var x, y, z int32
	binary.Read(buf, binary.BigEndian, &x)
	binary.Read(buf, binary.BigEndian, &y)
	binary.Read(buf, binary.BigEndian, &z)

	yawByte, _ := buf.ReadByte()
	pitchByte, _ := buf.ReadByte()

	fx := float32(x) / 32.0
	fy := float32(y) / 32.0
	fz := float32(z) / 32.0

	if session.IsPlayerEntity(eid) {
		fy += 1.62
	}

	session.SetEntityPos(eid, fx, fy, fz)

	yaw := float32(yawByte) * (360.0 / 256.0)
	pitch := float32(pitchByte) * (360.0 / 256.0)
	sendMoveEntity(session, eid, fx, fy, fz, yaw, pitch)
	return nil
}

func handleDestroyEntities(session xlat.Session, payload []byte) error {
	buf := bytes.NewReader(payload)
	count, _ := buf.ReadByte()

	logger.ProxyDebug("<<< S13 DestroyEntities", "count", count)

	if sender, ok := session.GetPE().(PESender); ok {
		for i := 0; i < int(count); i++ {
			var eid int32
			binary.Read(buf, binary.BigEndian, &eid)

			session.RemoveEntity(eid)

			pkt := protocol.NewRemoveEntityPacket()
			pkt.EntityID = int64(eid)
			sender.SendPacket(pkt)
		}
	}
	return nil
}

type TimeTranslator struct{}

func (t *TimeTranslator) Translate(session xlat.Session, packetID int32, payload []byte) error {
	buf := bytes.NewReader(payload)

	var worldAge int64
	binary.Read(buf, binary.BigEndian, &worldAge)
	var timeOfDay int64
	binary.Read(buf, binary.BigEndian, &timeOfDay)

	pkt := protocol.NewSetTimePacket()
	pkt.Time = int32(timeOfDay % 24000)
	pkt.Started = true

	if sender, ok := session.GetPE().(PESender); ok {
		sender.SendPacket(pkt)
	}

	return nil
}
