package main

import (
	"fmt"
	"github.com/iamyh/ymtcp"
	"net"
	"time"
)

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:8989")
	if err != nil {
		fmt.Println("err when new tcpAddr:", err)
		return
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println("err when new conn:", err)
		return
	}

	ep := &ymtcp.EchoProtocolImpl{}
	for i := 0; i < 5; i++ {
		buff, err := ymtcp.MakePacket([]byte("hello"))
		if err != nil {
			fmt.Println("make packet err:", err)
			continue
		}

		conn.Write(buff)
		bytes, err := ep.ReadBytes(conn)
		if err == nil {
			fmt.Println("sever reply:", string(bytes))
		} else {
			fmt.Println("server err:", err)
		}

		time.Sleep(1 * time.Second)
	}

	conn.Close()
}
