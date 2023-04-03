package extra

import (
	"context"
	log "github.com/ge-fei-fan/gefflog"
	"golang.design/x/clipboard"
	"sync"
)

type Appclipboard struct {
	//watch的cancel
	Cancel map[clipboard.Format]context.CancelFunc
	//转发数据
	notifyChans []chan Package
	sync.Mutex
}

type Package struct {
	T MIME
	B []byte
}

func newAppcb() *Appclipboard {
	return &Appclipboard{
		Cancel: make(map[clipboard.Format]context.CancelFunc),
	}
}

// 开启一个 clipboard.Format类型的 粘贴板监听
func (a *Appclipboard) Start(t clipboard.Format) {
	_, ok := a.Cancel[t]
	if ok {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.Mutex.Lock()
	a.Cancel[t] = cancel
	a.Mutex.Unlock()
	textCh := clipboard.Watch(ctx, t)
	format := MIMEPlainText
	if t == clipboard.FmtImage {
		format = MIMEImagePNG
	}
	log.Info(format, "开启粘贴板监控")
	defer func() {
		log.Info(format, "关闭粘贴板监控，线程退出...")
	}()
	for text := range textCh {
		if len(a.notifyChans) == 0 {
			continue
		}
		p := Package{T: format, B: text}
		for _, notify := range a.notifyChans {
			notify <- p
		}
	}
}

// 结束clipboard.Format类型粘贴板监听
func (a *Appclipboard) Stop(t clipboard.Format) {
	c, has := a.Cancel[t]
	if has {
		c()
		delete(a.Cancel, t)
		log.Info("取消粘贴板监控成功")
	}
}

func (a *Appclipboard) Watch() chan Package {
	ch := make(chan Package, 0)
	a.notifyChans = append(a.notifyChans, ch)
	return ch
}

func (a *Appclipboard) StopWatch(ch chan Package) {
	off := 0
	for i, notify := range a.notifyChans {
		if notify == ch {
			off = i
			close(ch)
			continue
		}
	}
	a.notifyChans = append(a.notifyChans[:off], a.notifyChans[off+1:]...)
}
