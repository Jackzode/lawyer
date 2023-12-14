package initServer

import (
	"github.com/spf13/viper"
	"lawyer/config"
	"lawyer/dao/downstream"
)

var v *viper.Viper

func InitServer() (err error) {
	v = viper.New()
	//load config
	v.SetConfigFile("./config/server.toml")
	err = v.ReadInConfig()
	if err != nil {
		panic(err)
	}
	err = InitLog()
	InitConfig()
	err = downstream.InitDownStream("./config/db_config.toml")
	return err
}

func InitConfig() {
	config.SmtpUsername = v.GetString("email.smtpUsername")
	config.SmtpPassWord = v.GetString("email.smtpPassword")
	config.SmtpPort = v.GetInt("email.smtpPort")
	config.SMTPHost = v.GetString("email.smtpHost")
	config.Host = v.GetString("server.addr")
	config.Mode = v.GetString("server.mode")
	return
}

// InitLog todo
func InitLog() error {

	return nil
}
