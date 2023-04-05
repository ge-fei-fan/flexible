package core

import (
	log "github.com/ge-fei-fan/gefflog"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

type MyMainWindow struct {
	//app *App
	*walk.MainWindow
	hWnd        win.HWND
	TabWidget   *walk.TabWidget
	TabChangeFn map[string]func()
}

func (a *App) Initwiondow() {
	a.win = new(MyMainWindow)
	a.win.TabChangeFn = make(map[string]func())
	//mmw.app = a
	icon, err := walk.Resources.Icon("ico.ico")
	if err != nil {
		log.Err(err)
	}
	if err := (MainWindow{
		Title:    "flexible",
		AssignTo: &a.win.MainWindow,
		Icon:     icon,
		Size:     Size{400, 350},
		Visible:  false,
		//Layout:   VBox{MarginsZero: true, SpacingZero: true},
		Layout: VBox{},
		Children: []Widget{
			TabWidget{
				AssignTo: &a.win.TabWidget,
				OnCurrentIndexChanged: func() {
					e := a.win.TabWidget.CurrentIndexChanged()
					e.Attach(func() {
						index := a.win.TabWidget.CurrentIndex()
						if index == 0 {
							a.win.SetSize(walk.Size{400, 350})
						} else {
							//name := a.GetModule()[index-1].name
							//fn, has := mmw.TabChangeFn[name]
							//if has {
							//	fn()
							//}
							i := 0
							for _, m := range a.module {
								if m.tab == nil {
									continue
								}
								i = i + 1
								if i == index {
									fn, has := a.win.TabChangeFn[m.name]
									if has {
										fn()
										break
									}
								}
							}
						}
					})
				},
			},
		},
	}.Create()); err != nil {
		log.Err("创建gui出错", err)
	}
	a.win.hWnd = a.win.Handle()
	//窗口置顶，暂时能用，可能开了弹窗后就不好用了
	win.SetWindowPos(a.win.hWnd,
		win.HWND_TOPMOST, 0, 0, 0, 0,
		win.SWP_NOACTIVATE|win.SWP_NOMOVE|win.SWP_NOSIZE)
	a.win.removeStyle(^win.WS_SIZEBOX)
	// 去使能最小化按钮
	a.win.removeStyle(^win.WS_MINIMIZEBOX)
	// 去使能最大化按钮
	a.win.removeStyle(^win.WS_MAXIMIZEBOX)
	//mmw.removeStyle(^win.SW_HIDE)
	// 设置窗体生成在屏幕的正中间
	// 窗体横坐标 = ( 屏幕宽度 - 窗体宽度 ) / 2
	// 窗体纵坐标 = ( 屏幕高度 - 窗体高度 ) / 2
	a.win.SetX((int(win.GetSystemMetrics(0)) - a.win.Width()) / 2)
	a.win.SetY((int(win.GetSystemMetrics(1)) - a.win.Height()) / 2)
	//关闭事件不退出
	a.win.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		*canceled = true
		a.win.Hide()
	})

}
func (mw *MyMainWindow) removeStyle(style int32) {
	currStyle := win.GetWindowLong(mw.hWnd, win.GWL_STYLE)
	win.SetWindowLong(mw.hWnd, win.GWL_STYLE, currStyle&style)
}

func (mw *MyMainWindow) addStyle(style int32) {
	currStyle := win.GetWindowLong(mw.hWnd, win.GWL_STYLE)
	win.SetWindowLong(mw.hWnd, win.GWL_STYLE, currStyle|style)
}
