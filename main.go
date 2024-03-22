package main

import (
	"fmt"
	"github.com/lawyer/commons/logger"
	"github.com/lawyer/initServer"
)

func main() {
	// todo init log
	glog.InitLog("stdout")
	defer func() {
		err := glog.Logger.Sync()
		fmt.Println(err)
	}()
	filename := `.\conf\config.yaml`
	application := initServer.Init(filename)
	err := application.Run()
	if err != nil {
		panic(err)
	}
}
