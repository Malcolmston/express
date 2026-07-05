// Package xid generates globally unique, sortable 12-byte identifiers encoded as
// 20-character base32 strings, a Go port of the "xid" scheme.
package xid

import (
	"crypto/md5"
	"crypto/rand"
	"errors"
	"os"
	"sync/atomic"
)

// encoding is the base32-hex lowercase alphabet used by xid.
const encoding = "0123456789abcdefghijklmnopqrstuv"

var dec [256]byte

var (
	machineID [3]byte
	pid       uint16
	counter   uint32
)

func init() {
	for i := range dec {
		dec[i] = 0xff
	}
	for i := 0; i < len(encoding); i++ {
		dec[encoding[i]] = byte(i)
	}
	machineID = readMachineID()
	pid = uint16(os.Getpid())
	var seed [4]byte
	if _, err := rand.Read(seed[:]); err == nil {
		counter = uint32(seed[0])<<16 | uint32(seed[1])<<8 | uint32(seed[2])
	}
}

func readMachineID() [3]byte {
	var id [3]byte
	host, err := os.Hostname()
	if err == nil && host != "" {
		sum := md5.Sum([]byte(host))
		copy(id[:], sum[:3])
		return id
	}
	rand.Read(id[:])
	return id
}

// New generates an xid for the given unix seconds using the process machine id,
// pid, and an atomically incrementing counter.
func New(unixSeconds int64) string {
	c := atomic.AddUint32(&counter, 1)
	return NewWithData(unixSeconds, machineID, pid, c)
}

// NewWithData builds an xid deterministically from all its components.
func NewWithData(unixSeconds int64, machine [3]byte, p uint16, c uint32) string {
	var b [12]byte
	t := uint32(unixSeconds)
	b[0] = byte(t >> 24)
	b[1] = byte(t >> 16)
	b[2] = byte(t >> 8)
	b[3] = byte(t)
	b[4] = machine[0]
	b[5] = machine[1]
	b[6] = machine[2]
	b[7] = byte(p >> 8)
	b[8] = byte(p)
	b[9] = byte(c >> 16)
	b[10] = byte(c >> 8)
	b[11] = byte(c)
	return encode(b)
}

// encode renders 12 bytes as a 20-char base32-hex string.
func encode(b [12]byte) string {
	s := make([]byte, 20)
	s[0] = encoding[b[0]>>3]
	s[1] = encoding[(b[0]&0x07)<<2|b[1]>>6]
	s[2] = encoding[(b[1]>>1)&0x1f]
	s[3] = encoding[(b[1]&0x01)<<4|b[2]>>4]
	s[4] = encoding[(b[2]&0x0f)<<1|b[3]>>7]
	s[5] = encoding[(b[3]>>2)&0x1f]
	s[6] = encoding[(b[3]&0x03)<<3|b[4]>>5]
	s[7] = encoding[b[4]&0x1f]
	s[8] = encoding[b[5]>>3]
	s[9] = encoding[(b[5]&0x07)<<2|b[6]>>6]
	s[10] = encoding[(b[6]>>1)&0x1f]
	s[11] = encoding[(b[6]&0x01)<<4|b[7]>>4]
	s[12] = encoding[(b[7]&0x0f)<<1|b[8]>>7]
	s[13] = encoding[(b[8]>>2)&0x1f]
	s[14] = encoding[(b[8]&0x03)<<3|b[9]>>5]
	s[15] = encoding[b[9]&0x1f]
	s[16] = encoding[b[10]>>3]
	s[17] = encoding[(b[10]&0x07)<<2|b[11]>>6]
	s[18] = encoding[(b[11]>>1)&0x1f]
	s[19] = encoding[(b[11]&0x01)<<4]
	return string(s)
}

// Decode parses a 20-char xid string back into its 12 raw bytes.
func Decode(id string) ([12]byte, error) {
	var b [12]byte
	if len(id) != 20 {
		return b, errors.New("xid: invalid length")
	}
	var v [20]byte
	for i := 0; i < 20; i++ {
		d := dec[id[i]]
		if d == 0xff {
			return b, errors.New("xid: invalid character")
		}
		v[i] = d
	}
	b[0] = v[0]<<3 | v[1]>>2
	b[1] = v[1]<<6 | v[2]<<1 | v[3]>>4
	b[2] = v[3]<<4 | v[4]>>1
	b[3] = v[4]<<7 | v[5]<<2 | v[6]>>3
	b[4] = v[6]<<5 | v[7]
	b[5] = v[8]<<3 | v[9]>>2
	b[6] = v[9]<<6 | v[10]<<1 | v[11]>>4
	b[7] = v[11]<<4 | v[12]>>1
	b[8] = v[12]<<7 | v[13]<<2 | v[14]>>3
	b[9] = v[14]<<5 | v[15]
	b[10] = v[16]<<3 | v[17]>>2
	b[11] = v[17]<<6 | v[18]<<1 | v[19]>>4
	return b, nil
}

// Time decodes the unix-seconds timestamp encoded in an xid string.
func Time(id string) (int64, error) {
	b, err := Decode(id)
	if err != nil {
		return 0, err
	}
	t := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	return int64(t), nil
}
