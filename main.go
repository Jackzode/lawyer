package main

import (
	"fmt"
	"github.com/lawyer/commons/logger"
	"github.com/lawyer/initServer"
)

func main() {
	// todo init log
	glog.InitLogger("./log/test.log", "debug")
	defer func() {
		err := glog.Klog.Sync()
		fmt.Println(err)
	}()
	fmt.Println("server is starting...")
	filename := `.\conf\config.yaml`
	application := initServer.Init(filename)
	err := application.Run(":8081")
	if err != nil {
		panic(err)
	}
}
