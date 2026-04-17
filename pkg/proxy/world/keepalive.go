package world

import (
	"github.com/scaxe/scaxe-go/pkg/proxy/xlat"
)

type KeepAliveTranslator struct{}

func (t *KeepAliveTranslator) Translate(session xlat.Session, packetID int32, payload []byte) error {
	jeConn := session.GetJE()
	if jeConn == nil {
		return nil
	}

	return jeConn.WritePacket(0x00, payload)
}
