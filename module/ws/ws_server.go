//go:build ignore
// +build ignore

package main

import (
	log "github.com/ge-fei-fan/gefflog"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
} // use default options
var TextMessage chan []byte

type Text struct {
	Text string `json:"text" form:"text"`
}

var (
	pongWait         = 60 * time.Second  //等待时间
	pingPeriod       = 9 * pongWait / 10 //周期54s
	maxMsgSize int64 = 512               //消息最大长度
	writeWait        = 10 * time.Second  //
)

func clipboard(c *gin.Context) {
	var t Text
	err := c.ShouldBind(&t)
	if err != nil {
		log.Err("解析表单错误", err)
		c.JSON(http.StatusOK, gin.H{
			"msg":  "复制失败",
			"code": -1,
		})
		return
	}
	if t.Text != "" {
		TextMessage <- []byte(t.Text)
		c.JSON(http.StatusOK, gin.H{
			"msg":  "复制成功",
			"code": 0,
		})
	}

}

func wsServe(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Err("Upgrade:", err)
		return
	}
	defer conn.Close()
	done := make(chan struct{})
	go writer(conn, done)
	reader(conn, done)
}
func writer(ws *websocket.Conn, d chan struct{}) {
	defer func() {
		ws.Close()
		log.Debug("write out")
	}()
	for {
		select {
		case <-d:
			return
		//接收到消息就写ws
		case text := <-TextMessage:
			err := ws.WriteMessage(websocket.TextMessage, text)
			if err != nil {
				log.Err("write:", err)
				break
			}
		}
	}
}
func reader(ws *websocket.Conn, d chan struct{}) {
	defer func() {
		ws.Close()
		close(d)
		log.Debug("read out")
	}()
	ws.SetReadLimit(maxMsgSize)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	//ws.SetPingHandler(func(string) error { fmt.Println("接到ping"); return nil })
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func main() {
	log.ChangeLogger(log.DEBUG | log.INFO | log.ERROR)
	TextMessage = make(chan []byte)
	r := gin.Default()

	r.GET("/wsServe", wsServe)
	r.POST("/clipboard", clipboard)

	err := r.Run(":28080")
	if err != nil {
		log.Err("Run:", err)
	}
}
