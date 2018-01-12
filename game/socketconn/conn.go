package socketconn

import (
	"github.com/gorilla/websocket"
	"errors"
	"sync"
)
//所有用户token对应的websocket连接
var clients = make(map[string]*websocket.Conn)
//全局的websocketpoll
var websocketpoll WebSocketPoll
//在加载前就做的事情
//初始化websocketpoll
func init() {
	websocketpoll=&myWebSocketPoll{mutex:*new(sync.Mutex)}
}

type WebSocketPoll interface {
	//根据token关闭用户的连接
	Close(token string)error
	//保存一个用户连接
	Put(token string,cnn *websocket.Conn)
	//取出连接
	Take(tiken string)(*websocket.Conn,bool)
	////发送信息
	//SendMsg(token string,msg string)error
}

type myWebSocketPoll struct {
	mutex sync.Mutex
}

func WebSocketPollInstance()  WebSocketPoll{
	return websocketpoll
}
//根据token关闭用户的连接
func (poll *myWebSocketPoll)Close(token string)error{
	cnn,ok:=clients[token]
	if !ok {
		return errors.New("没有这个连接，这个连接可能已经关闭")
	}
	poll.mutex.Lock()
	defer poll.mutex.Unlock()
	cnn.Close()
	delete(clients, token)
	return nil
}
//保存一个用户连接
func (poll *myWebSocketPoll)Put(token string,cnn *websocket.Conn){
	poll.mutex.Lock()
	defer poll.mutex.Unlock()
	clients[token]=cnn
}
//发送信息
func (poll *myWebSocketPoll)SendMsg(token string,msg string)error{
	_,ok:=clients[token]
	if ok {
		poll.mutex.Lock()
		defer poll.mutex.Unlock()
		return clients[token].WriteJSON(msg)
	}else {
		return errors.New("池中没有该用户对应的连接")
	}
}
//取出连接
func (poll *myWebSocketPoll)Take(token string)(*websocket.Conn,bool){
	cnn,ok:=clients[token]
	return cnn,ok
}


