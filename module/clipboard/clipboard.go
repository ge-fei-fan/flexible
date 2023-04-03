package clipboard

import (
	"bytes"
	_ "embed"
	"flexible/core"
	"flexible/extra"
	"fmt"
	"fyne.io/systray"
	log "github.com/ge-fei-fan/gefflog"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/ncruces/zenity"
	"golang.design/x/clipboard"
	"sync"
)

//go:embed connect.ico
var connect []byte

//go:embed disconnect.ico
var disconnect []byte

//go:embed unopened.ico
var unopened []byte

var cp *core.Module
var cpf = clipboardConfig{
	websocketStatus: UNOPENDED,
}

var wsItem *systray.MenuItem

const (
	UNOPENDED  = "unopened"
	DISCONNECT = "disconnect"
	CONNECT    = "connect"
)

type clipboardConfig struct {
	sync.Mutex
	buf             []byte
	WebsocketOpen   bool `json:"WebsocketOpen"`
	websocketStatus string
	client          *extra.WsClient
}

func ExportModule() *core.Module {
	cp = core.NewModule("clipboard", "快速复制", "快速复制", onReady, nil, router, nil)
	return cp
}

func onReady(item *systray.MenuItem) {
	//没有默认配置写入配置
	if cp.Config == nil {
		cp.InitConfig(cpf)
	}
	//加载配置
	cp.UnmarshalConfig(&cpf)
	err := clipboard.Init()
	if err != nil {
		log.Err("clipboard init错误", err)
		zenity.Error("clipboard模块出错，请检查后重新启动软件！！！")
	}

	wsItem = item.AddSubMenuItem("连接远程服务器", "")
	infoItem := item.AddSubMenuItem("使用说明", "")
	wsItem.SetIcon(unopened)
	if cpf.WebsocketOpen {
		connectWs()
	}
	for {
		select {
		case <-wsItem.ClickedCh:
			//fmt.Println(cpf.client)
			//开启ws连接
			if cpf.WebsocketOpen {
				//没连接上：点击一下提示重连还是关闭
				if cpf.websocketStatus == DISCONNECT {
					err := zenity.Question("远程服务器未连接成功，是否重新连接?",
						zenity.Title("clipboard"),
						zenity.OKLabel("关闭"),
						zenity.CancelLabel("重连"))
					if err != nil { //点的重连
						connectWs()
						break
					} else {
						cpf.WebsocketOpen = false
						cpf.client.Close()
						changeWsStatus(UNOPENDED, unopened)
					}
				}
				//已连接上：点击一下直接关闭
				if cpf.websocketStatus == CONNECT {
					cpf.WebsocketOpen = false
					cpf.client.Close()
					changeWsStatus(UNOPENDED, unopened)
				}
			} else { //没开启ws
				//点击一下开启
				cpf.WebsocketOpen = true
				connectWs()
			}
			cp.SaveConfig(cpf)

		case <-infoItem.ClickedCh:
			//fmt.Println(cpf.client)
			zenity.Info("使用post请求服务器路由地址: '/clipboard',表单格式: 'text=内容'，即可快速复制内容到win电脑的粘贴板上",
				zenity.Title("快速复制使用说明"))
		}

	}
}

func router(group *ghttp.RouterGroup) {
	group.POST("/", func(r *ghttp.Request) {
		text := r.GetFormString("text")
		if text != "" {
			ok := cpf.Write(clipboard.FmtText, []byte(text))
			if ok {
				r.Response.WriteJson(g.Map{
					"msg":  "复制成功",
					"code": 0,
				})
			} else {
				r.Response.WriteJson(g.Map{
					"msg":  "复制失败",
					"code": -1,
				})
			}
		}
	})
}

func (cp *clipboardConfig) Read() (t clipboard.Format, buf []byte) {
	cp.Lock()
	defer cp.Unlock()
	defer func() {
		cp.buf = buf
	}()
	t = -1
	buf = clipboard.Read(clipboard.FmtText)
	if buf != nil {
		t = clipboard.FmtText
		return
	}
	return
}

func (cp *clipboardConfig) Write(t clipboard.Format, buf []byte) bool {
	cp.Lock()
	defer cp.Unlock()

	// if the local copy is the same with the write, do not bother.
	if bytes.Equal(cp.buf, buf) {
		return true // but we recognize it as a success write
	}
	cp.buf = buf
	clipboard.Write(t, buf)

	return true
}

func connectWs() {
	if cpf.client == nil {
		c, err := cp.InitWs("clipboard", "192.168.2.122:28080", "wsServe")
		//c, err := cp.InitWs("clipboard", "119.91.149.102:28080", "wsServe")
		cpf.client = c
		go func() {
			for {
				select {
				case s := <-cpf.client.ConnectStatus:
					if s {
						changeWsStatus(CONNECT, connect)
						fmt.Println("连接成功")
					} else {
						changeWsStatus(DISCONNECT, disconnect)
						fmt.Println("连接失败")
					}
				case msg := <-cpf.client.Message:
					ok := cpf.Write(clipboard.FmtText, msg)
					if !ok {
						log.Err("ws msg write err")
					}
				}

			}
		}()
		if err != nil {
			log.Err(cpf.client.Name, "链接websocket错误:", err)
			zenity.Error("连接远程服务器失败", zenity.Title("clipboard"))
			changeWsStatus(DISCONNECT, disconnect)
			return
		}

		changeWsStatus(CONNECT, connect)
		return
	} else {
		err := cpf.client.Dial()
		if err != nil {
			log.Err(cpf.client.Name, "连接websocket错误:", err)
			changeWsStatus(DISCONNECT, disconnect)
		} else {
			changeWsStatus(CONNECT, connect)
		}
	}
}
func changeWsStatus(s string, b []byte) {
	cpf.websocketStatus = s
	wsItem.SetIcon(b)
}
