package ymtcp

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

type TcpConfig struct {
	Port         int
	SendLimit    int
	ReceiveLimit int
}

type TcpService struct {
	listener *net.TCPListener
	config   TcpConfig
	protocol Protocol
	callback CallBack
	chanExit chan struct{}
	wg       sync.WaitGroup
}

type Protocol interface {
	ReadBytes(conn *net.TCPConn) ([]byte, error)
}

type CallBack interface {
	ConnectCome(*Conn) bool
	MessageCome(*Conn, []byte) error
	CloseCome(*Conn)
}

func NewService(config TcpConfig, protocol Protocol, callback CallBack) *TcpService {
	return &TcpService{
		config:   config,
		protocol: protocol,
		callback: callback,
		chanExit: make(chan struct{}),
		wg:       sync.WaitGroup{},
	}
}

func (s *TcpService) Init() error {
	addr, err := net.ResolveTCPAddr("tcp4", ":"+strconv.Itoa(s.config.Port))
	if err != nil {
		return err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	s.listener = listener

	return nil
}

func (s *TcpService) Service() {

	s.wg.Add(1)
	defer func() {
		s.listener.Close()
		s.wg.Done()
	}()

	for {
		select {
		case <-s.chanExit:
			return
		default:
		}

		conn, err := s.listener.AcceptTCP()
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				fmt.Println("stop tcp accept tpc connection")
			}

			continue
		}

		//run
		go NewConn(conn, s).Go()
	}
}

func (s *TcpService) InitFromFD(fd uintptr) error {

	file := os.NewFile(fd, "/tmp/ymtcp-restart")
	listener, err := net.FileListener(file)

	if err != nil {
		return err
	}

	tcpListener, ok := listener.(*net.TCPListener)
	if !ok {
		return fmt.Errorf("can not convert to *net.TCPListener")
	}

	s.listener = tcpListener
	return nil
}

func (s *TcpService) GetListenerFD() (uintptr, error) {
	file, err := s.listener.File()
	if err != nil {
		return 0, err
	}

	return file.Fd(), nil
}

func (s *TcpService) StopAccept() {
	s.listener.SetDeadline(time.Now())
	close(s.chanExit)
}

func (s *TcpService) WaitAndDone() {
	s.wg.Wait()
}
