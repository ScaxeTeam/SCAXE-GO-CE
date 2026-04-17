package auth

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/scaxe/scaxe-go/pkg/protocol"
	pnet "github.com/scaxe/scaxe-go/pkg/proxy/net"
	"github.com/scaxe/scaxe-go/pkg/proxy/xlat"
)

type PELoginTranslator struct {
	JEAddress	string
	JEPort		uint16
}

type PESender interface {
	SendPacket(packet protocol.DataPacket)
}

func (t *PELoginTranslator) Translate(session xlat.Session, pePkt protocol.DataPacket) error {
	loginPkt, ok := pePkt.(*protocol.LoginPacket)
	if !ok {
		return fmt.Errorf("expected LoginPacket")
	}

	addr := fmt.Sprintf("%s:%d", t.JEAddress, t.JEPort)
	jeConn, err := pnet.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to dial JE: %v", err)
	}

	session.SetJE(jeConn)
	session.SetState(1)

	err = sendHandshake(jeConn, 5, t.JEAddress, t.JEPort, 2)
	if err != nil {
		return err
	}

	err = sendLoginStart(jeConn, loginPkt.Username)
	if err != nil {
		return err
	}

	return nil
}

func sendHandshake(conn *pnet.Conn, protocolVer int32, address string, port uint16, nextState int32) error {
	buf := new(bytes.Buffer)
	pnet.WriteVarInt(buf, protocolVer)
	pnet.WriteString(buf, address)
	binary.Write(buf, binary.BigEndian, port)
	pnet.WriteVarInt(buf, nextState)

	return conn.WritePacket(0x00, buf.Bytes())
}

func sendLoginStart(conn *pnet.Conn, username string) error {
	buf := new(bytes.Buffer)
	pnet.WriteString(buf, username)
	return conn.WritePacket(0x00, buf.Bytes())
}

type JELoginTranslator struct{}

func (t *JELoginTranslator) Translate(session xlat.Session, packetID int32, payload []byte) error {
	state := session.GetState()

	if state == 1 {
		if packetID == 0x02 {

			session.SetState(2)

			if sender, ok := session.GetPE().(PESender); ok {
				fmt.Println("[DEBUG] PESender matched! Sending PlayStatus = 0 to PE Client!")
				statusPkt := protocol.NewPlayStatusPacket()
				statusPkt.Status = protocol.PlayStatusLoginSuccess
				sender.SendPacket(statusPkt)
			} else {
				fmt.Println("[ERROR] session.GetPE() failed to cast to PESender! Type mismatch!")
			}
			return nil
		} else if packetID == 0x00 {
			return fmt.Errorf("login disconnected by JE back-end")
		} else if packetID == 0x01 {
			return fmt.Errorf("online-mode MUST be false on backend server")
		}
	} else if state == 2 {
		if packetID == 0x01 {
			return sendStartGameToPE(session, payload)
		}
	}
	return nil
}

func sendStartGameToPE(session xlat.Session, payload []byte) error {
	buf := bytes.NewReader(payload)
	var entityID int32
	binary.Read(buf, binary.BigEndian, &entityID)

	gamemode, _ := buf.ReadByte()
	dimension, _ := buf.ReadByte()

	sg := protocol.NewStartGamePacket()
	sg.Seed = 0
	sg.Dimension = byte(dimension)
	sg.Generator = 1
	sg.Gamemode = int32(gamemode)
	if player, ok := session.GetPE().(interface{ GetID() int64 }); ok {
		sg.EntityID = player.GetID()
		sg.RuntimeID = player.GetID()
	} else {
		sg.EntityID = int64(entityID)
		sg.RuntimeID = int64(entityID)
	}
	sg.X = 0
	sg.Y = 70
	sg.Z = 0
	sg.SpawnX = 0
	sg.SpawnY = 70
	sg.SpawnZ = 0
	sg.LevelID = "world"

	if sender, ok := session.GetPE().(PESender); ok {
		sender.SendPacket(sg)
	}

	session.SetState(3)
	return nil
}
