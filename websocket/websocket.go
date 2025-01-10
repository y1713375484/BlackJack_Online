package websocket

import (
	"balckJack/game"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"time"
)

type WebSocket struct{}

// 房间，默认房间内第一个用户为庄
var Room = map[int]*websocket.Conn{}
var RoomLock sync.RWMutex
var RoomSession = []string{}
var RoomSessionLock sync.RWMutex

// 房间用户人数
var GameUserNum = 2

// 游戏中剩余牌
var OtherPoker = map[string]int{}

// 游戏剩余牌索引
var OtherPokerKeys = []string{}

// 游戏玩家手牌
var GameUserPoker = map[int][]map[string]int{}

// 消息模版
type JsonMsg struct {
	Type    string                 `json:"type"`
	Data    map[string]interface{} `json:"data"`
	Message string                 `json:"message"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (this *WebSocket) InitGame() {
	//for _, conn := range Room {
	//	conn.Close()
	//}
	Room = map[int]*websocket.Conn{}
	OtherPoker = map[string]int{}
	OtherPokerKeys = []string{}
	GameUserPoker = map[int][]map[string]int{}
	RoomSession = []string{}
}

func (this *WebSocket) Handfunc(c *gin.Context) {
	conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil) // 忽略错误处理
	defer conn.Close()
	session := sessions.Default(c)
	// 启动一个 goroutine 定期向客户端发送消息
	go func() {
		for {
			pingMap := map[string]string{}
			pingMap["ping"] = "pong"
			marshal, _ := json.Marshal(pingMap)
			err := conn.WriteMessage(websocket.TextMessage, marshal)
			if err != nil {
				log.Println(conn.RemoteAddr(), "心跳发送失败:", err)
				break
			}
			time.Sleep(40 * time.Second) //40秒发送一次心跳包，如果设置60s秒以上须在nginx设置超时时间
		}
	}()

	for {
		// 读取消息
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				// 连接正常关闭或正在关闭
				this.broadcastMessage(JsonMsg{Type: "ErrorGame", Message: "有用户离开了游戏，可以重新开始了"})
			} else {
				this.broadcastMessage(JsonMsg{Type: "ErrorGame", Message: "有用户离开了游戏，可以重新开始了"})
				log.Println("websocket连接异常关闭", err)
			}
			return
		}

		this.OnDoFromMsg(msg, conn, session) //处理到来的消息
	}
}

// 根据客户端发来的消息执行相应操作
func (this *WebSocket) OnDoFromMsg(msg []byte, conn *websocket.Conn, session sessions.Session) {

	jsonMsg := JsonMsg{}
	err := json.Unmarshal(msg, &jsonMsg)
	if err != nil {
		fmt.Println(err)
	}
	g := game.Game{}
	switch jsonMsg.Type {
	//加入房间
	case "joinRoom":
		//检查当前用户session是否在房间中，不存在就添加，存在就不能重复添加
		sessionId := session.Get("session_id")
		RoomSessionLock.Lock()
		for _, s := range RoomSession {
			if s == sessionId.(string) {
				//无论如何都要解锁，否则会死锁
				RoomSessionLock.Unlock()
				return
			}
		}
		RoomSession = append(RoomSession, sessionId.(string))
		RoomSessionLock.Unlock()
		this.JoinRoom(conn)
		break
	case "stopPoker":
		stopData := jsonMsg.Data
		userIdFloat := stopData["userId"].(float64)
		userId := int(userIdFloat)
		switch userId {
		case 2:
			//闲家停牌，告诉庄家可以要牌了
			this.SendMsg(Room[1], JsonMsg{Type: "YouRound"})
			this.SendMsg(Room[2], JsonMsg{Type: "Wait", Message: "等待庄家要牌"})
			break
		case 1:
			//庄家停牌就开始比大小了
			gameFinalMsg := g.GameFinal(GameUserPoker)
			this.broadcastMessage(JsonMsg{Type: "GameOver", Message: gameFinalMsg, Data: map[string]interface{}{
				"poker": GameUserPoker,
			}})
			this.InitGame()
			break

		}
		break
	case "wantPoker":
		stopData := jsonMsg.Data
		userIdFloat := stopData["userId"].(float64)
		userId := int(userIdFloat)
		if len(GameUserPoker[userId]) == 5 {
			//当前用户凑够5张牌了
			this.SendMsg(conn, JsonMsg{Type: "ServerNotPoker", Message: "您以凑够5张牌，无法要牌了"})
			return
		}
		//新发的牌
		newPoker, ok := g.SendPoker(userId, GameUserPoker, OtherPoker, &OtherPokerKeys)

		//通知全体玩家有玩家要了一张牌
		this.broadcastMessage(JsonMsg{Type: "ServerSendPoker", Data: map[string]interface{}{
			"userId": userId,
			"poker":  newPoker,
		}})
		if !ok {
			//明牌牌面总数大于21点，游戏结束
			switch userId {
			case 1:
				this.broadcastMessage(JsonMsg{Type: "GameOver", Message: "游戏结束，闲家胜利，庄家牌面大于21点", Data: map[string]interface{}{
					"poker": GameUserPoker,
				}})

				this.InitGame()
				break
			case 2:
				this.broadcastMessage(JsonMsg{Type: "GameOver", Message: "游戏结束，庄家胜利，闲家牌面大于21点", Data: map[string]interface{}{
					"poker": GameUserPoker,
				}})
				this.InitGame()
				break
			}
		}
		break

	}

}

// 加入房间
func (this *WebSocket) JoinRoom(conn *websocket.Conn) {
	RoomLock.Lock()
	if len(Room) >= 2 {
		RoomLock.Unlock()
		return
	}
	id := len(Room) + 1 //通过当前房间长度分配id
	Room[id] = conn
	RoomLock.Unlock()
	bytes, err := json.Marshal(JsonMsg{Type: "OKJoinRoom", Data: map[string]interface{}{
		"id": id,
	}})
	if err != nil {
		fmt.Println(err)
	}
	conn.WriteMessage(websocket.TextMessage, bytes)
	RoomLock.RLock()
	userNum := len(Room)
	RoomLock.RUnlock()

	//房间内容玩家人数足够，通知全体玩家开始游戏
	if userNum >= GameUserNum {
		g := game.Game{}
		for k, _ := range Room {
			GameUserPoker[k] = []map[string]int{}
		}
		g.Init(GameUserPoker, OtherPoker, &OtherPokerKeys)
		//fmt.Println("当前玩家手牌", GameUserPoker)
		//fmt.Println("剩余手牌", OtherPoker, len(OtherPoker))
		//fmt.Println("剩余牌索引", OtherPokerKeys, len(OtherPokerKeys))
		//给全体玩家发牌
		this.broadcastMessage(JsonMsg{Type: "ServerInitPoker", Data: map[string]interface{}{
			"poker": GameUserPoker,
		}})
		//通知闲家可以进行要牌了
		this.SendMsg(Room[2], JsonMsg{Type: "YouRound"})
		//通知庄家等待闲家要牌
		this.SendMsg(Room[1], JsonMsg{Type: "Wait", Message: "等待闲家要牌"})
	}

}

// 给某个玩家发送消息
func (this *WebSocket) SendMsg(conn *websocket.Conn, jsonMsg JsonMsg) {
	marshal, err := json.Marshal(jsonMsg)
	if err != nil {
		fmt.Println(err)
	}
	conn.WriteMessage(websocket.TextMessage, marshal)
}

// 群发消息
func (this *WebSocket) broadcastMessage(jsonMsg JsonMsg) {
	RoomLock.Lock()
	bytes, err := json.Marshal(jsonMsg)
	if err != nil {
		fmt.Println(err)
	}
	for _, conn := range Room {
		conn.WriteMessage(websocket.TextMessage, bytes)
	}
	RoomLock.Unlock()
}
