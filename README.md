昨天晚上在玩头脑小游戏的时候突然发生了服务器崩了的事情，就对这个游戏进行了一系列的思考。

这个游戏最主要的模块就是好友对战和排位战：

而这两个模块最基本的都可以类似于一个一个聊天室，因此我模仿这个操作写了一个系统。

这个系统采用的是WebSocket若对WebSocket不清楚的话就可以参考[使用WebSocket创建即时聊天室](https://www.jianshu.com/p/5d000523e2bd)

### 项目结构

	|---brain
		|---main
			|---main.go
		|---gain
			|---api
				|---api.go
				|---friendplayapi.go
				|---rankplayapi.go
			|---friendplay
				|---friendplay.go
			|---rankplay
				|---rankplay.go
			|---room
				|---room.go
				|---roompool.go
			|---socketconn
				|---conn.go
		|---vendor

### 代码

#### main.go

	package main
	
	import (
		"brain/game/api"
	)
	
	func main() {
		api.InIt()
	}

#### api.go

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

#### friendplayapi.go

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


#### rankplayapi.go

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

#### friendplay.go

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



#### rankplay.go

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

#### room.go

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

#### roompool.go

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


#### conn.go

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



