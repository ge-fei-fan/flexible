package bili

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"time"
)

type TableModel struct {
	walk.TableModelBase
	items []*videoInfo
}

func (tm *TableModel) Value(row, col int) interface{} {
	item := tm.items[row]

	switch col {
	case 0:
		return ""
	case 1:
		return item.Name

	case 2:
		return item.Quality

	case 3:
		return item.FinishTime

	case 4:
		return item.Result
	}

	panic("unexpected col")
}
func (tm *TableModel) RowCount() int {
	return len(tm.items)
}
func newVideoInfo(name, quality, result string) *videoInfo {
	now := time.Now()
	return &videoInfo{
		Name:       name,
		Quality:    quality,
		FinishTime: now.Format("2006-01-02 15:04:05"),
		Result:     result,
	}
}
func (vi *videoInfo) handle() {
	defer biliTab.tb.PublishRowsReset()
	for i, _ := range biliTab.tb.items {
		if biliTab.tb.items[i].Name == vi.Name {
			biliTab.tb.items[i] = vi
			return
		}
	}
	biliTab.tb.items = append(biliTab.tb.items, vi)
}

var TableTabCom *walk.Composite

func (bt *bilibiliTab) InitVideoTableTab() (err error) {
	tp, err := walk.NewTabPage()
	if err != nil {
		return err
	}
	tp.SetLayout(walk.NewHBoxLayout())
	err = tp.SetTitle(" 下载视频 ")
	if err != nil {
		return err
	}
	if err = (Composite{
		//AssignTo: &bt.Composite,
		Layout: VBox{},
		Children: []Widget{
			Composite{
				AssignTo: &TableTabCom,
				Layout: VBox{
					Margins: Margins{Top: 1, Bottom: 1},
				},
				//MinSize: Size{Height: 200},
				Children: []Widget{
					TableView{
						AssignTo:    &bt.tv,
						ToolTipText: "表格数据",
						Columns: []TableViewColumn{
							{Title: "#", Width: 1},
							// Name is needed for settings persistence
							{Title: "视频名称", Alignment: AlignCenter, Width: 160},
							{Title: "清晰度", Alignment: AlignCenter, Width: 100},
							{Title: "下载日期", Format: "2006-01-02 15:04:05", Alignment: AlignCenter, Width: 130},
							{Title: "下载结果", Alignment: AlignCenter, Width: 100},
						},
						Model: bt.tb,
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
