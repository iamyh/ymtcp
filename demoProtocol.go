package ymtcp

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
)

const (
	HeadLength    = 4
	MaxBodyLength = 1024
)

type EchoProtocolImpl struct {
}

func MakePacket(bytes []byte) ([]byte, error) {

	bodyLength := len(bytes)
	if bodyLength > MaxBodyLength {
		err := fmt.Errorf("packet's size %d is too large than %d", bodyLength, MaxBodyLength)
		return nil, err
	}

	buff := make([]byte, HeadLength+len(bytes))
	binary.BigEndian.PutUint32(buff[0:HeadLength], uint32(len(bytes)))

	copy(buff[HeadLength:], bytes)
	return buff, nil
}

func (ep *EchoProtocolImpl) ReadBytes(conn *net.TCPConn) ([]byte, error) {

	defer func() {
		if x := recover(); x != nil {
			var st = func(all bool) string {
				// Reserve 1K buffer at first
				buf := make([]byte, 512)

				for {
					size := runtime.Stack(buf, all)
					// The size of the buffer may be not enough to hold the stacktrace,
					// so double the buffer size
					if size == len(buf) {
						buf = make([]byte, len(buf)<<1)
						continue
					}
					break
				}

				return string(buf)
			}
			log.Println(st(false))
		}

	}()

	lengthBytes := make([]byte, HeadLength)
	if _, err := io.ReadFull(conn, lengthBytes); err != nil {
		return nil, err
	}

	var length int
	if length = int(binary.BigEndian.Uint32(lengthBytes)); length > MaxBodyLength {
		return nil, fmt.Errorf("the size of packet %d is larger than the limit %d", length, MaxBodyLength)
	}

	buff := make([]byte, length)
	if _, err := io.ReadFull(conn, buff); err != nil {
		return nil, err
	}

	return buff, nil
}
