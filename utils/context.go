package utils

import (
	"context"
	"strconv"
)

const CURRENTUSERID = "CURRENTUSERID"

//将用户ID保存在context中
func SaveUserInfoToContext(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, CURRENTUSERID, value)
}

func GetUserInfoFromContext(ctx context.Context) uint64 {
	value := ctx.Value(CURRENTUSERID).(string)
	u, _ := strconv.ParseUint(value, 10, 32)
	return uint64(u)
}
