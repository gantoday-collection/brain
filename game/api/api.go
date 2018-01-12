package api

import (
	"net/http"
	"github.com/gorilla/websocket"
)

// 定义消息对象
type Message struct {
	Token    string `json:"token"`
	RoomId string`json:"roomId"`
}
//初始化网络路由
func InIt(){
	// 创建一个简单的静态文件路由
	fs := http.FileServer(http.Dir("./views"))
	http.Handle("/static", fs)
	// 创建一个朋友对决的房间
	http.HandleFunc("/game/friendplay", friendPlay)
	// 加入一个朋友对决的房间
	http.HandleFunc("/game/joinfriendplay", joinfriendplay)
	//进入排位赛
	http.HandleFunc("/game/rankplay", rankPlay)
}
// configure的升级程序
/**

CheckOrigin returns true if the request Origin header is acceptable. If
CheckOrigin is nil, the host in the Origin header must not be set or
must match the host of the request.
checkorigin返回true，如果请求源头是可以接受的。如果

checkorigin是零，在源头的主机不能设置或

必须与请求主机匹配。
 */
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}