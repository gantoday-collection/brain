package rankplay

import (
	"brain/game/room"
	"github.com/gorilla/websocket"
	"brain/game/socketconn"
)
var rankPlay RankPlay
func init() {
	rankPlay=&myRankPlay{}
}
type RankPlay interface {
	//对消息的操作
	MsgFunc(token string,room room.Room)
}
func RankPlayInstance() RankPlay {
	return rankPlay
}

type myRankPlay struct {

}
// 定义消息对象
type Message struct {
	Token    string `json:"token"`
	RoomId string`json:"roomId"`
}
//one发送消息的操作
func (this *myRankPlay)MsgFunc(token string,room room.Room){
	conn,_:= socketconn.WebSocketPollInstance().Take(token)
	msgFunc(token,room,conn)
}
//对请求数据处理
func msgFunc(token string,room room.Room,conn *websocket.Conn){
	for   {
		var msg Message
		// 将新消息读入JSON并将其映射到消息对象
		err := conn.ReadJSON(&msg)
		if err != nil {
			break
		}
	}
}