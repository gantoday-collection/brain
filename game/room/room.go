package room

import (
	"sync"
	"errors"
	"brain/game/socketconn"
	"github.com/gorilla/websocket"
	"io"
	"encoding/base64"
	"crypto/rand"
)
//Room接口
type Room interface {
	Id()string
	//添加房主
	SetOwner(token string)error
	//添加房客
	SetTenant(token string)error
	//发送消息给房间的所有人
	//返回的错误为左右两边token对应的Socket错误
	SendToAll(msg string)error
	//发送消息给另一个
	//token：当前用户的token
	SendMsgToOtherOne(token string,msg string)error
	//发送消息给token所对应的人
	//token：当前用户的token
	SendMsgToOne(token string,msg string)error
	//发送消息给参观者
	SendMsgWatchs(msg string)error
	//移出参观者
	RemoveWatchs(msg string)error
	//添加参观者
	AddWatchs(token string)error
	//房间关闭的信号
	CloseSignChan()chan bool
	//关闭房间
	closed()
} 

type room struct {
	id string
	houseOwner string
	tenant string
	close bool//是否已经关闭
	closeSign chan bool
	//给所有人发送消息的通道
	msgAll chan string
	//游戏者的连接map
	clients  map[string]*websocket.Conn
	//围观者
	watchs 	map[string]*websocket.Conn
	mutex sync.Mutex
}

func newRoom() (Room) {
	id:=GetGuid()//生成id
	msgAll:=make(chan string,5)//创建给所有人发送消息的通道
	closeSign:=make(chan bool,2)//创建给所有人发送消息的通道
	var clients = make(map[string]*websocket.Conn)//创建游戏者的连接map
	var watchs = make(map[string]*websocket.Conn)////围观者
	roo:=&room{id:id,close:false,closeSign:closeSign,msgAll:msgAll,clients:clients,watchs:watchs,mutex:*new(sync.Mutex)}
	go roo.sendToAll()
	return roo
}
//添加房主
func (this *room)SetOwner(token string)error{
	if this.close {
		return errors.New("房间已经关闭")
	}
	pool:=socketconn.WebSocketPollInstance()
	connn,ok:=pool.Take(token)
	if !ok {
		return errors.New(token+"用户所对应的用户连接不存在")
	}
	this.houseOwner=token
	this.clients[token]=connn
	return nil
}
//添加房客
func (this *room)SetTenant(token string)error{
	if this.close {
		return errors.New("房间已经关闭")
	}
	pool:=socketconn.WebSocketPollInstance()
	connn,ok:=pool.Take(token)
	if !ok {
		return errors.New(token+"用户所对应的用户连接不存在")
	}
	this.tenant=token
	this.clients[token]=connn
	return nil
}
//发送消息给房间的所有人
func (this *room)SendToAll(msg string)error{
	if this.close {
		return errors.New("房间已经关闭")
	}
	this.msgAll<-msg
	return nil
}
//发送消息给房间的所有人
func (this *room)sendToAll(){
	go func() {
		//var msg string
		//ok:=true
		for   {
			select{
			case msg,ok:=<-this.msgAll:
				if !ok {
					//End
					break
				}else{
					//continue
					this.mutex.Lock()
					defer this.mutex.Unlock()
					for _,v:=range this.clients{
						v.WriteJSON(msg)
					}
					for _,v:=range this.watchs{
						v.WriteJSON(msg)
					}
				}
			}
		}
	}()
}
//发送消息给参观者
func (this *room)SendMsgWatchs(msg string)error{
	if this.close {
		return errors.New("房间已经关闭")
	}
	this.mutex.Lock()
	defer this.mutex.Unlock()
	for _,v:=range this.watchs{
		v.WriteJSON(msg)
	}
	return nil
}
//发送消息给另一个
//token：当前用户的token
func (this *room)SendMsgToOtherOne(token string,msg string)error{
	if this.close {
		return errors.New("房间已经关闭")
	}
	if this.houseOwner==token {
		this.sendMsgToOne(this.tenant,msg)
		return nil
	}else if this.tenant==token {
		this.sendMsgToOne(this.houseOwner,msg)
		return nil
	}else {
		return errors.New("token:"+token+"不在这个房间内")
	}
}
//发送消息给token所对应的人
//token：当前用户的token
func (this *room)SendMsgToOne(token string,msg string)error{
	if this.close {
		return errors.New("房间已经关闭")
	}
	if this.houseOwner==token {
		return this.sendMsgToOne(this.houseOwner,msg)
	}else if this.tenant==token {
		return this.sendMsgToOne(this.tenant,msg)
	}else {
		return errors.New("token:"+token+"不在这个房间内")
	}
}
//token：当前用户的token
func (this *room)sendMsgToOne(token string,msg string)error{
	if this.close {
		return errors.New("房间已经关闭")
	}
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return this.clients[token].WriteJSON(msg)
}

//获取id
func (this *room)Id()string{
	return this.id
}
//获取guid
func GetGuid() string {
	b := make([]byte, 48)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}
//添加参观者
func (this *room)AddWatchs(token string)error{
	if this.close {
		return errors.New("房间已经关闭")
	}
	cnn,ok:=socketconn.WebSocketPollInstance().Take(token)
	if !ok {
		return errors.New("token为:"+token+"的连接不存在")
	}
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.watchs[token]=cnn
	return nil
}
//移出参观者
func (this *room)RemoveWatchs(token string)error{
	if this.close {
		return errors.New("房间已经关闭")
	}
	this.mutex.Lock()
	defer this.mutex.Unlock()
	delete(this.watchs, token)
	return nil
}
//关闭房间
func (this *room)closed(){
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.clients=nil
	this.watchs=nil
	this.close=true
	this.closeSign<-true
	this.closeSign<-true
	close(this.msgAll)//关闭给所有人发送消息的通道
	close(this.closeSign)
}
//关闭房间
func (this *room)CloseSignChan()chan bool{
	return this.closeSign
}