package main

import (
	"context"
	"fmt"
	"github.com/apache/incubator-answer/commons/logger"
	"github.com/apache/incubator-answer/initServer"
)

func main() {

	glog.InitLog("stdout")
	defer func() {
		err := glog.Logger.Sync()
		fmt.Println(err)
	}()
	//启动服务
	filename := `C:\jackzhi\go_project\incubator-answer\conf\config.yaml`
	application := initServer.Start(filename)
	err := application.Run(context.TODO())
	if err != nil {
		panic(err)
	}
}
