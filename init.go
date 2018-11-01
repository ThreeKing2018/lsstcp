package serverRoom

import (
	"flag"
	"github.com/ThreeKing2018/goutil/golog"
	"github.com/ThreeKing2018/goutil/golog/conf"
	"path"
)



func Init() {
	flag.Parse()

	//打印版本并退出
	if Arg.Getver() {
		printVersion()
	}

	golog.SetLogger(
		golog.ZAPLOG,
		conf.WithLogType(conf.LogJsontype),
		conf.WithLogLevel(conf.DebugLevel),
		conf.WithFilename(path.Join(Arg.logdir,ServiceName+".log")),
	)
}

