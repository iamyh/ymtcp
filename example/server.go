package main

import (
	"fmt"
	"github.com/iamyh/ymtcp"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

type CallBack struct{}

func (this *CallBack) ConnectCome(c *ymtcp.Conn) bool {
	addr := c.GetConnRemoteAddr()
	fmt.Println("connection come:", addr.String())
	return true
}

func (this *CallBack) MessageCome(c *ymtcp.Conn, bytes []byte) error {
	addr := c.GetConnRemoteAddr()
	fmt.Printf("message come from:[%s] [%s]\n", addr.String(), string(bytes))
	buff, err := ymtcp.MakePacket(bytes)
	if err != nil {
		return err
	}

	c.WriteBytesToChan(buff)
	return nil
}

func (this *CallBack) CloseCome(c *ymtcp.Conn) {
	addr := c.GetConnRemoteAddr()
	fmt.Println("conntion close:", addr.String())
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	config := ymtcp.TcpConfig{
		Port:         8989,
		SendLimit:    88,
		ReceiveLimit: 88,
	}

	srv := ymtcp.NewService(config, &ymtcp.EchoProtocolImpl{}, &CallBack{})
	var err error
	if os.Getenv("YMTCP_RESTART") == "true" {
		err = srv.InitFromFD(3)
	} else {
		err = srv.Init()
	}

	if err != nil {
		log.Fatalf("tcp service init err:%s", err)
	}

	go srv.Service()
	fmt.Println("listen:", config.Port)

	chanSignal := make(chan os.Signal)
	signal.Notify(chanSignal, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	for receiveSignal := range chanSignal {
		if receiveSignal == syscall.SIGTERM {

			log.Println("receive signal SIGTERM,stop service.pid:", os.Getpid())
			srv.StopAccept()
			srv.WaitAndDone()
			os.Exit(0)
		} else {
			log.Println("receive signal SIGHUP")

			//停止接受新连接
			srv.StopAccept()
			fd, err := srv.GetListenerFD()
			if err != nil {
				log.Fatalf("get tcp service fd err:%s", err)
			}

			//设置标记，用来子进程的开启判断,见上述
			os.Setenv("YMTCP_RESTART", "true")

			execSpec := &syscall.ProcAttr{
				Env: os.Environ(), //父进程环境
				//标准输入0，标准输出1，标准错误2,tcp listener fd 为3，用于上述的子进程开启Initfromfd
				Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd(), fd},
			}

			//产生了子进程
			fork, err := syscall.ForkExec(os.Args[0], os.Args, execSpec)
			if err != nil {
				log.Fatalln("Fail to fork", err)
			}

			log.Printf("pid:%d restart graceful,new pid:%d start", os.Getpid(), fork)
			//等待父进程处理完所有已有的连接
			srv.WaitAndDone()

			//退出父进程

			os.Exit(0)
		}
	}
}
