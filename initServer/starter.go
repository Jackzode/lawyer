package initServer

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/config"
	"github.com/lawyer/commons/handler"
	"github.com/lawyer/initServer/initRepo"
	"github.com/lawyer/initServer/initServices"
)

func checkErr(err error) {
	if err != nil {
		fmt.Printf(err.Error())
		panic(err)
	}
}

func Init(filename string) *gin.Engine {
	// init conf file "conf/config.yaml"
	c, err := config.ReadConfig(filename)
	checkErr(err)
	//init db
	err = handler.InitDBandCacheHandler(c.Debug, c.Data, c.Cache)
	checkErr(err)
	// init i18n
	err = initTranslator(c.I18n)
	checkErr(err)
	repo.InitRepo()
	services.InitServices()
	application, err := initApplication(c.Debug, c.Server, c.ServiceConfig)
	checkErr(err)
	return application
}
