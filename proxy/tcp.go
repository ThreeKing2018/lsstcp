package proxy

import (
	"context"
	"fmt"
	"github.com/ThreeKing2018/goutil/golog"
	"io"
	"net"
	"syscall"
	"time"
)



type TCP struct {
	listenAddr string
	remoteAddr string
	state bool
	ctx context.Context
}

func (tcp *TCP) SetremoteAddr(remoteAddr string) {
	tcp.remoteAddr = remoteAddr
}

func NewTCP(ctx context.Context,localAddr,remoteAddr string) *TCP {
	 t :=&TCP{
		listenAddr:localAddr,
		remoteAddr:remoteAddr,
		state:false,
		ctx:ctx,
	}

	 go t.Start()
	 return t
}

func (tcp *TCP) Start() (err error) {
	ln, err := net.Listen("tcp", tcp.listenAddr)

	if err != nil {
		panic(err)
	}

	defer func() {
		if !tcp.state {
			ln.Close()
		}
	}()
	go func() {
		select {
		case <-tcp.ctx.Done():
			ln.Close()
			tcp.state = true
			golog.Infow("tcp server close",
				"addr",tcp.listenAddr)
		}
	}()


	golog.Debugw("tcp 监听地址",
		"ipaddr",tcp.listenAddr)


	var tempDelay time.Duration
	for {
		conn, err := ln.Accept()
		if err != nil {
			/*如果错误是暂时的,那么sleep一定时间在提供服务,否则就直接 return退出程序*/
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}

				golog.Error(fmt.Sprintf("accept error: %s; 本次 sleep 时间 %v", err.Error(), tempDelay))

				time.Sleep(tempDelay)
				continue
			}
			return err
		}

		go tcp.handler(conn)
	}

	return
}


func (tcp *TCP) handler(conn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			golog.Error("[tcp]--recover-- ",err)
		}
	}()

	//连接远端
	remote, err := net.Dial("tcp",tcp.remoteAddr)

	if err != nil {
		if ne, ok := err.(*net.OpError); ok &&
			(ne.Err == syscall.EMFILE || ne.Err == syscall.ENFILE) {
			// log too many open file error
			// EMFILE is process reaches open file limits, ENFILE is system limit
			golog.Error("dial error:too many open file error","err", err)
		} else {
			golog.Warnw("warn connecting",
				"remoteaddr",tcp.remoteAddr,
				"err",err)
		}
		return
	}

	//仿照 shadowsocks
	go pipeThenClose(conn, remote)
	pipeThenClose(remote, conn)

	return

}




/*src读取数据 写入 dst*/
func pipeThenClose(src,dst net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			golog.Errorw("--recover-- ","err",err)
		}
	}()

	//var buf = make([]byte,4096)
	buf := bufPool.Get()

	defer func() {
		dst.Close()
		bufPool.Put(buf)
	}()


	var (
		err error
		n int
	)
	for {
		n, err = src.Read(buf)
		if n > 0 {
			//响应的大小
			_, err = dst.Write(buf[0:n])
			if err != nil {
				golog.Warnw("","err",err)
				break
			}
			golog.Debugw("","dst write byte",n)

		}

		if err != nil || n == 0 {
			// Always "use of closed network connection", but no easy way to
			// identify this specific error. So just leave the error along for now.
			// More info here: https://code.google.com/p/go/issues/detail?id=4373
			if err != io.EOF {
				golog.Debugw("err != nil  n == 0 ","err",err)
			}
			break
		}
	}

}


