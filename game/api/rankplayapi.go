package api

import (
	"net/http"
	"brain/utils"
	"brain/game/socketconn"
	"log"
	"brain/game/room"
	"brain/game/rankplay"
)
var newplear chan string
var roomIdSign map[string] chan string
func init() {
	newplear=make(chan string,10)
	roomIdSign=make(map[string] chan string)
	go newPlear()
}
//排位请求处理
func rankPlay(w http.ResponseWriter, r *http.Request){
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
	idchan,ok:=roomIdSign[msg.Token]//如果以前有连接的话就关闭以前的连接
	if ok {
		//如果右链接就关闭保存的idchan
		close(idchan)
	}
	sig:=make(chan string,1)
	roomIdSign[msg.Token]=sig
	roomId:=<-sig
	//关闭通道
	close(sig)
	roome,_:=room.RoomPoolInstance().Take(roomId)
	sign:=roome.CloseSignChan()
	//获得关闭信号就关闭连接
	<-sign
	//从roomIdSign中删除
	delete(roomIdSign, msg.Token)
}
//新人对决的函数
func newPlear()  {
	go func() {
		for   {
			 one,ok:=<-newplear
			 if !ok {
				 //End
				 break
			 }
			two,ok:=<-newplear
			if !ok {
				//End
				break
			}
			roomm:=room.RoomPoolInstance().TakeNewRoom()
			roomm.SetOwner(one)
			roomm.SetTenant(two)
			go rankplay.RankPlayInstance().MsgFunc(one,roomm)
			go rankplay.RankPlayInstance().MsgFunc(two,roomm)
			go func() {
				id:=roomm.Id()
				roomIdSign[one]<-id
				roomIdSign[two]<-id
			}()
		}
	}()
}