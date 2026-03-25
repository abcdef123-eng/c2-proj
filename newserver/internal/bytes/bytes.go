package bytehandler

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type RegisterClientData struct {
	Guid       string
	Username   string
	Hostname   string
	Arch       byte
	Ip         string
	WinVersion string
	Pid        uint32
}

type Reader struct {
	r   *bytes.Reader
	err error
}

func (r *Reader) Read4() uint32 {
	if r.err != nil {
		return 0
	}
	var val uint32
	r.err = binary.Read(r.r, binary.LittleEndian, &val)
	return val
}

func (r *Reader) Read1() byte {
	if r.err != nil {
		return 0
	}
	val, err := r.r.ReadByte()
	r.err = err
	return val
}

func (r *Reader) Read2() uint16 {
	if r.err != nil {
		return 0
	}
	var val uint16
	r.err = binary.Read(r.r, binary.LittleEndian, &val)
	return val
}

func (r *Reader) ReadString(len uint32) string {
	if r.err != nil {
		return ""
	}
	buf := make([]byte, len)
	_, r.err = io.ReadFull(r.r, buf)
	return string(buf)
}

/*
 * [Guid Length] 4 bytes
 * [Gudi string] N bytes
 * [Username length] 4 bytes
 * [username string] N bytes
 * [Hostname length] 4 bytes
 * [Hostname string] N bytes
 * [arch] 1 byte
 * [Major] 2 bytes
 * [Minor] 2 bytes
 * [build] 2 bytes
 * [pid] 4 bytes
 */
func ParseClientRegister(r *bytes.Reader, ip string) (RegisterClientData, error) {
	rd := &Reader{r: r}

	guid := rd.ReadString(rd.Read4())
	username := rd.ReadString(rd.Read4())
	hostname := rd.ReadString(rd.Read4())
	arch := rd.Read1()
	major := rd.Read2()
	minor := rd.Read2()
	build := rd.Read2()
	pid := rd.Read4()

	if rd.err != nil {
		return RegisterClientData{}, rd.err
	}

	return RegisterClientData{
		Guid:       guid,
		Username:   username,
		Hostname:   hostname,
		Arch:       arch,
		Ip:         ip,
		WinVersion: ParseWindowsVersion(int(major), int(minor), int(build)),
		Pid:        pid,
	}, nil
}

var windowsVersions = map[string]string{
	"10.0.19041": "Windows 10 20H1",
	"10.0.19042": "Windows 10 20H2",
	"10.0.19043": "Windows 10 21H1",
	"10.0.19044": "Windows 10 21H2",
	"10.0.19045": "Windows 10 22H2",
	"10.0.22000": "Windows 11 21H2",
	"10.0.22621": "Windows 11 22H2",
	"10.0.22631": "Windows 11 23H2",
	"10.0.26100": "Windows 11 24H2",
	"6.1.7601":   "Windows 7 SP1",
	"6.2.9200":   "Windows 8",
	"6.3.9600":   "Windows 8.1",
}

func ParseWindowsVersion(major, minor, build int) string {
	key := fmt.Sprintf("%d.%d.%d", major, minor,
		build)
	if v, ok := windowsVersions[key]; ok {
		return v
	}
	return key // fallback to raw version string
}
