package bytehandler

import (
	"bytes"
	"encoding/binary"
	"io"
)

type TaskData struct {
	TaskID      int32
	CommandCode int
	Param1      string
	Param2      string
}

func Write4(w io.Writer, val any) error {
	return binary.Write(w, binary.LittleEndian, val)
}

func WriteString(w io.Writer, str string) error {
	strLen := uint32(len(str))
	if err := Write4(w, strLen); err != nil {
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

var CommandParamCount = map[int]int{
	0x1: 1, // ls
	0x2: 1, // cd
	0x3: 1, // rm
	0x4: 2, // mv
	0x5: 1, // cat
	0x6: 0, // get-privs
}

func CraftTaskResponse(Tasks []TaskData) ([]byte, error) {
	var buffer bytes.Buffer
	for _, c := range Tasks {
		err := Write4(&buffer, c.TaskID)
		if err != nil {
			return nil, err
		}
		err = Write4(&buffer, int32(c.CommandCode))
		if err != nil {
			return nil, err
		}
		ParamCount := CommandParamCount[c.CommandCode]
		if ParamCount >= 1 {
			if err := WriteString(&buffer, c.Param1); err != nil {
				return nil, err
			}
		}
		if ParamCount == 2 {
			if err := WriteString(&buffer, c.Param2); err != nil {
				return nil, err
			}
		}
	}

	return buffer.Bytes(), nil
}
