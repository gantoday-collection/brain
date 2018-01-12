package db

import (
	"time"
	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
    _"github.com/go-sql-driver/mysql"
	"reflect"
	"fmt"
	"errors"
	"strconv"
	"brain/conf/config"
)

type dB interface {
	Gorm()*gorm.DB
	Id()uint32//ID的获取方法
}

type myDB struct {
	db *gorm.DB
	id uint32//ID
}

type DBPoll interface {
	Take()(dB,error)//取出实体
	Return(entity dB)(error)//归还实体
	Total()uint32//实体的容量
	Used()uint32//实体中已经被使用的实体数量
}
type myDBPoll struct {
	pool  DBPoll //实体池
	etype reflect.Type    //池内实体的类型
}
//生成DB的函数类型
type genDB func() dB
func newDBPoll(total uint32,gen genDB)(DBPoll,error)  {
	etype:=reflect.TypeOf(gen())
	genEntity:= func() dB{return gen()}
	pool,err:= NewPool(total,etype,genEntity)
	if err!=nil {
		return nil,err
	}
	dbpool:=&myDBPoll{pool,etype}
	return dbpool,nil
}

func (db *myDB)Id()uint32  {
	return db.id
}
func (db *myDB)Gorm()*gorm.DB {
	return db.db
}
//取出实体
func (pool *myDBPoll)Take()(dB,error){
	entity,err:=pool.pool.Take()
	if err!=nil {
		return nil,err
	}
	dl,ok:=entity.(dB)//强制类型转换
	if !ok {
		errMsg:=fmt.Sprintf("The type of entity id NOT %s\n",pool.etype)
		panic(errors.New(errMsg))
	}
	return dl,nil
}
//归还实体
func (pool *myDBPoll)Return(entity dB)(error){
	return pool.pool.Return(entity)
}
//实体的容量
func (pool *myDBPoll)Total()uint32{
	return pool.pool.Total()
}
//实体中已经被使用的实体数量
func (pool *myDBPoll)Used()uint32{
	return pool.pool.Used()
}

var dbPoll DBPoll

func InitDB() {

	total := config.Conf.Total
	to,_:=strconv.Atoi(total)
	dbPoll,_=newDBPoll(uint32(to),initDb)
}
//func GetDBPollInstance() DBPoll {
//	return dbPoll
//}
func GetDBInstance() (dB,error) {
	db,err:=dbPoll.Take()
	if err!=nil {
		return nil, err
	}
	return db,nil
}
func ReturnDB(db dB) error {
	return dbPoll.Return(db)
}
func initDb()  dB{
	var db *gorm.DB
	var err error
	path := config.Conf.DBURL               //从env获取数据库连接地址
	logrus.Info("path:", string(path)) //打印数据库连接地址
	for {
		db, err = gorm.Open("mysql", string(path)) //使用gorm连接数据库
		if err != nil {
			logrus.Error(err, "Retry in 2 seconds!")
			time.Sleep(time.Second * 2)
			continue
		}
		logrus.Info("DB connect successful!")
		break
	}
	return &myDB{db:db,id:idGenertor.GetUint32()}
}
var idGenertor IdGenertor = NewIdGenertor()
