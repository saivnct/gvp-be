package appws

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"gbb.go/gvp/dao"
	"gbb.go/gvp/model"
	"gbb.go/gvp/utils"
	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
	"golang.org/x/exp/slices"
	"log"
	"os"
	"sync"
	"time"
)

type WsCreds struct {
	User *model.User
}

//type WsResponse struct {
//	Status  int    `json:"status"`
//	Message string `json:"message"`
//}

type WS struct {
	Server           *socketio.Server
	SendWSMsgChannel chan *WSSendMsg
}

type WSSendMsg struct {
	Topic string
	Msg   string
}

const (
	WS_CRED_AUTH = "authorization"
)

var (
	WSNotFoundSocketErr = errors.New("Not found socket session")
	rWlock              = sync.RWMutex{}

	socketlist   = make(map[string][]string) //username - list sockerID
	totalSession = 0

	singletonWS *WS
	onceWS      sync.Once
)

func GetWS() *WS {
	onceWS.Do(func() {
		fmt.Println("Init WS...")

		server, err := WSServer()
		if err != nil {
			log.Fatalf("Failed to init WS: %v", err)
		}

		singletonWS = &WS{
			Server: server,
		}
	})
	return singletonWS
}

func (ws *WS) Start() {
	sendWSMsgChannel := make(chan *WSSendMsg, 1000) //make a buffered channel size 1000
	go func(sendWSMsgChannel chan *WSSendMsg) {
		for wsSendMsg := range sendWSMsgChannel {
			err := ws.doSendMsg(wsSendMsg)
			if err != nil {
				log.Println("socketio - sendMsg error", err)
			}
		}
	}(sendWSMsgChannel)

	ws.SendWSMsgChannel = sendWSMsgChannel
}

func (ws *WS) Stop() {
	if ws.SendWSMsgChannel != nil {
		close(ws.SendWSMsgChannel)
	}
}

func (ws *WS) SendMsg(topic string, msg string) {
	ws.SendWSMsgChannel <- &WSSendMsg{
		Topic: topic,
		Msg:   msg,
	}
}

func WSServer() (*socketio.Server, error) {
	wsTransport := websocket.Default
	//pollingTransport := polling.Default

	opts := &engineio.Options{
		PingTimeout:  15 * time.Second,
		PingInterval: 10 * time.Second,
		Transports: []transport.Transport{
			wsTransport,
			//pollingTransport,
		},
	}

	server := socketio.NewServer(opts)

	_, err := server.Adapter(&socketio.RedisAdapterOptions{
		Addr:     fmt.Sprintf("%v:%v", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWD"),
		Prefix:   "socket.io",
	})
	if err != nil {
		log.Fatal("error:", err)
	}

	server.OnConnect("/", func(s socketio.Conn) error {
		//NOTE : Calling Emit inside the OnConnect handler breaks because the connect request is not completed when the handler is called - https://github.com/googollee/go-socket.io/pull/221
		//log.Printf("socketio - OnConnect")

		url := s.URL()
		jwtToken := url.Query().Get(WS_CRED_AUTH)

		jwtPayload, err := hex.DecodeString(jwtToken)
		if err != nil {
			log.Println("WS - Invalid decode jwtToken", err)
			return nil
		}

		claims, _ := utils.ParseJWTToken(string(jwtPayload))
		userName := claims.Subject

		user, _ := dao.GetUserDAO().FindByUserName(context.Background(), userName)

		s.SetContext(WsCreds{
			User: user,
		})

		s.Join(userName)

		rWlock.Lock() //lock any Writer and Readers

		totalSession++

		sockets, existed := socketlist[userName]
		if !existed {
			sockets = []string{}
		}

		sockets = append(sockets, s.ID())
		socketlist[userName] = sockets

		rWlock.Unlock()

		log.Printf("socketio - new client of %v connected, socketId: %v - total session of %v: %v - total system session: %v\n",
			userName, s.ID(), userName, len(sockets), totalSession)

		return nil
	})

	server.OnEvent("/", "sendMsg", func(s socketio.Conn, data string) *string {
		//log.Println("socketio - on sendMsg")
		wsCreds, ok := s.Context().(WsCreds)
		if !ok {
			log.Println("socketio - on sendSignalMsg error - can not cast WsCreds")
			return nil
		}

		log.Printf("socketio - on sendMsg from %v: %v", wsCreds.User.Username, data)

		retsult := ""
		return &retsult
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		log.Println("socketio - OnError:", e)
		//s.Emit("disconnect") // => FORCE CLIENT DISCONNECT
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		//log.Println("socketio - client disconnect:", s.ID(), reason)
		wsCreds, ok := s.Context().(WsCreds)
		if !ok {
			log.Println("socketio - client disconnect can not cast WsCreds")
			return
		}

		rWlock.Lock() //lock any Writer and Readers

		totalSession--

		sockets, existed := socketlist[wsCreds.User.Username]
		if !existed {
			log.Println("socketio - client disconnect session not found", wsCreds.User.Username, s.ID())
		} else {
			idx := slices.IndexFunc(sockets, func(socketID string) bool { return socketID == s.ID() })
			if idx >= 0 {
				sockets = slices.Delete(sockets, idx, idx+1)
			}

			if len(sockets) == 0 {
				delete(socketlist, wsCreds.User.Username)
			} else {
				socketlist[wsCreds.User.Username] = sockets
			}

			log.Printf("socketio - client of %v disconnected, socketId: %v, - total session of %v: %v - total system session: %v\n",
				wsCreds.User.Username, s.ID(), wsCreds.User.Username, len(sockets), totalSession)
		}
		rWlock.Unlock()

	})

	return server, nil
}

func (ws *WS) doSendMsg(wsSendMsg *WSSendMsg) error {
	//TODO
	return nil
}
