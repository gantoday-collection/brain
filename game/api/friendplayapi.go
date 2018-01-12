package api

import (
	"brain/utils"
	"brain/game/room"
	"brain/game/socketconn"
	"brain/game/friendplay"
	"net/http"
	"log"
)

//创建一个朋友挑战的房间
func friendPlay(w http.ResponseWriter, r *http.Request) {
	// 最初的GET请求一个WebSocket升级
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	var msg Message
	// 将新消息读入JSON并将其映射到消息对象
	// 确保函数返回时关闭连接。
	err = ws.ReadJSON(&msg)
	if err != nil {
		defer ws.Close()
		var result utils.Result
		result.ResultCode="0000"
		result.ResultMsg="接口调用成功"
		result.ResultSubCode="0001"
		result.ResultSubMsg="参数不能为空"
		ws.WriteJSON(result)
		return
	}
	//将连接保存到连接池里面
	socketconn.WebSocketPollInstance().Put(msg.Token,ws)
	//获取一个房间
	friendplay.FriendPlayInstance()
	roomm:=room.RoomPoolInstance().TakeNewRoom()
	roomm.SetOwner(msg.Token)//设置房主
	var result utils.Result
	result.ResultCode="0000"
	result.ResultMsg="接口调用成功"
	result.ResultSubCode="0000"
	subMsg:=make(map[string]string)
	subMsg["roomId"]=roomm.Id()
	result.ResultSubMsg=subMsg
	ws.WriteJSON(result)
	go 	friendplay.FriendPlayInstance().OwnerMsgFunc(msg.Token,roomm,ws)
	sign:=roomm.CloseSignChan()
	//获得关闭信号就关闭连接
	<-sign
}
//利用房间id加入一个房间
func joinfriendplay(w http.ResponseWriter, r *http.Request) {
	// 最初的GET请求一个WebSocket升级
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	var msg Message
	// 将新消息读入JSON并将其映射到消息对象
	// 确保函数返回时关闭连接。
	err = ws.ReadJSON(&msg)
	if err != nil {
		defer ws.Close()
		var result utils.Result
		result.ResultCode="0000"
		result.ResultMsg="接口调用成功"
		result.ResultSubCode="0001"
		result.ResultSubMsg="参数不能为空"
		ws.WriteJSON(result)
		return
	}
	roomm,ok:=room.RoomPoolInstance().Take(msg.RoomId)
	if !ok {
		defer ws.Close()
		var result utils.Result
		result.ResultCode="0000"
		result.ResultMsg="接口调用成功"
		result.ResultSubCode="0005"
		result.ResultSubMsg="房间号为：‘"+msg.RoomId+"’的房间不存在"
		ws.WriteJSON(result)
		return
	}
	//将连接保存到连接池里面
	socketconn.WebSocketPollInstance().Put(msg.Token,ws)
	roomm.SetTenant(msg.Token)//设置房客
	var result utils.Result
	result.ResultCode="0000"
	result.ResultMsg="接口调用成功"
	result.ResultSubCode="0000"
	subMsg:=make(map[string]string)
	subMsg["roomId"]=roomm.Id()
	result.ResultSubMsg=subMsg
	ws.WriteJSON(result)
	go friendplay.FriendPlayInstance().TenantMsgFunc(msg.Token,roomm,ws)
	sign:=roomm.CloseSignChan()
	//获得关闭信号就关闭连接
	<-sign
}
