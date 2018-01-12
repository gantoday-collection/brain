package config

import (
	"github.com/Sirupsen/logrus"
	"github.com/koding/multiconfig"
	"github.com/joho/godotenv"
)

// Config 配置结构体
type Config struct {
	DB    string `default:"mysql"`
	DBURL string `default:"localhost"`
	Debug bool `default:"true"`
	Total string `default:"50"`
	Maxidle string `default:"100"`
	Maxactive string `default:"30"`
	Idletimeout string `default:"5"`
	Redisserver string `default:"127.0.0.1:6379"`
}

var Conf Config

// initConfig 初始化config
func InitConfig() {
	//获取.env文件
	err := godotenv.Overload("conf/config/.env")
	if err != nil {
		logrus.Infof("Error loading .env file: %+v", err)
	}
	//读取config default值
	m := multiconfig.MultiLoader(
		&multiconfig.TagLoader{},
		&multiconfig.EnvironmentLoader{
			CamelCase: true,
		},
	)
	//载入全局变量config
	m.Load(&Conf)
	logrus.Info(Conf)
	//设置日志等级
	if Conf.Debug {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "06-01-02 15:04:05.00",
		})
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "06-01-02 15:04:05.00",
		})
		logrus.SetLevel(logrus.InfoLevel)
	}
}
