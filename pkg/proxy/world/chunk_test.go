package world

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"testing"
)

func createMockChunkPayload() []byte {
	uncompressed := new(bytes.Buffer)

	uncompressed.Write(make([]byte, 65536))

	uncompressed.Write(make([]byte, 32768))

	uncompressed.Write(make([]byte, 32768))

	uncompressed.Write(make([]byte, 32768))

	uncompressed.Write(make([]byte, 256))

	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(uncompressed.Bytes())
	w.Close()
	compressedData := b.Bytes()

	payload := new(bytes.Buffer)
	binary.Write(payload, binary.BigEndian, int32(0))
	binary.Write(payload, binary.BigEndian, int32(0))
	payload.WriteByte(1)
	binary.Write(payload, binary.BigEndian, uint16(0xFFFF))
	binary.Write(payload, binary.BigEndian, uint16(0))
	binary.Write(payload, binary.BigEndian, int32(len(compressedData)))
	payload.Write(compressedData)

	return payload.Bytes()
}

func TestChunkTruncation(t *testing.T) {
	translator := &ChunkTranslator{}

	sender := &mockPESender{}
	session := &mockSession{mockSender: sender}

	payload := createMockChunkPayload()

	err := translator.Translate(session, 0x21, payload)
	if err != nil {
		t.Fatalf("Chunk translation failed to handle 256-height: %v", err)
	}

	if sender.lastPacket == nil {
		t.Fatalf("Expected FullChunkDataPacket to be output to PE terminal")
	}

}
