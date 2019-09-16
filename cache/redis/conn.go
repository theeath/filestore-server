package redis

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"time"
)
//config set requirepass test123 临时设置Redis密码
var(
	pool *redis.Pool
	redisHost = "127.0.0.1:6379"
	redisPass = "test1234"
)
//创建连接池
func newRedisPool()*redis.Pool  {
	return &redis.Pool{


		MaxIdle:         50,
		MaxActive:       30,
		IdleTimeout:     300 * time.Second,
		Dial: func() ( redis.Conn,  error) {
			//打开连接
			c,err := redis.Dial("tcp",redisHost)
			if err != nil {
				fmt.Println(err)
				return nil,err
			}
			//权限认证
			if _, err := c.Do("AUTH",redisPass);err !=nil{
				c.Close()
				return nil,err
			}
			return c,nil

		},
		//检测Redis的可用性
		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			// func Since(t Time) Duration 两个时间点的间隔
			if time.Since(t) < time.Minute{
				return nil
			}
			_, err := conn.Do("PING")
			return err
		},
	}
}
func init()  {
	pool = newRedisPool()
}
func RedisPool()*redis.Pool  {
	return pool
}