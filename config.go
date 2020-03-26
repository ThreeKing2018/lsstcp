package serverRoom

import (
	"github.com/fsnotify/fsnotify"
	"strings"
	"sync"

	//goconfig "gogs.163.com/feiyu/goutil/config"
	"github.com/spf13/viper"
)

//配置文件操作

var v *viper.Viper
var once sync.Once

func GetConfigInstance() *viper.Viper {
	once.Do(load)
	return v
}

//
//配置文件初始化
func load() {
	v = viper.New()

	v.SetConfigName(strings.Trim(Arg.configFile, ".json")) // name of config file (without extension)
	v.SetConfigType("json")                                // REQUIRED if the config file does not have the extension in the name
	v.AddConfigPath(".")
	err := v.ReadInConfig()
	if err != nil {
		panic(err)
	}

}

type ConfResponse struct {
	Action string
	Key    string
	Value  interface{}
	Error  error
}

func ConfWatch(stop chan struct{}) <-chan *ConfResponse {
	respChan := make(chan *ConfResponse, 10)

	go func() {
		//inode
		watcher, err := fsnotify.NewWatcher()
		//监视配置文件inode 出错了,退出程序
		if err != nil {
			panic(err)
		}

		watcher.Add(Arg.configFile)

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
					watcher.Remove(Arg.configFile)
					watcher.Add(Arg.configFile)

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
