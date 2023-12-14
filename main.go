package main

import (
	"lawyer/config"
	"lawyer/initServer"
	"lawyer/routes"
	"log"
)

func main() {
	//初始化服务下游、log、db
	err := initServer.InitServer()
	if err != nil {
		log.Fatal(err)
	}
	//init router
	engine := routes.InitRouter()
	//run server
	err = engine.Run(config.Host)
	if err != nil {
		log.Fatal(err)
	}

}
