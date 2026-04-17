package net

import (
	"errors"
	"io"
)

var ErrVarIntTooBig = errors.New("VarInt is too big")

func ReadVarInt(r io.ByteReader) (int32, error) {
	var numRead int
	var result int32
	for {
		read, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		value := int32(read & 0b01111111)
		result |= (value << (7 * numRead))

		numRead++
		if numRead > 5 {
			return 0, ErrVarIntTooBig
		}

		if (read & 0b10000000) == 0 {
			break
		}
	}
	return result, nil
}

func WriteVarInt(w io.Writer, value int32) (int, error) {
	var bytes []byte
	uvalue := uint32(value)
	for {
		temp := byte(uvalue & 0b01111111)
		uvalue >>= 7
		if uvalue != 0 {
			temp |= 0b10000000
		}
		bytes = append(bytes, temp)
		if uvalue == 0 {
			break
		}
	}
	return w.Write(bytes)
}

func ReadString(r io.Reader) (string, error) {
	length, err := ReadVarInt(r.(io.ByteReader))
	if err != nil {
		return "", err
	}
	if length < 0 || length > 32767 {
		return "", errors.New("string length out of bounds")
	}
	buf := make([]byte, length)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func WriteString(w io.Writer, val string) error {
	bytes := []byte(val)
	_, err := WriteVarInt(w, int32(len(bytes)))
	if err != nil {
		return err
	}
	_, err = w.Write(bytes)
	return err
}
