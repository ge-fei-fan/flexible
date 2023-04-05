package core

import (
	_ "embed"
	"flexible/extra"
	"fmt"
	"fyne.io/systray"
	log "github.com/ge-fei-fan/gefflog"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/lxn/walk"
	"github.com/ncruces/zenity"
	ghook "github.com/robotn/gohook"
	"golang.design/x/clipboard"
	"net"
	"strings"
	"sync"
	"time"
)

//go:embed ico.ico
var iconBs []byte

type App struct {
	title   []*title
	module  []*Module
	config  config
	lock    sync.Mutex
	g       *ghttp.Server
	name    string
	version string
	ico     []byte
	ext     extra.Extra
	localIp string

	//gui
	tabs []*walk.TabPage
	win  *MyMainWindow
}

func NewApp(name, version string, ioc []byte) (*App, error) {

	app := App{
		title:   make([]*title, 0),
		module:  make([]*Module, 0),
		config:  config{configName: "config.json", Module: make(map[string]ModuleConfig)},
		name:    name,
		version: version,
		ico:     ioc,
		//win:     new(MyMainWindow),
		//win:     Flexwiondow(),
	}
	//app.win.TabChangeFn = make(map[string]func())

	if Exists(app.config.configName) {
		err := app.config.load()
		if err != nil {
			return nil, err
		}
	} else {
		err := app.config.save()
		if err != nil {
			return nil, err
		}
	}
	app.ext = extra.NewExtra()
	app.g = g.Server()
	if app.config.HttpPort == 0 {
		app.config.HttpPort = 19999
	}
	app.g.SetPort(int(app.config.HttpPort))
	ip, err := GetOutBoundIP()
	if err != nil {
		log.Err("GetOutBoundIP err:", err)
	}
	app.localIp = ip
	app.Initwiondow()
	return &app, nil
}

func (a *App) onReady() {
	if a.ico != nil {
		systray.SetTemplateIcon(a.ico, a.ico)
	} else {
		systray.SetTemplateIcon(iconBs, iconBs)
	}
	//模块名称添加到systray
	for _, module := range a.module {
		item := systray.AddMenuItem(module.itemName, module.tooltip)
		if module.onReady != nil {
			go module.onReady(item)
		}
	}
	//ip, err := GetOutBoundIP()
	//if err != nil {
	//	log.Err("GetOutBoundIP err:", err)
	//}
	//a.localIp = ip
	//粘贴版监听text
	//todo 监听图片
	go a.ext.Clipboard.Start(clipboard.FmtText)
	go a.ext.Clipboard.Start(clipboard.FmtImage)

	systray.SetTooltip(fmt.Sprintf("%s\n版本号:%s", a.name, a.version))
	infoItem := systray.AddMenuItem("服务器信息", "查看服务器信息")
	setItem := systray.AddMenuItem("配置", "打开配置")
	go func() {
		for {
			select {
			case <-setItem.ClickedCh:
				a.win.Show()
				//cmd := exec.Command("cmd", "/c", "start", a.config.configName)
				//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
				//err := cmd.Run()
				//if err != nil {
				//	log.Err("打开配置文件出错：", err)
				//}
			case <-infoItem.ClickedCh:
				zenity.Info(fmt.Sprintf("当前服务器监听: %s:%d", a.localIp, a.config.HttpPort),
					zenity.Title("服务器基本信息"))
			}
		}

	}()
	quit := systray.AddMenuItem("退出", "退出程序")
	go func() {
		select {
		case <-quit.ClickedCh:
			systray.Quit()
		}
	}()

	//go a.doTitle()
	go a.doTitle2()
	go a.g.Run()

}
func GetOutBoundIP() (ip string, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		return "未知", err
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip = strings.Split(localAddr.String(), ":")[0]
	return ip, nil
}
func (a *App) onExit() {

	// clean up here
	for _, module := range a.module {
		if module.exit != nil {
			go module.exit()
		}
	}
	//粘贴板监控全部退出
	for _, c := range a.ext.Clipboard.Cancel {
		c()
	}
	for _, c := range a.ext.Websocket.Clients {
		c.Close()
		c.Wg.Wait()
	}
	a.win.Synchronize(func() {
		walk.App().Exit(0)
	})
}

func (a *App) Start() error {
	//获取模块配置
	for _, m := range a.module {
		if c, has := a.config.Module[m.name]; has {
			m.Config = c.Config
		}
		m.Port = a.config.HttpPort
		if m.route != nil {
			group := a.g.Group("/" + m.name)
			m.route(group)
		}
	}

	//协程运行systray
	go systray.Run(a.onReady, a.onExit)

	//初始化gui
	//a.Flexwiondow()
	//先注册基本配置tab
	//a.RegisterTab(ExportTab())

	err := a.initTab()
	if err != nil {
		log.Err(err)
	}
	for _, m := range a.module {
		if m.tab != nil {
			a.RegisterTab(m.tab)
		}
	}
	time.Sleep(time.Second)
	//gui run
	a.WinStart()
	return nil
}
func (a *App) Win() *MyMainWindow {
	return a.win
}
func (a *App) WinStart() {
	for _, t := range a.tabs {
		err := a.win.TabWidget.Pages().Add(t)
		if err != nil {
			log.Err("TabWidget add tabpage err:", err)
		}
	}
	a.win.Run()
}
func (a *App) addTitle(module *Module, titleText string) {
	for i := range a.title {
		t := a.title[i]
		if t.module == module {
			t.content = titleText
			return
		}
	}
	a.title = append(a.title, &title{module: module, content: titleText})
}
func (a *App) removeTitle(module *Module) {
	off := 0
	for i := range a.title {
		if a.title[i].module == module {
			off = i
			continue
		}
	}
	a.title = append(a.title[:off], a.title[off+1:]...)
}
func (a *App) doTitle() {
	for {
		for _, t := range a.title {
			a.lock.Lock()
			systray.SetTitle(t.content)
			a.lock.Unlock()
		}
		time.Sleep(time.Second)
	}
}

// 测试一下只设置一次可以嘛
func (a *App) doTitle2() {
	for _, t := range a.title {
		a.lock.Lock()
		systray.SetTitle(t.content)
		a.lock.Unlock()
	}
}
func (a *App) RegisterModule(module ...*Module) {
	//a.Initwiondow()
	for i := range module {
		m := module[i]
		mc, has := a.config.Module[m.name]
		if has && !mc.Enable {
			continue
		}
		m.app = a
		a.module = append(a.module, m)
	}
}
func (a *App) GetModule() []*Module {
	return a.module
}
func (a *App) RegisterTab(tabpage ...*walk.TabPage) {
	for _, tp := range tabpage {
		if tp != nil {
			a.tabs = append(a.tabs, tp)
		}
	}
}

//func RegisterTab(tabpage ...*walk.TabPage) {
//	for _, tp := range tabpage {
//		if tp != nil {
//			tabs = append(tabs, tp)
//		}
//	}
//}

type title struct {
	module  *Module
	content string
}
type Tab struct {
}
type Module struct {
	onReady   func(item *systray.MenuItem)
	exit      func()
	app       *App
	name      string
	itemName  string
	tooltip   string
	Config    interface{}
	route     func(*ghttp.RouterGroup)
	Port      uint16
	eventChan chan ghook.Event
	tab       *walk.TabPage
}

func NewModule(name, itemName, tooltip string, onReady func(item *systray.MenuItem), exit func(), route func(*ghttp.RouterGroup), tab *walk.TabPage) *Module {
	module := Module{
		onReady:  onReady,
		exit:     exit,
		name:     name,
		itemName: itemName,
		tooltip:  tooltip,
		route:    route,
		tab:      tab,
	}
	return &module
}
func (m *Module) UnmarshalConfig(dst interface{}) error {
	return Unmarshal(m.Config, dst)
}
func (m *Module) SaveConfig(c interface{}) {
	m.Config = c
	m.app.config.saveConfig(m, m.Config)
	err := m.app.config.save()
	if err != nil {
		log.Err(err)
	}
}
func (m *Module) InitConfig(c interface{}) {
	m.Config = c
	m.app.config.initConfig(m, m.Config)
	err := m.app.config.save()
	if err != nil {
		log.Err(err)
	}
}
func (m *Module) SetTitle(title string) {
	m.app.addTitle(m, title)
}
func (m *Module) RemoveTitle() {
	m.app.removeTitle(m)
}
func (m *Module) GetAPPName() string {
	return m.app.name
}
func (m *Module) GetAPPVersion() string {
	return m.app.version
}
func (m *Module) RegEvent() chan ghook.Event {
	m.eventChan = make(chan ghook.Event, 0)
	return m.eventChan
}

// 监控粘贴板
func (m *Module) Watch() chan extra.Package {
	return m.app.ext.Clipboard.Watch()
}
func (m *Module) StopWatch(ch chan extra.Package) {
	m.app.ext.Clipboard.StopWatch(ch)
}

// websocket服务
func (m *Module) InitWs(name, host, path string) (*extra.WsClient, error) {
	c, has := m.app.ext.Websocket.Clients[name]
	if has {
		return c, nil
	}
	c = extra.NewWsClient(name, host, path)
	err := c.Dial()
	if err != nil {
		return c, err
	}
	return c, nil
}

func (m *Module) GetWindowS() *walk.MainWindow {
	return m.app.win.MainWindow
	//return MMW.MainWindow
}

func (m *Module) SetTabChangeFn(fn func()) {
	m.app.win.TabChangeFn[m.name] = fn
}
