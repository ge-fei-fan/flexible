package ws

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"testing"
	"time"
)

func TestS(t *testing.T) {

	u := url.URL{Scheme: "ws", Host: "192.168.2.122:28080", Path: "/clipboardws"}
	log.Printf("connecting to %s", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", message)
		}
	}()
	pingTicker := time.NewTicker(5 * time.Second)
	defer pingTicker.Stop()
	for {
		select {
		case <-pingTicker.C:
			c.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := c.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				break
			}
			fmt.Println("client send ping")
		}

	}
	//interrupt := make(chan os.Signal, 1)
	//<-interrupt
	//time.Sleep(30 * time.Second)
	//err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	//if err != nil {
	//	log.Println("write close:", err)
	//	return
	//}

}
