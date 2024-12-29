package sudoku

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"log"
)

const (
	Version1 = 0x01
)

// response status list
const (
	StatusOK                  = 0x00
	StatusBadRequest          = 0x01
	StatusUnauthorized        = 0x02
	StatusForbidden           = 0x03
	StatusTimeout             = 0x04
	StatusServiceUnavailable  = 0x05
	StatusHostUnreachable     = 0x06
	StatusNetworkUnreachable  = 0x07
	StatusInternalServerError = 0x08
	NotUnique                 = 0x09
)

const (
	ObfDomain = "www.bing.com"
	ObfPort   = 80
)

var (
	ErrBadVersion = errors.New("bad version")
	ErrNotUnique  = "not unique"
)

// Request is a sudoku client request.
//
// Protocol spec:
//
// +---------+-----+---------+---------+----------+----------+
// |TLS OBF  | VER | SB CODE | OBF LEN | OBF PORT | OBF ADDR |
// +---------+-----+---------+---------+----------+----------+
// |3        | 1   | 1       | 1       | 2        | VAR      |
// +---------+-----+---------+---------+----------+----------+
//
// TLS OBF - TLS obfuscation, 3 bytes.
// VER - protocol version, 1 byte.
// SB CODE - sudoku code, 1 byte.
// OBF LEN - obfuscated address length, 1 byte.
// OBF PORT - obfuscated port, 2 bytes.
// OBF ADDR - obfuscated address, variable length.

type Request struct {
	TlsObf  [3]byte
	Version uint8
	Code    uint8
	ObfLen  uint8
	ObfPort uint16
	ObfAddr []byte
}

// default Request
var DefaultRequest = &Request{
	TlsObf:  [3]byte{0x16, 0x03, 0x03},
	Version: Version1,
	Code:    0x01,
	ObfLen:  uint8(len(ObfDomain)),
	ObfPort: uint16(ObfPort),
	ObfAddr: []byte(ObfDomain),
}

func (req *Request) ReadFrom(r io.Reader) (n int64, err error) {
	var header [8]byte
	nn, err := io.ReadFull(r, header[:])
	n += int64(nn)
	if err != nil {
		return
	}
	// 保存前三个字节到 TlsObf
	copy(req.TlsObf[:], header[0:3])
	req.Version = header[3]
	if req.Version != Version1 {
		err = ErrBadVersion
		return
	}
	req.Code = header[4]
	req.ObfLen = header[5]
	req.ObfPort = binary.BigEndian.Uint16(header[6:8])
	req.ObfAddr = make([]byte, req.ObfLen)
	nn, err = io.ReadFull(r, req.ObfAddr)
	n += int64(nn)
	if err != nil {
		return
	}
	// 读完之后打log
	log.Printf("sudoku request: %v", req.Bytes())
	return
}

func (req *Request) WriteTo(w io.Writer) (n int64, err error) {
	var buf bytes.Buffer

	buf.Write(req.TlsObf[:])
	buf.WriteByte(req.Version)
	buf.WriteByte(req.Code)
	buf.WriteByte(req.ObfLen)
	binary.Write(&buf, binary.BigEndian, req.ObfPort)
	buf.Write(req.ObfAddr)

	return buf.WriteTo(w)
}

func (r *Request) Bytes() []byte {
	buf := make([]byte, 8+len(r.ObfAddr))
	copy(buf[0:3], r.TlsObf[:])
	buf[3] = r.Version
	buf[4] = r.Code
	buf[5] = r.ObfLen
	binary.BigEndian.PutUint16(buf[6:8], r.ObfPort)
	copy(buf[8:], r.ObfAddr)
	return buf
}

// Response is a relay server response.
//
// Protocol spec:
//
// +---------+-----+------+---------+
// |TLS OBF  | VER | STAT | SB CODE |
// +---------+-----+------+---------+
// |3        | 1   | 1    | 1       |
// +---------+-----+------+---------+
//
// TLS OBF - TLS obfuscation, 3 bytes.
// VER - protocol version, 1 byte.
// STAT - status code, 1 byte.
// SB CODE - sudoku code, 1 byte.

type Response struct {
	TlsObf  [3]byte
	Version uint8
	Status  uint8
	Code    uint8
}

func (resp *Response) ReadFrom(r io.Reader) (n int64, err error) {
	var header [6]byte
	nn, err := io.ReadFull(r, header[:])
	n += int64(nn)
	if err != nil {
		return
	}
	// 保存前三个字节到 TlsObf
	copy(resp.TlsObf[:], header[0:3])

	if header[3] != Version1 {
		err = ErrBadVersion
		return
	}
	resp.Version = header[3]
	resp.Status = header[4]
	resp.Code = header[5]

	// 读完之后打log
	log.Printf("sudoku response: %v", resp.Bytes())

	return
}

func (resp *Response) WriteTo(w io.Writer) (n int64, err error) {
	var buf bytes.Buffer

	buf.Write(resp.TlsObf[:])
	buf.WriteByte(resp.Version)
	buf.WriteByte(resp.Status)
	buf.WriteByte(resp.Code)

	return buf.WriteTo(w)
}

func (r *Response) Bytes() []byte {
	buf := make([]byte, 6)
	copy(buf[0:3], r.TlsObf[:])
	buf[3] = r.Version
	buf[4] = r.Status
	buf[5] = r.Code
	return buf
}
