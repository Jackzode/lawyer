package downstream

import (
	"context"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/spf13/viper"
	"log"
)

var (
	MysqlEngine *xorm.Engine
	RedisClient *redis.Client
)

func InitDownStream(configFile string) error {
	//load config
	viper.SetConfigFile(configFile)
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("read config error : ", err)
	}
	driver := viper.GetString("data.driver")
	dsn := viper.GetString("data.dsn")
	MysqlEngine, err = xorm.NewEngine(driver, dsn)
	MysqlEngine.ShowSQL(true)
	if err != nil {
		log.Println(err)
		return err
	}
	err = MysqlEngine.Ping()
	if err != nil {
		log.Println(err)
		return err
	}
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.addr"),
		Password: "", // 密码
		PoolSize: 20, // 连接池大小
	})
	pong := RedisClient.Ping(context.TODO())
	if pong.Err() != nil {
		return pong.Err()
	}
	return nil
}
