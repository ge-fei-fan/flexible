package core

import (
	log "github.com/ge-fei-fan/gefflog"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

//基础设置的tab

type appTab struct {
	*walk.Composite
}

var baseTab = new(appTab)

func ExportTab() *walk.TabPage {
	tp, err := baseTab.init()
	if err != nil {
		log.Err("init tab err", err)
		return nil
	}
	return tp
}

func (at *appTab) init() (tp *walk.TabPage, err error) {

	tp, err = walk.NewTabPage()
	if err != nil {
		return nil, err
	}

	err = tp.SetTitle(" 基本配置 ")
	tp.SetLayout(walk.NewHBoxLayout())
	if err != nil {
		return nil, err
	}
	if err = (Composite{
		AssignTo: &at.Composite,
		//Layout: Flow{
		//	Alignment: AlignHCenterVCenter,
		//},
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
						Text:     "192.192.192.192",
						ReadOnly: true,
					},
					Label{
						Text: ":",
					},
					NumberEdit{
						MinValue: float64(10000),
						MaxValue: float64(65535),
						Value:    float64(12111),
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
								Text: "111",
							},
						},
					},
					Composite{
						Layout: HBox{},
						Children: []Widget{
							HSpacer{},
							PushButton{
								Text: "打开所在文件夹",
							},
							PushButton{
								Text: "打开配置文件",
							},
						},
					},
				},
			},
		},
	}).Create(NewBuilder(tp)); err != nil {
		log.Err(err)
	}
	return
}
