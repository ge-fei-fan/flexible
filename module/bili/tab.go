package bili

import (
	"flexible/core"
	log "github.com/ge-fei-fan/gefflog"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type bilibiliTab struct {
	win *core.MyMainWindow
	*walk.Composite
	autoDownloadCheck *walk.CheckBox
	collectionCheck   *walk.CheckBox
	TabWidget         *walk.TabWidget
	nameTE            *walk.LineEdit
	tv                *walk.TableView
	userInfo          *userInfo
	tb                *TableModel
	exitPB            *walk.PushButton
	qrLoginDlg        *walk.Dialog
	changeAutoD       func(bool)
}
type userInfo struct {
	Name string
	//IsLogin          bool
	Sessdata         string
	CheckLoginHandle func()
}

var biliTab = &bilibiliTab{
	//autoDownloadCheck: new(walk.CheckBox),
	userInfo: &userInfo{},
	tb:       &TableModel{},
}

func ExportTab(win *core.MyMainWindow) *walk.TabPage {
	biliTab.win = win
	tp, err := biliTab.InitTab()
	if err != nil {
		log.Err("init tab err", err)
		return nil
	}
	return tp
}

type videoInfo struct {
	Name       string
	Quality    string
	FinishTime string
	Result     string
}

func (bt *bilibiliTab) InitTab() (tp *walk.TabPage, err error) {
	var intervalEdit *walk.NumberEdit

	tp, err = walk.NewTabPage()
	if err != nil {
		return nil, err
	}
	//设置背景颜色
	//bgColor := walk.RGB(249, 249, 249)
	//scb, _ := SolidColorBrush{bgColor}.Create()
	//tp.SetBackground(scb)
	//设置水平布局，感觉垂直水平没差
	tp.SetLayout(walk.NewHBoxLayout())
	err = tp.SetTitle(" B站 ")
	if err != nil {
		return nil, err
	}
	if err = (Composite{
		AssignTo: &bt.Composite,
		Layout: VBox{
			Margins: Margins{Bottom: 1},
		},
		//Background: SolidColorBrush{Color: walk.RGB(255, 0, 0)},
		Children: []Widget{
			Composite{
				Layout: VBox{
					Margins: Margins{Top: 1},
				},
				//MinSize:    Size{Width: 450},
				//Background: SolidColorBrush{Color: walk.RGB(255, 0, 0)},
				Children: []Widget{
					Composite{
						Layout:    HBox{MarginsZero: true},
						Alignment: AlignHNearVCenter,
						Children: []Widget{
							CheckBox{
								AssignTo:       &bt.autoDownloadCheck,
								Text:           "自动下载",
								TextOnLeftSide: true,
								ToolTipText:    "选中后能够监听粘贴板自动是否有B站视频链接",
								//Checked:        false, //注意：true 默认选中，false:默认未选
								OnClicked: func() {
									//fmt.Println(bt.autoDownloadCheck.Checked())
									biliTab.changeAutoD(bt.autoDownloadCheck.Checked())
								},
							},
							HSpacer{MaxSize: Size{Width: 20}},
							CheckBox{
								AssignTo:       &bt.collectionCheck,
								Text:           "采集视频",
								TextOnLeftSide: true,
								ToolTipText:    "选中后能够自动获取up主最新投稿",
								Checked:        false, //注意：true 默认选中，false:默认未选
								OnClicked: func() {
									if bt.collectionCheck.Checked() {
										conf.IsCollect = true
										go conf.userHub.Start()
									} else {
										conf.IsCollect = false
										close(conf.userHub.done)
									}
									bili.SaveConfig(conf)
								},
							},
							HSpacer{MaxSize: Size{Width: 20}},
							Label{
								Text: "采集间隔: ",
							},
							NumberEdit{
								AssignTo: &intervalEdit,
								Value:    float64(conf.ViedoInterval),
								OnValueChanged: func() {
									conf.ViedoInterval = int64(intervalEdit.Value())
									//walk.MsgBox(bt.win.MainWindow, "提示", "重新开启采集功能才会生效", walk.MsgBoxIconInformation)
								},
							},
							Label{
								Text:        "※",
								ToolTipText: "修改采集时间后，需要重新开启采集功能才会生效",
							},
							HSpacer{},
						},
					},
					TabWidget{
						AssignTo: &bt.TabWidget,
						OnCurrentIndexChanged: func() {
							e := bt.TabWidget.CurrentIndexChanged()
							e.Attach(func() {
								index := bt.TabWidget.CurrentIndex()
								switch index {
								case 0:
									bt.win.MainWindow.SetSize(walk.Size{400, 350})
									TableTabCom.SetSize(walk.Size{Height: 0})
								case 1:
									bt.win.MainWindow.SetSize(walk.Size{600, 400})
									TableTabCom.SetSize(walk.Size{Height: 200})
								}
							})
						},
					},
				},
			},
		},
	}).Create(NewBuilder(tp)); err != nil {
		return nil, err
	}
	//if conf.Auto {
	//	biliTab.autoDownloadCheck.SetCheckState(walk.CheckChecked) //设置walk控件勾选
	//}
	bt.win.TabChangeFn["bilibili"] = func() {
		index := bt.TabWidget.CurrentIndex()
		switch index {
		case 0:
			bt.win.MainWindow.SetSize(walk.Size{400, 350})
			TableTabCom.SetSize(walk.Size{Height: 0})
		case 1:
			bt.win.MainWindow.SetSize(walk.Size{600, 400})
			TableTabCom.SetSize(walk.Size{Height: 200})
		}
	}
	err = bt.InitMyInfoTab()
	if err != nil {
		log.Err(err)
	}
	err = bt.InitVideoTableTab()
	if err != nil {
		log.Err(err)
	}

	return
}
