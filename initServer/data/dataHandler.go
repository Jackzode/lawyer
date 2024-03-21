package data

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
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

/*
// NewCache new cache instance
func NewCache(c *RedisConf) (cache.Cache, func(), error) {
	var pluginCache plugin.Cache
	_ = plugin.CallCache(func(fn plugin.Cache) error {
		pluginCache = fn
		return nil
	})
	if pluginCache != nil {
		return pluginCache, func() {}, nil
	}

	// TODO What cache type should be initialized according to the configuration file
	memCache := memory.NewCache()

	if len(c.FilePath) > 0 {
		cacheFileDir := filepath.Dir(c.FilePath)
		log.Debugf("try to create cache directory %s", cacheFileDir)
		err := dir.CreateDirIfNotExist(cacheFileDir)
		if err != nil {
			log.Errorf("create cache dir failed: %s", err)
		}
		log.Infof("try to load cache file from %s", c.FilePath)
		if err := memory.Load(memCache, c.FilePath); err != nil {
			log.Warn(err)
		}
		go func() {
			ticker := time.Tick(time.Minute)
			for range ticker {
				if err := memory.Save(memCache, c.FilePath); err != nil {
					log.Warn(err)
				}
			}
		}()
	}
	cleanup := func() {
		log.Infof("try to save cache file to %s", c.FilePath)
		if err := memory.Save(memCache, c.FilePath); err != nil {
			log.Warn(err)
		}
	}
	return memCache, cleanup, nil
}
*/

//// NewData new data instance
//func NewData(db *xorm.Engine, cache *redis.Client) (*Data, func(), error) {
//	cleanup := func() {
//		log.Info("closing the data resources")
//		db.Close()
//	}
//	return &Data{DB: db, Cache: cache}, cleanup, nil
//}

// Data data
//type Data struct {
//	DB    *xorm.Engine
//	Cache *redis.Client
//}
