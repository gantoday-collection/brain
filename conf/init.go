package conf

import (
	"brain/conf/config"
	"brain/conf/db"
	"brain/conf/redisconf"
	"brain/conf/router"
)

func InitConf()  {
	//初始化.env文件配置
	config.InitConfig()
	//初始化数据库
	db.InitDB()
	//初始化redis
	redisconf.InitRedisPool()
	//初始化路由
	router.InitRouter()

}