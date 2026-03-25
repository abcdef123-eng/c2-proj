package bytehandler

import (
	"bytes"
	"encoding/binary"
	"io"
)

func Write4(w io.Writer, val uint32) error {
	return binary.Write(w, binary.LittleEndian, val)
}

func WriteString(w io.Writer, str string) error {
	Len := uint32(len(str))
	if err := Write4(w, Len); err != nil {
		return err
	}
	if _, err := w.Write([]byte(str)); err != nil {
		return err
	}
	return nil
}

func CraftJwtResponse(token string) ([]byte, error) {

	var buf bytes.Buffer

	if err := WriteString(&buf, token); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil

}
