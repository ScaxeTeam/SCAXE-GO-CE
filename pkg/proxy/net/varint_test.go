package net

import (
	"bytes"
	"testing"
)

func TestReadWriteVarInt(t *testing.T) {
	tests := []int32{
		0,
		1,
		127,
		128,
		255,
		2147483647,
		-1,
		-100,
		-2147483648,
	}

	for _, tc := range tests {
		buf := new(bytes.Buffer)
		_, err := WriteVarInt(buf, tc)
		if err != nil {
			t.Fatalf("Failed to write %d: %v", tc, err)
		}

		decoded, err := ReadVarInt(buf)
		if err != nil {
			t.Fatalf("Failed to read: %v", err)
		}

		if decoded != tc {
			t.Errorf("Expected %d, got %d", tc, decoded)
		}
	}
}

func TestVarIntTooBig(t *testing.T) {

	badData := []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01}
	buf := bytes.NewReader(badData)
	_, err := ReadVarInt(buf)
	if err != ErrVarIntTooBig {
		t.Errorf("Expected ErrVarIntTooBig, got %v", err)
	}
}
