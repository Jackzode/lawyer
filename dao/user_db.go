package dao

import (
	"context"
	"fmt"
	"lawyer/common"
	"lawyer/dao/downstream"
	"lawyer/types"
)

// FindUserByOneCondition
/*
数据库里的数据封装到bean里，返回值代表是否get到数据
*/
func FindUserByOneCondition(bean interface{}, condition, value string) (bool, error) {

	get, err := downstream.MysqlEngine.Where(condition+"=?", value).Get(bean)
	return get, err
}

func SaveUserEmail(userInfo *types.UserInfo) error {
	insert, err := downstream.MysqlEngine.Insert(userInfo)
	if err != nil || insert != 1 {
		fmt.Println("downstream.MysqlEngine.Insert(userInfo)---", err, "==", insert)
	}
	return err
}

func GetUserCaptcha(ctx context.Context, key string) string {
	get := downstream.RedisClient.Get(ctx, key)
	return get.Val()
}

func SaveUserToken(ctx context.Context, uid, token string) error {
	setEX := downstream.RedisClient.SetEX(ctx, uid, token, common.TokenExpiration)
	if setEX.Err() != nil {
		fmt.Println("set token err ", setEX.Err().Error())
	}
	return setEX.Err()
}

func DeleteUserToken(ctx context.Context, key string) {
	cmd := downstream.RedisClient.Del(ctx, key)
	if cmd.Err() != nil {
		fmt.Println("delete token....err...", cmd.Err())
	}
	return
}
