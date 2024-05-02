package handler

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"time"
	"xorm.io/xorm"
	"xorm.io/xorm/log"
	"xorm.io/xorm/names"
)

var (
	Engine      *xorm.Engine
	RedisClient *redis.Client
)

// Database database config
type Database struct {
	Driver          string `json:"driver" mapstructure:"driver" yaml:"driver"`
	Connection      string `json:"connection" mapstructure:"connection" yaml:"connection"`
	ConnMaxLifeTime int    `json:"conn_max_life_time" mapstructure:"conn_max_life_time" yaml:"conn_max_life_time,omitempty"`
	MaxOpenConn     int    `json:"max_open_conn" mapstructure:"max_open_conn" yaml:"max_open_conn,omitempty"`
	MaxIdleConn     int    `json:"max_idle_conn" mapstructure:"max_idle_conn" yaml:"max_idle_conn,omitempty"`
}

// RedisConf cache
type RedisConf struct {
	Addr       string `json:"addr" yaml:"addr" mapstructure:"addr"`
	MaxOpen    int    `json:"maxOpen" yaml:"maxOpen" mapstructure:"maxOpen"`
	MaxIdle    int    `json:"maxIdle" yaml:"maxIdle" mapstructure:"maxIdle"`
	MaxConnect int    `json:"maxConnect" yaml:"maxConnect" mapstructure:"maxConnect"`
	Timeout    int    `json:"timeout" yaml:"timeout" mapstructure:"timeout"`
	PoolSize   int    `json:"poolSize" yaml:"poolSize" mapstructure:"poolSize"`
	Auth       string `json:"auth" yaml:"auth" mapstructure:"auth"`
}

func InitDBandCacheHandler(debug bool, dbConf *Database, cacheConf *RedisConf) (err error) {
	Engine, err = NewDB(debug, dbConf)
	if err != nil {
		return err
	}
	RedisClient, err = NewRedisCache(cacheConf)
	if err != nil {
		return err
	}
	return nil
}

func NewRedisCache(conf *RedisConf) (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Auth,     // 密码
		PoolSize: conf.PoolSize, // 连接池大小
	})
	err := redisClient.Ping(context.TODO()).Err()
	if err != nil {
		return nil, err
	}
	return redisClient, nil
}

// NewDB new database instance
func NewDB(debug bool, dataConf *Database) (*xorm.Engine, error) {

	engine, err := xorm.NewEngine(dataConf.Driver, dataConf.Connection)
	if err != nil {
		return nil, err
	}

	if debug {
		engine.ShowSQL(true)
	} else {
		engine.SetLogLevel(log.LOG_ERR)
	}

	if err = engine.Ping(); err != nil {
		return nil, err
	}

	if dataConf.MaxIdleConn > 0 {
		engine.SetMaxIdleConns(dataConf.MaxIdleConn)
	}
	if dataConf.MaxOpenConn > 0 {
		engine.SetMaxOpenConns(dataConf.MaxOpenConn)
	}
	if dataConf.ConnMaxLifeTime > 0 {
		engine.SetConnMaxLifetime(time.Duration(dataConf.ConnMaxLifeTime) * time.Second)
	}
	engine.SetColumnMapper(names.GonicMapper{})
	return engine, nil
}
