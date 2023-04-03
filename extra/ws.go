package extra

import (
	"errors"
	log "github.com/ge-fei-fan/gefflog"
	"github.com/gorilla/websocket"
	"net/url"
	"sync"
	"time"
)

var (
	pongWait         = 60 * time.Second  //等待时间
	pingPeriod       = 9 * pongWait / 10 //周期54s
	maxMsgSize int64 = 512               //消息最大长度
	writeWait        = 10 * time.Second  //
)
var hub *wsHub

type wsHub struct {
	Clients       map[string]*WsClient
	retryInterval int64
}

type WsClient struct {
	Name  string
	conn  *websocket.Conn
	wsUrl url.URL
	mutex sync.Mutex
	Wg    sync.WaitGroup
	done  chan struct{}
	//消息接收通道
	Message chan []byte
	//发送通道
	Send chan []byte
	//是否是主动关闭
	isClose       bool
	ConnectStatus chan bool
}

//	type pack struct {
//		T int
//		Msg []byte
//	}
func newWsHub() *wsHub {
	hub = &wsHub{
		Clients:       make(map[string]*WsClient),
		retryInterval: 60,
	}
	return hub
}
func (h *wsHub) reConnect(ws *WsClient) {
	//重连时间
	var interval int64 = 0
	//重连5次,每次重试间隔加60s
	for i := 0; i < 5; i++ {
		err := ws.Dial()
		if err == nil {
			ws.ConnectStatus <- true
			break
		}
		time.Sleep(time.Duration(interval) * time.Second)
		interval += h.retryInterval
	}
}

func NewWsClient(name, host, path string) *WsClient {
	c := WsClient{
		Name: name,
		wsUrl: url.URL{
			Scheme: "ws",
			Host:   host,
			Path:   path,
		},
		Message:       make(chan []byte),
		Send:          make(chan []byte),
		isClose:       false,
		ConnectStatus: make(chan bool),
	}
	return &c
}

func (ws *WsClient) Dial() error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	if hub == nil {
		return errors.New("wsHub没有初始化")
	}
	_, has := hub.Clients[ws.Name]
	if has {
		return nil
	}
	c, _, err := websocket.DefaultDialer.Dial(ws.wsUrl.String(), nil)
	if err != nil {
		return err
	}
	ws.conn = c
	ws.isClose = false //重置是否主动退出
	ws.done = make(chan struct{})
	hub.Clients[ws.Name] = ws
	go ws.writer()
	go ws.reader()
	return nil
}

func (ws *WsClient) reader() {
	log.Info(ws.Name, ":reader开始")
	ws.Wg.Add(1)
	defer func() {
		log.Info(ws.Name, ":reader退出")
		ws.Wg.Done()
	_:
		ws.conn.Close()
		close(ws.done)
		delete(hub.Clients, ws.Name)
		if !ws.isClose {
			ws.ConnectStatus <- false
			go hub.reConnect(ws)
		}

	}()
	ws.conn.SetReadLimit(maxMsgSize)
	ws.conn.SetReadDeadline(time.Now().Add(pongWait))
	ws.conn.SetPongHandler(func(string) error {
		ws.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, msg, err := ws.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Err(err)
			}
			break
		}
		ws.Message <- msg
	}
}
func (ws *WsClient) writer() {
	log.Info(ws.Name, ":writer开始")
	ws.Wg.Add(1)
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		log.Info(ws.Name, ":writer退出")
		ws.Wg.Done()
		ticker.Stop()
		ws.conn.Close()
	}()
	for {
		select {
		case <-ws.done:
			return
		case msg := <-ws.Send:
			if ws.isClose {
				err := ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Err(ws.Name, ":write close:", err)
					return
				}
			} else {
				err := ws.conn.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					log.Err(ws.Name, ":write:", err)
					return
				}
			}
		case <-ticker.C:
			ws.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
func (ws *WsClient) Close() {
	//检查一下连接是否健康，是不是还在hub里
	_, has := hub.Clients[ws.Name]
	if has {
		ws.isClose = true
		ws.Send <- []byte("close")
	}

}
