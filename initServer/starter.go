package initServer

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lawyer/commons/config"
	"github.com/lawyer/commons/handler"
	"github.com/lawyer/repo"
	"github.com/lawyer/router"
	"github.com/lawyer/service"
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
	err = service.InitTranslator(c.I18n)
	checkErr(err)
	repo.InitRepo()
	service.InitServices()
	application, err := initApplication(c.Debug)
	checkErr(err)
	return application
}

func initApplication(debug bool) (*gin.Engine, error) {

	ginEngine := router.NewHTTPServer(debug)
	return ginEngine, nil

}
