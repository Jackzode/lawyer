package config

import (
	"fmt"
	"github.com/lawyer/commons/handler"
	"github.com/spf13/viper"
	"path/filepath"
)

const (
	DefaultConfigFileName                  = "config.yaml"
	DefaultCacheFileName                   = "cache.db"
	DefaultReservedUsernamesConfigFileName = "reserved-usernames.json"
)

type ServiceConfig struct {
	UploadPath string `json:"upload_path" mapstructure:"upload_path" yaml:"upload_path"`
}

type I18n struct {
	BundleDir string `json:"bundle_dir" mapstructure:"bundle_dir" yaml:"bundle_dir"`
}

type HTTP struct {
	Addr string `json:"addr" mapstructure:"addr"`
}

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
	HTTP *HTTP `json:"http" mapstructure:"http" yaml:"http"`
}

// AllConfig all config
type AllConfig struct {
	Debug         bool               `json:"debug" mapstructure:"debug" yaml:"debug"`
	Server        *Server            `json:"server" mapstructure:"server" yaml:"server"`
	I18n          *I18n              `json:"i18n" mapstructure:"i18n" yaml:"i18n"`
	ServiceConfig *ServiceConfig     `json:"service_config" mapstructure:"service_config" yaml:"service_config"`
	Data          *handler.Database  `json:"data" mapstructure:"data" yaml:"data"`
	Cache         *handler.RedisConf `json:"redis" mapstructure:"redis" yaml:"redis"`
}
