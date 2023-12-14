package dao

import (
	"context"
	"fmt"
	"lawyer/dao/downstream"
	"lawyer/types"
)

func AddQuestion(ctx context.Context, question *types.Question) int64 {
	insert, err := downstream.MysqlEngine.Insert(question)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	return insert
}
