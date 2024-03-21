package initServer

import (
	"fmt"
	"github.com/apache/incubator-answer/initServer/data"
	"github.com/apache/incubator-answer/internal/base/cron"
	"github.com/apache/incubator-answer/internal/base/server"
	"github.com/apache/incubator-answer/internal/base/translator"
	sc "github.com/apache/incubator-answer/internal/service/service_config"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman"
	"github.com/segmentfault/pacman/contrib/server/http"
	"github.com/spf13/viper"
	"path/filepath"
)

func checkErr(err error) {
	if err != nil {
		fmt.Printf(err.Error())
		panic(err)
	}
}

func Start(filename string) *pacman.Application {
	//读取配置文件
	c, err := ReadConfig(filename) // /conf/config.yaml
	checkErr(err)
	err = data.InitDBandCacheHandler(c.Debug, c.Data, c.Cache)
	checkErr(err)
	err = initTranslator(c.I18n)
	checkErr(err)
	initRepo()
	initModel(c.ServiceConfig)
	application, err := initApplication(c.Debug, c.Server, c.ServiceConfig)
	checkErr(err)
	return application
}

// go build -ldflags "-X github.com/apache/incubator-answer/cmd.Version=x.y.z"
var (
	// Name is the name of the project
	Name = "answer"
	// Version is the version of the project
	Version = "0.0.0"
	// Revision is the git short commit revision number
	// If built without a Git repository, this field will be empty.
	Revision = ""
	// Time is the build time of the project
	Time = ""
	// GoVersion is the go version of the project
	GoVersion = "1.19"
)

func newApplication(serverConf *Server, server *gin.Engine, manager *cron.ScheduledTaskManager) *pacman.Application {
	manager.Run()
	return pacman.NewApp(
		pacman.WithName(Name),
		pacman.WithVersion(Version),
		pacman.WithServer(http.NewServer(server, serverConf.HTTP.Addr)),
	)
}

const (
	DefaultConfigFileName                  = "config.yaml"
	DefaultCacheFileName                   = "cache.db"
	DefaultReservedUsernamesConfigFileName = "reserved-usernames.json"
)

var (
	ConfigFileDir  = "./conf/"
	UploadFilePath = "/uploads/"
	I18nPath       = "/i18n/"
	CacheDir       = "/cache/"
)

func GetConfigFilePath() string {
	return filepath.Join(ConfigFileDir, DefaultConfigFileName)
}

func ReadConfig(configFilePath string) (c *AllConfig, err error) {
	fmt.Println(configFilePath)
	c = &AllConfig{}
	v := viper.New()
	v.SetConfigFile(configFilePath)
	err = v.ReadInConfig()
	if err != nil {
		return nil, err
	}
	if err = v.Unmarshal(&c); err != nil {
		return nil, err
	}
	fmt.Println(*c.Data, "  ||  ", *c.Cache)
	return c, nil
}

// Server server config
type Server struct {
	HTTP *server.HTTP `json:"http" mapstructure:"http" yaml:"http"`
}

// AllConfig all config
type AllConfig struct {
	Debug         bool              `json:"debug" mapstructure:"debug" yaml:"debug"`
	Server        *Server           `json:"server" mapstructure:"server" yaml:"server"`
	I18n          *translator.I18n  `json:"i18n" mapstructure:"i18n" yaml:"i18n"`
	ServiceConfig *sc.ServiceConfig `json:"service_config" mapstructure:"service_config" yaml:"service_config"`
	Data          *data.Database    `json:"data" mapstructure:"data" yaml:"data"`
	Cache         *data.RedisConf   `json:"redis" mapstructure:"redis" yaml:"redis"`
}
