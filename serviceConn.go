package ymtcp

import (
	"fmt"
	"net"
	"runtime"
	"sync"
)

type Conn struct {
	srv         *TcpService
	conn        *net.TCPConn
	closeOnce   sync.Once
	closed      bool
	closeChan   chan struct{}
	sendChan    chan []byte
	receiveChan chan []byte
}

func NewConn(conn *net.TCPConn, srv *TcpService) *Conn {
	return &Conn{
		srv:         srv,
		conn:        conn,
		closeChan:   make(chan struct{}),
		sendChan:    make(chan []byte, srv.config.SendLimit),
		receiveChan: make(chan []byte, srv.config.ReceiveLimit),
	}
}

func (c *Conn) Close() {
	c.closeOnce.Do(func() {
		c.closed = true
		close(c.closeChan)
		c.conn.Close()
		c.srv.callback.CloseCome(c)
	})
}

func (c *Conn) GetConnRemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Conn) IsClosed() bool {
	return c.closed
}

func (c *Conn) WriteBytesToChan(bytes []byte) (err error) {
	if c.IsClosed() {
		err = fmt.Errorf("use of closed network connection")
		return
	}

	select {
	case c.sendChan <- bytes:
		return
	case <-c.closeChan:
		err = fmt.Errorf("use of closed network connection")
		return
	}
}

func (c *Conn) Go() {
	if !c.srv.callback.ConnectCome(c) {
		return
	}

	go c.ReadInLoop()
	go c.WriteInLoop()
	go c.CallBackInLoop()
}

func (c *Conn) ReadInLoop() {
	c.srv.wg.Add(1)
	var err error
	defer func() {
		if err != nil {
			fmt.Println("ReadInLoop occur err:", err)
		}

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

			fmt.Println(st(false))
		}

		c.Close()
		c.srv.wg.Done()
	}()

	for {
		select {
		case <-c.srv.chanExit:
			return
		case <-c.closeChan:
			return
		default:
		}

		bytes, err := c.srv.protocol.ReadBytes(c.conn)
		if err != nil {
			return
		}

		c.receiveChan <- bytes
	}
}

func (c *Conn) WriteInLoop() {
	c.srv.wg.Add(1)
	var err error

	defer func() {
		if err != nil {
			fmt.Println("WriteInLoop occur err:", err)
		}

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

			fmt.Println(st(false))
		}

		c.Close()
		c.srv.wg.Done()
	}()

	for {
		select {
		case <-c.srv.chanExit:
			return
		case <-c.closeChan:
			return
		case bytes := <-c.sendChan:
			if _, err := c.conn.Write(bytes); err != nil {
				return
			}
		}
	}
}

func (c *Conn) CallBackInLoop() {

	c.srv.wg.Add(1)
	var err error

	defer func() {
		if err != nil {
			fmt.Println("CallBackInLoop occur err:", err)
		}

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

			fmt.Println(st(false))
		}

		c.Close()
		c.srv.wg.Done()
	}()

	for {
		select {
		case <-c.srv.chanExit:
			return
		case <-c.closeChan:
			return
		case bytes := <-c.receiveChan:
			if err = c.srv.callback.MessageCome(c, bytes); err != nil {
				return
			}
		}
	}
}
