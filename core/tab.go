package core

import (
	log "github.com/ge-fei-fan/gefflog"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"os"
	"os/exec"
	"syscall"
)

//基础设置的tab

//type appTab struct {
//	*walk.Composite
//}
//
//var baseTab = new(appTab)
//
//func ExportTab() *walk.TabPage {
//	tp, err := baseTab.init()
//	if err != nil {
//		log.Err("init tab err", err)
//		return nil
//	}
//	return tp
//}

func (a *App) initTab() error {
	//fmt.Println(a.localIp)
	tp, err := walk.NewTabPage()
	if err != nil {
		return err
	}
	err = tp.SetTitle(" 基本配置 ")
	tp.SetLayout(walk.NewHBoxLayout())
	if err != nil {
		return err
	}
	if err = (Composite{
		Layout: VBox{
			Margins: Margins{Bottom: 1},
		},
		Children: []Widget{
			GroupBox{
				Title:  "服务器信息",
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "监听地址:",
					},
					LineEdit{
						MaxSize:  Size{Width: 180},
						Text:     a.localIp,
						ReadOnly: true,
					},
					Label{
						Text: ":",
					},
					NumberEdit{
						MinValue: float64(10000),
						MaxValue: float64(65535),
						Value:    float64(a.config.HttpPort),
					},
				},
			},
			GroupBox{
				Title: "配置文件存储位置",
				//Layout: VBox{MarginsZero: true},
				Layout: Grid{Rows: 2},
				Children: []Widget{
					Composite{
						Layout: HBox{},
						Children: []Widget{
							Label{
								Text: "路径:",
							},
							LineEdit{
								Text: a.config.configName,
							},
						},
					},
					Composite{
						Layout: HBox{},
						Children: []Widget{
							HSpacer{},
							PushButton{
								Text: "打开所在文件夹",
								OnClicked: func() {
									dir, err := os.Getwd()
									if err != nil {
										log.Err(err)
										return
									}
									//fmt.Println(dir)
									cmd := exec.Command("cmd", "/c", "explorer", dir)
									cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
									_, _ = cmd.Output()
									//fmt.Println(out)
									//if err != nil {
									//	fmt.Println(err)
									//	log.Err("打开配置文件出错：", err)
									//}
								},
							},
							PushButton{
								Text: "打开配置文件",
								OnClicked: func() {
									cmd := exec.Command("cmd", "/c", "start", a.config.configName)
									cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
									_, err := cmd.Output()
									if err != nil {
										log.Err("打开配置文件出错：", err)
									}
								},
							},
						},
					},
				},
			},
		},
	}).Create(NewBuilder(tp)); err != nil {
		log.Err(err)
	}
	a.tabs = append(a.tabs, tp)
	return err
}
