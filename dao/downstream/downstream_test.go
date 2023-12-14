package downstream

import (
	"fmt"
	"lawyer/types"
	"testing"
)

func TestInitDownStream(t *testing.T) {

	err := InitDownStream("../../config/db_config.toml")
	//
	err = MysqlEngine.CreateTables(&types.Question{})
	err = MysqlEngine.Ping()
	fmt.Println(err)
	fmt.Println(err)
	//user := &types.UserInfo{
	//	EMail:    "867838901@qq.com",
	//	Username: "test",
	//	PassWord: utils.BcryptHash("12345"),
	//}
	//insert, err := MysqlEngine.Insert(user)
	//fmt.Println("insert...", insert)
	//if err != nil {
	//	fmt.Println("insert...", insert)
	//	return
	//}
	//fmt.Println(err)

	// 创建一个日志记录器
	//l, _ := zap.NewProduction()

	// 将日志写入文件
	//logger, _ = l.WriteToFile("app.log")

	// 记录日志
	//logger.Info("This is an info message")
}
