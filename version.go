package serverRoom

import (
	"fmt"
	"os"
)

const (
	VERSION = "0.1"
	ServiceName = "lsstcp"
)


func printVersion() {
	fmt.Sprintln("%s version: %s", ServiceName,VERSION)
	os.Exit(0)
}

