package main

import (
	"context"
	. "lsstcp"
	"lsstcp/proxy"
	"os"
	"os/signal"
	"syscall"
)


type a struct {
	cancel context.CancelFunc
	value  string
	tcp   *proxy.TCP
}
func main() {
	Init()
	conf := GetConfigInstance()
	tcprun := make(map[string]a)

	fn := func(k,v string) {
		ctx,cancel := context.WithCancel(context.Background())
		tcp := proxy.NewTCP(ctx,k,v)
		tcprun[k] = a{cancel:cancel,value:v,tcp:tcp}
	}

	for k,v := range conf.GetStringMapString("tcp") {
		fn(k,v)
	}


	stopChan := make(chan struct{})
	go func() {
		for {
			select {
			case <- ConfWatch(stopChan):
				conf.ReadConfig()
				tcpmap := conf.GetStringMapString("tcp")
				//TODO 对比 运行新的,修改旧的
				for k,v := range tcpmap {
					if v1 , ok := tcprun[k] ;ok {
						if v1.value == v {
							continue
						}
						v1.tcp.SetremoteAddr(v)

					}else {
						fn(k,v)
					}
				}

				//TODO 删除old
				for k,v := range tcprun {
					if _, ok := tcpmap[k]; !ok {
						v.cancel()
						delete(tcprun,k)
					}
				}

			}
		}

	}()

	waitSignal()
	stopChan <- struct{}{}
	//关闭tcp监听端口 回收资源
	for k,v := range tcprun {
		v.cancel()
		delete(tcprun,k)
	}

}



//阻塞，只有执行信号才执行
func waitSignal() {
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-osSignals
}


