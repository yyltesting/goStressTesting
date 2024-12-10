package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

type TokenCacheInfo struct {
	Userid     string `redis:"userid" json:"userid"`
	LoginIp    string `redis:"login_ip" json:"login_ip"`
	Ua         string `redis:"ua" json:"ua"`
	Token      string `redis:"token" json:"token"`
	LoginTime  int64  `redis:"login_time" json:"login_time"`
	IsLogin    bool   `redis:"is_login" json:"is_login"`
	DeviceName string `redis:"device_name" json:"device_name"` // 登录设备名称
	DevicePlat string `redis:"device_plat" json:"device_plat"` // 设备平台 ios 安卓 win...
}

func Creat_csvtoken_redis(csvlist [][]string) error {
	// 创建 Redis 集群客户端
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{"10.0.115.223:6379", "10.0.115.223:6379"},
		Password: "xap4eh0OLa3eequ4Geiteing4Ozootha",
	})
	// 上下文
	ctx := context.Background()
	// 测试连接
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		fmt.Println("无法连接到 Redis 集群:", err)
		return nil
	}
	fmt.Println("成功连接到 Redis 集群")

	for i,x:= range csvlist{
		userid := x
		key1 := "HABOX_TOKEN_INFO:"+userid[0]
		tc := TokenCacheInfo{
			Userid:     userid[0],
			LoginIp:    "125.70.78.82",
			Ua:         "",
			Token:      userid[0],
			LoginTime:  time.Now().Unix(),
			IsLogin:    true,
			DeviceName: "iphone 11",
			DevicePlat: "web.pc",
		}
		err := rdb.HSet(ctx,key1,tc.Userid,tc.DeviceName,tc.DevicePlat,tc.IsLogin,tc.Token,tc.Ua,tc.LoginTime,tc.LoginIp).Err()
		if err != nil {
			fmt.Println(err)
			return err
		}
		rdb.Expire(ctx,key1, -1)
		if i==501{
			break
		}
		time.Sleep(time.Millisecond*10)

	}
	//fmt.Println("1完成")

	for _,x:= range csvlist{
		userid := x
		key := "HABOX_TOKEN_INFO:"+userid[0]
		tc := TokenCacheInfo{
			Userid:     userid[0],
			LoginIp:    "125.70.78.82",
			Ua:         "",
			Token:      userid[0],
			LoginTime:  time.Now().Unix(),
			IsLogin:    true,
			DeviceName: "iphone 11",
			DevicePlat: "web.pc",
		}
		err := rdb.HSet(ctx,key,tc.Userid,tc.DeviceName,tc.DevicePlat,tc.IsLogin,tc.Token,tc.Ua,tc.LoginTime,tc.LoginIp).Err()
		if err != nil {
			fmt.Println(err)
			return err
		}
		rdb.Expire(ctx,key, -1)
		fmt.Println(userid[0])
		time.Sleep(time.Millisecond*10)
	}
	fmt.Println("生成redis成功")
	return nil
}



