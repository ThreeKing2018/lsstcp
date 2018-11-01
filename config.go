package serverRoom

import (
	"github.com/fsnotify/fsnotify"
	goconfig "gogs.163.com/feiyu/goutil/config"
	"sync"
)
//配置文件操作


type singleton goconfig.Viperable


var v singleton
var once sync.Once


func GetConfigInstance() singleton {
	once.Do(load)
	return v
}

//配置文件初始化
func load() {

	v = goconfig.New()
	v.SetConfig(Arg.configfile,"json","/etc","/root",".")
	err := v.ReadConfig()
	if err != nil {
		panic(err)
	}

	v.Getconfig()

}



type ConfResponse struct {
	Action string
	Key   string
	Value interface{}
	Error error
}


func ConfWatch(stop chan struct{})  <-chan *ConfResponse{
	respChan := make(chan *ConfResponse, 10)

	go func() {
		//inode
		watcher, err := fsnotify.NewWatcher()
		//监视配置文件inode 出错了,退出程序
		if err != nil {
			panic(err)
		}

		watcher.Add(Arg.configfile)

		go func() {
			<-stop
			watcher.Close()
		}()

		respdata := &ConfResponse{
			Error: nil,
		}

		for {
			select {
			case event := <-watcher.Events:
				//fmt.Println(event)
				if event.Op&fsnotify.Remove == fsnotify.Remove ||
					event.Op&fsnotify.Rename == fsnotify.Rename ||
					event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Create == fsnotify.Create {
					watcher.Remove(Arg.configfile)
					watcher.Add(Arg.configfile)

					//需要读取配置文件
					//通过chan通知
					respChan <- respdata
				}

			case err := <-watcher.Errors:
				respdata.Error = err
				respChan <- respdata
			}

		}
	}()


	return respChan
}

