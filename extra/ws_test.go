package extra

import (
	"fmt"
	"log"
	"testing"
	"time"
)

//func TestName(t *testing.T) {
//	a := make(map[int]string)
//	a[1] = "1"
//	v, has := a[1]
//	fmt.Println(v, has)
//}

func TestClient(t *testing.T) {
	hub = newWsHub()
	c := NewWsClient("test", "192.168.2.122:28080", "clipboardws")
	err := c.Dial()
	if err != nil {
		log.Println(err)
	}

	go func() {
		for {
			if <-c.ConnectStatus {
				log.Println("connect")
			} else {
				log.Println("disconnect")
			}
		}
	}()

	time.Sleep(time.Second * 70)
	_, has := hub.Clients[c.Name]
	fmt.Println(has)
	c.Close()
}
