package redisconf

import (
	"github.com/garyburd/redigo/redis"
	"time"
	"strconv"
	"brain/conf/config"
)
var pool *redis.Pool
func InitRedisPool()  {
	Maxidle,_:=strconv.Atoi(config.Conf.Maxidle)
	Maxactive,_:=strconv.Atoi(config.Conf.Maxactive)
	Idletimeout,_:=strconv.Atoi(config.Conf.Idletimeout)
	Redisserver:=config.Conf.Redisserver
	pool=initPoll(Redisserver,"",Idletimeout,Maxactive,Maxidle)
}
//获取redis的实例
func GetRedisInstance()redis.Conn {
	return pool.Get()
}
func initPoll(server, auth string,Idletimeout  int,Maxidle int,Maxactive int) *redis.Pool {
	return &redis.Pool{
		//最大空闲
		MaxIdle:Maxidle,
		//最大活动量
		MaxActive:  Maxactive,
		//空闲超时
		IdleTimeout: time.Duration(Idletimeout) * time.Millisecond,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if auth == "" {
				return c, err
			}
			if _, err := c.Do("AUTH", auth); err != nil {
				c.Close()
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}