package utils

import (
        "bytes"
        "encoding/binary"
        "fmt"
        "scurrier/types"
        "io"
)



func Read4(r *bytes.Reader) (uint32, error) {

   var Value uint32
   if err := binary.Read(r, binary.LittleEndian, &Value); err != nil {
       return 0, err
   }  
   return Value, nil
}


func ReadString(r *bytes.Reader, Lenghth uint32) (string, error) {
    buf := make([]byte, Lenghth)
    if _, err := io.ReadFull(r, buf); err != nil {
        return "", err
    }
    return string(buf), nil
}


func Write4Bytes(w io.Writer, value uint32) error {
    return binary.Write(w, binary.LittleEndian, value)
}

func WriteString(w io.Writer, str string) error {
    Length := uint32(len(str) + 1)
    if err := Write4Bytes(w, Length); err != nil {
        return err
    }

    // Write the string bytes
    if _, err := w.Write([]byte(str)); err != nil {
        return err
    }
    if _, err := w.Write([]byte{0}); err != nil {
        fmt.Println("error writing \\0")
        return err
    }

    return nil
}



func CraftErrorResponse(code uint32) []byte {
    var buf bytes.Buffer
    Write4Bytes(&buf, code)
    return buf.Bytes()
}

func ParseOutputData(data []byte) error {
    // All CommandOut should atleast have 4 bytes, All Output Starts with CommandID(4 bytes)
    if len(data) < 4 {
        return nil
    }
    reader := bytes.NewReader(data)

    CommandID, err := Read4(reader)
    if err != nil { return err }
    

    OutputLen, err := Read4(reader)
    if err != nil { return err }


    OutputData, err := ReadString(reader, OutputLen)
    if err != nil { return err }

    fmt.Printf("Output[%d]: %s\n", CommandID, OutputData)

    return nil
}

func CraftCommandBytes(cmd *types.CommandReply) []byte {
        /*
           [Error/Success 4 bytes] 0 = okay, >0 = error or no command
           [commandID 4 bytes]
           [CommandCode 4 bytes]
           [param1 Length 4 bytes]
           [param1 N bytes]
           [param2 Length 4 bytes]
           [param2 N bytes]
        */

        var buf bytes.Buffer

        // 0 == no error, just makes it easier for the RAT to quickly know if error or not
        Write4Bytes(&buf, uint32(0))

        Write4Bytes(&buf, cmd.CommandID)
        Write4Bytes(&buf, cmd.CommandCode)

        if cmd.Param1 != "" {
                err := WriteString(&buf, cmd.Param1)
                if err != nil {
                    fmt.Println("error 1")
                    return nil
                }
        }

        if cmd.Param2 != "" {
                err := WriteString(&buf, cmd.Param2)
                if err != nil {
                    fmt.Println("error 1")
                    return nil
                }
        }

        return buf.Bytes()
}



func ParseRegisterBytes(data []byte) (*types.ClientInfo, error) {
    reader := bytes.NewReader(data)
    cmd := &types.ClientInfo{}


    GuidLen, _ := Read4(reader)
 
    cmd.Guid, _ = ReadString(reader, GuidLen)

    UserLen, _ := Read4(reader)
    cmd.Username, _ = ReadString(reader, UserLen)

 
    HostLen, _ := Read4(reader)
    cmd.Hostname, _ = ReadString(reader, HostLen)
 
    cmd.PID, _ = Read4(reader)
    cmd.Arch, _ = Read4(reader)


    VersionLen, _ := Read4(reader)
    cmd.Version, _ = ReadString(reader, VersionLen)

    return cmd, nil
}

func CreateRegisterResponseBytes(Token string, RefreshToken string) []byte {

	/* 
       [Error/Success]
       [token Length 4 bytes]
	   [token N bytes]
	   [refresh length 4 bytes]
	   [refresh token N bytes]
	   [for success]
	   [ErrorCode 4 bytes]
	   [for error response]
        */

	var buf bytes.Buffer


	Write4Bytes(&buf, uint32(0))

	WriteString(&buf, Token)

    WriteString(&buf, RefreshToken)


	return buf.Bytes()

}

