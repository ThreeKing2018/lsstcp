package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	_,err:=net.DialTimeout("tcp","127.0.0.1:6378",1*time.Second)
	if err != nil {
		fmt.Println(err)
	}
	//
	//_,err = conn.Write([]byte("a"))
	//if err != nil {
	//	fmt.Println(err)
	//}
	fmt.Println("ok")
}
