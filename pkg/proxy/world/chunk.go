package world

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/scaxe/scaxe-go/pkg/logger"
	"github.com/scaxe/scaxe-go/pkg/protocol"
	"github.com/scaxe/scaxe-go/pkg/proxy/net"
	"github.com/scaxe/scaxe-go/pkg/proxy/xlat"
)

type ChunkTranslator struct{}

type PESender interface {
	SendPacket(packet protocol.DataPacket)
}

func (t *ChunkTranslator) Translate(session xlat.Session, packetID int32, payload []byte) error {
	switch packetID {
	case 0x21:
		return handleS21(session, payload)
	case 0x23:
		return handleS23(session, payload)
	case 0x24:
		return handleS24(session, payload)
	case 0x26:
		return handleS26(session, payload)
	case 0x08:
		return handleS08(session, payload)
	}
	return nil
}

type peChunkData struct {
	blocks		[32768]byte
	data		[16384]byte
	skyLight	[16384]byte
	blockLight	[16384]byte
	heightMap	[256]byte
	biomeColor	[256]uint32
}

func newPEChunkData() *peChunkData {
	c := &peChunkData{}

	for i := range c.skyLight {
		c.skyLight[i] = 0xFF
	}
	return c
}

func peIndex(x, y, z int) int {
	return (x << 11) | (z << 7) | y
}

func peNibbleSet(arr *[16384]byte, x, y, z int, val byte) {
	idx := peIndex(x, y, z)
	byteIdx := idx >> 1
	if idx&1 == 0 {
		arr[byteIdx] = (arr[byteIdx] & 0xF0) | (val & 0x0F)
	} else {
		arr[byteIdx] = (arr[byteIdx] & 0x0F) | ((val & 0x0F) << 4)
	}
}

func (c *peChunkData) setBlock(x, y, z int, id byte) {
	if y >= 128 {
		return
	}
	c.blocks[peIndex(x, y, z)] = id
}

func (c *peChunkData) setMeta(x, y, z int, meta byte) {
	if y >= 128 {
		return
	}
	peNibbleSet(&c.data, x, y, z, meta)
}

func (c *peChunkData) setSkyLight(x, y, z int, val byte) {
	if y >= 128 {
		return
	}
	peNibbleSet(&c.skyLight, x, y, z, val)
}

func (c *peChunkData) setBlockLight(x, y, z int, val byte) {
	if y >= 128 {
		return
	}
	peNibbleSet(&c.blockLight, x, y, z, val)
}

func (c *peChunkData) toBytes() []byte {
	buf := new(bytes.Buffer)
	buf.Grow(83204)
	buf.Write(c.blocks[:])
	buf.Write(c.data[:])
	buf.Write(c.skyLight[:])
	buf.Write(c.blockLight[:])
	buf.Write(c.heightMap[:])
	for _, color := range c.biomeColor {
		binary.Write(buf, binary.BigEndian, color)
	}
	binary.Write(buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func jeNibbleGet(nibbles []byte, jeIdx int) byte {
	byteIdx := jeIdx >> 1
	if jeIdx&1 == 0 {
		return nibbles[byteIdx] & 0x0F
	}
	return nibbles[byteIdx] >> 4
}

func parseSections(dataBuf *bytes.Reader, pMask uint16, skyLightPresent bool, pe *peChunkData) {

	type sectionRaw struct {
		blocks		[]byte
		meta		[]byte
		blockLight	[]byte
		skyLight	[]byte
	}

	sections := make([]*sectionRaw, 16)

	for i := 0; i < 16; i++ {
		if (pMask & (1 << i)) != 0 {
			raw := make([]byte, 4096)
			io.ReadFull(dataBuf, raw)
			if sections[i] == nil {
				sections[i] = &sectionRaw{}
			}
			sections[i].blocks = raw
		}
	}

	for i := 0; i < 16; i++ {
		if (pMask & (1 << i)) != 0 {
			raw := make([]byte, 2048)
			io.ReadFull(dataBuf, raw)
			if sections[i] == nil {
				sections[i] = &sectionRaw{}
			}
			sections[i].meta = raw
		}
	}

	for i := 0; i < 16; i++ {
		if (pMask & (1 << i)) != 0 {
			raw := make([]byte, 2048)
			io.ReadFull(dataBuf, raw)
			if sections[i] == nil {
				sections[i] = &sectionRaw{}
			}
			sections[i].blockLight = raw
		}
	}

	if skyLightPresent {
		for i := 0; i < 16; i++ {
			if (pMask & (1 << i)) != 0 {
				raw := make([]byte, 2048)
				io.ReadFull(dataBuf, raw)
				if sections[i] == nil {
					sections[i] = &sectionRaw{}
				}
				sections[i].skyLight = raw
			}
		}
	}

	for i := 0; i < 16; i++ {
		if sections[i] == nil || i >= 8 {
			continue
		}
		sec := sections[i]
		baseY := i * 16

		for y := 0; y < 16; y++ {
			globalY := baseY + y
			for z := 0; z < 16; z++ {
				for x := 0; x < 16; x++ {
					jeIdx := (y << 8) | (z << 4) | x

					if sec.blocks != nil {
						pe.setBlock(x, globalY, z, MapBlockID(sec.blocks[jeIdx]))
					}

					if sec.meta != nil {
						pe.setMeta(x, globalY, z, jeNibbleGet(sec.meta, jeIdx))
					}

					if sec.blockLight != nil {
						pe.setBlockLight(x, globalY, z, jeNibbleGet(sec.blockLight, jeIdx))
					}

					if sec.skyLight != nil {
						pe.setSkyLight(x, globalY, z, jeNibbleGet(sec.skyLight, jeIdx))
					}
				}
			}
		}
	}

	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			for y := 127; y >= 0; y-- {
				if pe.blocks[peIndex(x, y, z)] != 0 {
					pe.heightMap[(z<<4)|x] = byte(y + 1)
					break
				}
			}
		}
	}
}

func sendPEChunk(session xlat.Session, chunkX, chunkZ int32, pe *peChunkData) {
	pkt := protocol.NewFullChunkDataPacket()
	pkt.ChunkX = chunkX
	pkt.ChunkZ = chunkZ
	pkt.Data = pe.toBytes()

	if sender, ok := session.GetPE().(PESender); ok {
		sender.SendPacket(pkt)
	}
}

func handleS21(session xlat.Session, payload []byte) error {
	buf := bytes.NewReader(payload)
	var chunkX, chunkZ int32
	binary.Read(buf, binary.BigEndian, &chunkX)
	binary.Read(buf, binary.BigEndian, &chunkZ)

	groundUp, _ := buf.ReadByte()

	var primaryBitMask, addBitMask uint16
	binary.Read(buf, binary.BigEndian, &primaryBitMask)
	binary.Read(buf, binary.BigEndian, &addBitMask)
	_ = addBitMask

	var compressedSize int32
	binary.Read(buf, binary.BigEndian, &compressedSize)

	compressedData := make([]byte, compressedSize)
	io.ReadFull(buf, compressedData)

	r, err := zlib.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return err
	}
	defer r.Close()

	uncompressed, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	dataBuf := bytes.NewReader(uncompressed)
	pe := newPEChunkData()

	parseSections(dataBuf, primaryBitMask, true, pe)

	if groundUp == 1 {
		biomeData := make([]byte, 256)
		io.ReadFull(dataBuf, biomeData)
		for z := 0; z < 16; z++ {
			for x := 0; x < 16; x++ {
				pe.biomeColor[(z<<4)|x] = biomeGrassColor(biomeData[z*16+x])
			}
		}
	}

	sendPEChunk(session, chunkX, chunkZ, pe)
	return nil
}

func handleS08(session xlat.Session, payload []byte) error {
	buf := bytes.NewReader(payload)

	var x, eyeY, z float64
	binary.Read(buf, binary.BigEndian, &x)
	binary.Read(buf, binary.BigEndian, &eyeY)
	binary.Read(buf, binary.BigEndian, &z)

	var yaw, pitch float32
	binary.Read(buf, binary.BigEndian, &yaw)
	binary.Read(buf, binary.BigEndian, &pitch)

	onGround, _ := buf.ReadByte()

	logger.ProxyDebug("<<< S08 PlayerPosLook received",
		"x", fmt.Sprintf("%.2f", x),
		"eyeY", fmt.Sprintf("%.2f", eyeY),
		"feetY", fmt.Sprintf("%.2f", eyeY-1.62),
		"z", fmt.Sprintf("%.2f", z),
		"yaw", fmt.Sprintf("%.1f", yaw),
		"pitch", fmt.Sprintf("%.1f", pitch),
		"onGround", onGround,
		"state", session.GetState(),
		"payloadLen", len(payload))

	sender, ok := session.GetPE().(PESender)
	if !ok {
		logger.ProxyDebug("!!! S08 handler: no PESender available")
		return nil
	}

	feetY := eyeY - 1.62

	if session.GetState() == 3 {
		session.SetState(4)

		adv := protocol.NewAdventureSettingsPacket()
		adv.Flags = 0
		adv.UserPermission = 2
		adv.GlobalPermission = 2
		sender.SendPacket(adv)

		respawnPk := protocol.NewRespawnPacket()
		respawnPk.X = float32(x)
		respawnPk.Y = float32(feetY)
		respawnPk.Z = float32(z)
		sender.SendPacket(respawnPk)

		status := protocol.NewPlayStatusPacket()
		status.Status = protocol.PlayStatusPlayerSpawn
		sender.SendPacket(status)
	}

	movePk := protocol.NewMovePlayerPacket()
	if player, ok := session.GetPE().(interface{ GetID() int64 }); ok {
		movePk.EntityID = player.GetID()
	}
	movePk.X = float32(x)
	movePk.Y = float32(eyeY)
	movePk.Z = float32(z)
	movePk.Yaw = yaw
	movePk.BodyYaw = yaw
	movePk.Pitch = pitch
	movePk.Mode = 1
	movePk.OnGround = onGround == 1
	sender.SendPacket(movePk)

	jeConn := session.GetJE()
	if jeConn != nil {
		confirmBuf := new(bytes.Buffer)
		binary.Write(confirmBuf, binary.BigEndian, x)
		binary.Write(confirmBuf, binary.BigEndian, feetY)
		binary.Write(confirmBuf, binary.BigEndian, eyeY)
		binary.Write(confirmBuf, binary.BigEndian, z)
		binary.Write(confirmBuf, binary.BigEndian, yaw)
		binary.Write(confirmBuf, binary.BigEndian, pitch)
		confirmBuf.WriteByte(onGround)
		err := jeConn.WritePacket(0x06, confirmBuf.Bytes())
		logger.ProxyDebug(">>> C06 teleport confirmation sent",
			"x", fmt.Sprintf("%.2f", x),
			"feetY", fmt.Sprintf("%.2f", feetY),
			"headY", fmt.Sprintf("%.2f", eyeY),
			"z", fmt.Sprintf("%.2f", z),
			"err", err)
	} else {
		logger.ProxyDebug("!!! S08 handler: no JE conn for C06 confirm")
	}

	session.MarkTeleport()
	logger.ProxyDebug("--- Teleport cooldown activated (500ms)")

	return nil
}

func handleS26(session xlat.Session, payload []byte) error {
	buf := bytes.NewReader(payload)
	var count int16
	binary.Read(buf, binary.BigEndian, &count)

	var dataLen int32
	binary.Read(buf, binary.BigEndian, &dataLen)

	skyLightSent, _ := buf.ReadByte()
	logger.ProxyDebug("<<< S26 MapChunkBulk", "count", count, "dataLen", dataLen, "skyLight", skyLightSent)

	compressedData := make([]byte, dataLen)
	io.ReadFull(buf, compressedData)

	type chunkMeta struct {
		x, z		int32
		pMask, aMask	uint16
	}
	metadata := make([]chunkMeta, count)
	for i := 0; i < int(count); i++ {
		binary.Read(buf, binary.BigEndian, &metadata[i].x)
		binary.Read(buf, binary.BigEndian, &metadata[i].z)
		binary.Read(buf, binary.BigEndian, &metadata[i].pMask)
		binary.Read(buf, binary.BigEndian, &metadata[i].aMask)
	}

	r, err := zlib.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return err
	}
	defer r.Close()

	uncompressed, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	dataBuf := bytes.NewReader(uncompressed)

	for i := 0; i < int(count); i++ {
		m := metadata[i]
		pe := newPEChunkData()

		parseSections(dataBuf, m.pMask, skyLightSent == 1, pe)

		biomeData := make([]byte, 256)
		io.ReadFull(dataBuf, biomeData)
		for z := 0; z < 16; z++ {
			for x := 0; x < 16; x++ {
				pe.biomeColor[(z<<4)|x] = biomeGrassColor(biomeData[z*16+x])
			}
		}

		sendPEChunk(session, m.x, m.z, pe)
	}
	return nil
}

func MapBlockID(jeID byte) byte {
	return jeID
}

func biomeGrassColor(biomeID byte) uint32 {

	r, g, b := byte(0x91), byte(0xBD), byte(0x59)

	switch biomeID {
	case 0, 7, 11, 16:
		r, g, b = 0x8E, 0xB9, 0x71
	case 1, 24:
		r, g, b = 0x91, 0xBD, 0x59
	case 2, 17, 35, 36:
		r, g, b = 0xBF, 0xB7, 0x55
	case 3, 20, 34:
		r, g, b = 0x8A, 0xB6, 0x89
	case 4, 18, 132:
		r, g, b = 0x79, 0xC0, 0x5A
	case 5, 19, 33, 133:
		r, g, b = 0x86, 0xB7, 0x83
	case 6, 134:
		r, g, b = 0x6A, 0x70, 0x39
	case 10, 12, 13, 26, 140:
		r, g, b = 0x80, 0xB4, 0x97
	case 14, 15:
		r, g, b = 0x55, 0xC9, 0x3F
	case 21, 22, 23, 149:
		r, g, b = 0x59, 0xC9, 0x3C
	case 27, 28, 155, 156:
		r, g, b = 0x88, 0xBB, 0x67
	case 29, 157:
		r, g, b = 0x50, 0x7A, 0x32
	case 30, 31, 32, 158, 160, 161:
		r, g, b = 0x80, 0xB4, 0x97
	case 37, 38, 39, 163, 164, 165, 166, 167:
		r, g, b = 0x90, 0x81, 0x4D
	}

	return (uint32(biomeID) << 24) | (uint32(r) << 16) | (uint32(g) << 8) | uint32(b)
}

func handleS23(session xlat.Session, payload []byte) error {
	buf := bytes.NewReader(payload)
	var x int32
	binary.Read(buf, binary.BigEndian, &x)
	y, _ := buf.ReadByte()
	var z int32
	binary.Read(buf, binary.BigEndian, &z)
	blockId, _ := net.ReadVarInt(buf)
	meta, _ := buf.ReadByte()

	peBlockId := byte(blockId & 0xFF)

	pkt := protocol.NewUpdateBlockPacket(x, int32(y), z, peBlockId, meta)
	if sender, ok := session.GetPE().(PESender); ok {
		sender.SendPacket(pkt)
	}
	return nil
}

func handleS24(session xlat.Session, payload []byte) error {
	buf := bytes.NewReader(payload)
	var chunkX int32
	binary.Read(buf, binary.BigEndian, &chunkX)
	var chunkZ int32
	binary.Read(buf, binary.BigEndian, &chunkZ)
	var count int16
	binary.Read(buf, binary.BigEndian, &count)
	var dataLen int32
	binary.Read(buf, binary.BigEndian, &dataLen)

	if dataLen <= 0 || int(dataLen) != int(count)*4 {
		return nil
	}

	data := make([]byte, dataLen)
	io.ReadFull(buf, data)

	pkt := &protocol.UpdateBlockPacket{
		BasePacket:	protocol.BasePacket{PacketID: protocol.IDUpdateBlock},
		Records:	make([]protocol.BlockRecord, 0, count),
	}

	for i := 0; i < int(count); i++ {
		recordOff := i * 4
		coordShort := binary.BigEndian.Uint16(data[recordOff : recordOff+2])
		blockShort := binary.BigEndian.Uint16(data[recordOff+2 : recordOff+4])

		relX := int32((coordShort >> 12) & 15)
		relZ := int32((coordShort >> 8) & 15)
		relY := byte(coordShort & 255)

		blockId := byte((blockShort >> 4) & 0xFF)
		meta := byte(blockShort & 15)

		x := (chunkX * 16) + relX
		z := (chunkZ * 16) + relZ

		pkt.Records = append(pkt.Records, protocol.BlockRecord{
			X:		x,
			Z:		z,
			Y:		relY,
			BlockID:	blockId,
			BlockMeta:	meta,
			Flags:		protocol.UpdateBlockFlagNetwork | protocol.UpdateBlockFlagNeighborhood,
		})
	}

	if sender, ok := session.GetPE().(PESender); ok {
		sender.SendPacket(pkt)
	}
	return nil
}
