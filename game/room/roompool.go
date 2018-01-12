package room

import (
	"sync"
	"errors"
)
var roomPool RoomPool
func init() {
	roomMap:=make(map[string]Room)
	roomPool=&myRoomPool{roomMap:roomMap,mutex:*new(sync.Mutex)}
}
type RoomPool interface {
	//取出实体
	Take(id string)(Room,bool)
	//取出新的实体
	TakeNewRoom()(Room)
	//删除
	Close(id string)error
}
//实体池的实现类型
type myRoomPool struct {
	roomMap map[string] Room
	mutex sync.Mutex
}

func RoomPoolInstance() RoomPool {
	return roomPool
}
//取出实体
func (this *myRoomPool)Take(id string)(Room,bool){
	this.mutex.Lock()
	defer this.mutex.Unlock()
	room,ok:=this.roomMap[id]
	return  room,ok
}
//取出实体
func (this *myRoomPool)TakeNewRoom()(Room){
	this.mutex.Lock()
	defer this.mutex.Unlock()
	room:=newRoom()
	this.roomMap[room.Id()]=room
	return  room
}
//关闭房间
//id房间的id
func (this *myRoomPool)Close(id string)error{
	room,ok:=this.roomMap[id]
	if !ok {
		return errors.New("没有这个房间，这个房间可能已经关闭")
	}
	this.mutex.Lock()
	defer this.mutex.Unlock()
	room.closed()//关闭房间
	delete(this.roomMap,id )
	return nil
}
