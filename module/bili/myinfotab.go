package bili

import (
	log "github.com/ge-fei-fan/gefflog"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func (bt *bilibiliTab) InitMyInfoTab() (err error) {
	tp, err := walk.NewTabPage()
	if err != nil {
		return err
	}
	tp.SetLayout(walk.NewHBoxLayout())
	err = tp.SetTitle(" 基本信息 ")
	if err != nil {
		return err
	}
	if err = (Composite{
		Layout: VBox{
			Margins: Margins{Top: 1},
		},
		Children: []Widget{
			Composite{
				Layout: HBox{Margins: Margins{Top: 5, Bottom: 5}},
				//Alignment: AlignHNearVNear,
				DataBinder: DataBinder{
					Name:           "bt.userInfo",
					DataSource:     bt.userInfo,
					ErrorPresenter: ToolTipErrorPresenter{},
				},
				Children: []Widget{
					Label{
						Text: "个人信息: ",
					},
					LineEdit{
						AssignTo: &bt.nameTE,
						Text:     "未登录",
						ReadOnly: true,
					},
					PushButton{
						AssignTo:  &bt.exitPB,
						Name:      "exit",
						Text:      "退出登录",
						OnClicked: ExitLogin,
						Visible:   false,
					},
					PushButton{
						Text: "设置SESSDATA登录",
						OnClicked: func() {
							if cmd, err := SetDataLogin(bt.win.MainWindow, bt.userInfo); err != nil {
								log.Err(err)
							} else if cmd == walk.DlgCmdOK {
								conf.Sessdata = biliTab.userInfo.Sessdata
								bili.SaveConfig(conf)
								biliTab.userInfo.CheckLoginHandle()
							}
						},
						Visible: Bind("!exit.Visible"),
					},
					PushButton{
						Text: "扫码登录",
						OnClicked: func() {
							qr, err := GetQRCode()
							if err == nil {
								done := make(chan struct{})
								err = CheckQrLogin(qr, done)
								if err != nil {
									log.Err(err)
								} else {
									if cmd, err := QrLogin(bt.win.MainWindow); err != nil {
										log.Err(err)
									} else if cmd == walk.DlgCmdNone { //点了关闭应该把监控协程关了
										close(done)
									}
								}
							} else {
								log.Err(err)
							}
						},
						Visible: Bind("!exit.Visible"),
					},
				},
			},
			GroupBox{
				Title:  "视频配置",
				Layout: Grid{Rows: 2},
				Children: []Widget{
					Composite{
						Layout: HBox{},
						Children: []Widget{
							Label{
								Text: "视频存储路径:",
							},
							LineEdit{
								Text:     "",
								ReadOnly: true,
							},
						},
					},
					Composite{
						Layout: HBox{},
						Children: []Widget{
							HSpacer{},
							PushButton{
								Text: "打开文件夹",
							},
							PushButton{
								Text:    "更改文件夹",
								Enabled: false,
							},
						},
					},
				},
			},
		},
	}).Create(NewBuilder(tp)); err != nil {
		return err
	}
	err = bt.TabWidget.Pages().Add(tp)
	if err != nil {
		return err
	}
	return
}

func ExitLogin() {
	//Sessdata置空
	biliTab.userInfo.Sessdata = ""
	conf.Sessdata = ""
	conf.RefreshToken = ""
	bili.SaveConfig(conf)
	//获取一次个人信息
	biliTab.userInfo.CheckLoginHandle()
}
func SetDataLogin(owner walk.Form, uf *userInfo) (int, error) {
	var dlg *walk.Dialog
	var db *walk.DataBinder
	var acceptPB, cancelPB *walk.PushButton

	return Dialog{
		AssignTo:      &dlg,
		Title:         "SESSDATA登录",
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		DataBinder: DataBinder{
			AssignTo:       &db,
			Name:           "uf",
			DataSource:     uf,
			ErrorPresenter: ToolTipErrorPresenter{},
		},
		MinSize: Size{500, 150},
		Layout:  VBox{},
		Children: []Widget{
			Label{
				Text: "请输入SESSDATA: ",
			},
			LineEdit{
				Text: Bind("Sessdata"),
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						AssignTo: &acceptPB,
						Text:     "登录",
						OnClicked: func() {
							if err := db.Submit(); err != nil {
								log.Err(err)
								return
							}
							dlg.Accept()
						},
					},
					PushButton{
						AssignTo:  &cancelPB,
						Text:      "取消",
						OnClicked: func() { dlg.Cancel() },
					},
				},
			},
		},
	}.Run(owner)
}
func QrLogin(owner walk.Form) (int, error) {
	//walk.Resources.SetRootDirPath("./tab")

	return Dialog{
		AssignTo: &biliTab.qrLoginDlg,
		Title:    "请扫码登录",
		MinSize:  Size{300, 300},
		Layout:   VBox{},
		Children: []Widget{
			ImageView{
				Image: "qr.png",
				Mode:  ImageViewModeCenter,
			},
		},
	}.Run(owner)
}
