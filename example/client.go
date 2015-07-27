package main

import (
	"github.com/iamyh/ymtcp"
	"log"
	"net"
	"time"
)

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:8989")
	if err != nil {
		log.Println("err when new tcpAddr:", err)
		return
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Println("err when new conn:", err)
		return
	}

	ep := &ymtcp.EchoProtocolImpl{}
	for i := 0; i < 5; i++ {
		buff, err := ymtcp.MakePacket([]byte("hello"))
		if err != nil {
			log.Println("make packet err:", err)
			continue
		}

		conn.Write(buff)
		bytes, err := ep.ReadBytes(conn)
		if err == nil {
			log.Println("sever reply:", string(bytes))
		} else {
			log.Println("server err:", err)
		}

		time.Sleep(1 * time.Second)
	}

	conn.Close()
}
