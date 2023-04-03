package main

import (
	"flexible/core"
	"flexible/module/bili"
	"flexible/module/clipboard"
	"fmt"
	log "github.com/ge-fei-fan/gefflog"
	"runtime"
)

const (
	defaultStackSize = 4096
)

func getCurrentGoroutineStack() string {
	var buf [defaultStackSize]byte
	n := runtime.Stack(buf[:], false)
	return string(buf[:n])
}
func main() {
	defer func() {
		log.Info("主程序退出")
		if err := recover(); err != nil {
			log.Err(fmt.Sprintf("[panic] err: %v\nstack: %s\n", err, getCurrentGoroutineStack()))
		}
	}()
	var a *core.App
	var err error
	//core.RegisterTab(bili.ExportTab())
	a, err = core.NewApp("flexible", "v1.0.0", nil)
	if err != nil {
		log.Err(err)
		return
	}
	a.Flexwiondow()
	a.RegisterModule(
		bili.ExportModule(a.Win()),
		clipboard.ExportModule(),
	)
	err = a.Start()
	if err != nil {
		log.Err(err)
		return
	}
	//a.RegisterTab(
	//	bili.ExportTab(),
	//)
	a.WinStart()

}
