package friendplay

import (
	"brain/game/room"
	"github.com/gorilla/websocket"
)

var friendPlay FriendPlay
func init() {
	friendPlay=&myFriendPlay{}
}

type FriendPlay interface {
	//房主接收和发送消息操作
	OwnerMsgFunc(token string,room room.Room,conn *websocket.Conn)
	//房客接收和发送消息操作
	TenantMsgFunc(token string,room room.Room,conn *websocket.Conn)
}
func FriendPlayInstance() FriendPlay {
	return friendPlay
}

type myFriendPlay struct {

}
// 定义消息对象
type Message struct {
	Token    string `json:"token"`
	RoomId string`json:"roomId"`
}
//房主接收和发送消息操作
func (this *myFriendPlay)OwnerMsgFunc(token string,room room.Room,conn *websocket.Conn)  {
	for   {
		var msg Message
		// 将新消息读入JSON并将其映射到消息对象
		err := conn.ReadJSON(&msg)
		if err != nil {
			break
		}
	}
}

//房客接收和发送消息操作
func (this *myFriendPlay)TenantMsgFunc(token string,room room.Room,conn *websocket.Conn){

}

