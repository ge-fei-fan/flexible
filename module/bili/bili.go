package bili

import (
	"flexible/core"
	"fyne.io/systray"
	log "github.com/ge-fei-fan/gefflog"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/lxn/walk"
	"github.com/ncruces/zenity"
	"golang.design/x/clipboard"
	"time"
)

var bili *core.Module

type handleResult interface {
	handle()
}
type biliConfig struct {
	//视频存储路径
	Path         string `json:"path"`
	Sessdata     string `json:"sessdata"`
	RefreshToken string `json:"refreshtoken"`
	//自动检测下载
	Auto       bool `json:"auto"`
	userHub    *UserHub
	commonChan chan handleResult
}

// 下载结束通知
type finishNotify struct {
	msg string
	err error
}

func NewfinishNotify(m string, e error) *finishNotify {
	f := finishNotify{
		msg: m,
		err: e,
	}
	return &f
}
func (f *finishNotify) handle() {
	if f.err != nil {
		log.Err(f.msg, f.err)
		zenity.Error(f.msg)
	}
}

var conf = biliConfig{
	Path:       "bilibili",
	Auto:       false,
	userHub:    NewUserHub(),
	commonChan: make(chan handleResult, 20),
}

func ExportModule(win *core.MyMainWindow) *core.Module {
	bili = core.NewModule("bilibili", "bilibili", "B站管理", onReady, onExit, router, ExportTab(win))
	return bili
}
func onReady(item *systray.MenuItem) {
	//获取通知协程
	go func() {
		for c := range conf.commonChan {
			c.handle()
		}
	}()
	//没有默认配置写入配置
	if bili.Config == nil {
		bili.InitConfig(conf)
	}
	//加载配置
	bili.UnmarshalConfig(&conf)

	err := clipboard.Init()
	if err != nil {
		zenity.Error("获取粘贴板模块出错，无法使用自动下载功能")
		conf.Auto = false
	}
	if conf.Auto {
		biliTab.autoDownloadCheck.SetCheckState(walk.CheckChecked) //设置walk控件勾选
		go startWatchD()
	}

	autoUrlItem := item.AddSubMenuItemCheckbox("自动下载", "从粘贴板获取链接自动确认下载", conf.Auto)
	videoDown := item.AddSubMenuItem("视频下载", "")
	bConf := item.AddSubMenuItem("设置SESSDATA", "")
	loginMenu := item.AddSubMenuItem("账号未登录", "")
	biliTab.userInfo.CheckLoginHandle = func() {
		checkLogin(loginMenu)
	}
	biliTab.changeAutoD = func(check bool) {
		if check {
			autoUrlItem.Check()
			go startWatchD()
		} else {
			autoUrlItem.Uncheck()
			stopWatchD()
		}
		conf.Auto = check
		bili.SaveConfig(conf)
	}
	go checkLogin(loginMenu)
	go timeCheckLogin(loginMenu)
	//监控最新视频
	err = conf.userHub.load()
	if err != nil {
		log.Err("userHub load err:", err)
	}
	//go conf.userHub.Start()

	for {
		select {
		case <-bConf.ClickedCh:
			entry, err := zenity.Entry("请输入SESSDATA")
			if err == nil {
				conf.Sessdata = entry
				bili.SaveConfig(conf)
				checkLogin(loginMenu)
			}
		case <-videoDown.ClickedCh:
			entry, err := zenity.Entry("请输入视频地址")
			if err == nil {
				u := filiterCpUrl(entry)
				if u == "" {
					conf.commonChan <- NewfinishNotify("地址错误，获取bvid失败", nil)
					continue
				}
				go downloadVideo(u)
			}
		case <-autoUrlItem.ClickedCh:
			if autoUrlItem.Checked() {
				autoUrlItem.Uncheck()
				biliTab.autoDownloadCheck.SetCheckState(walk.CheckUnchecked)
				stopWatchD()
			} else {
				autoUrlItem.Check()
				biliTab.autoDownloadCheck.SetCheckState(walk.CheckChecked)
				go startWatchD()
			}
			conf.Auto = autoUrlItem.Checked()
			bili.SaveConfig(conf)
		}
	}
}
func onExit() {
	close(conf.commonChan)
	close(conf.userHub.done)
}

func router(group *ghttp.RouterGroup) {
	group.POST("/savevideo", func(r *ghttp.Request) {
		url := r.GetFormString("url")
		err := parseUrl(url)
		if err != nil {
			r.Response.WriteJson(g.Map{
				"msg":  "解析链接失败",
				"code": -1,
			})
		} else {
			r.Response.WriteJson(g.Map{
				"msg":  "开始下载",
				"code": 0,
			})
		}
	})
	group.GET("/adduser/:mid", func(r *ghttp.Request) {
		m := r.GetString("mid")
		mid := FiliterMid(m)
		if mid == "" {
			r.Response.WriteJson(g.Map{
				"msg":  "mid为空",
				"code": -1,
			})
			return
		}
		err := conf.userHub.AddUser(mid)
		if err != nil {
			log.Err("adduser", err)
			r.Response.WriteJson(g.Map{
				"msg":  "添加失败",
				"code": -1,
			})
		}
		r.Response.WriteJson(g.Map{
			"msg":  "添加成功",
			"code": 0,
		})
	})
}

// 定时检查登录状态
func timeCheckLogin(item *systray.MenuItem) {
	timer := time.NewTicker(3 * time.Hour)
	defer func() {
		timer.Stop()
	}()
	for {
		<-timer.C
		log.Info("检查bilibili登录状态")
		checkLogin(item)
	}
}
