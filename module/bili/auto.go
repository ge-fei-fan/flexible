package bili

import (
	"flexible/extra"
	log "github.com/ge-fei-fan/gefflog"
	"github.com/ncruces/zenity"
	"regexp"
)

var ch chan extra.Package

func startWatchD() {
	log.Info("startWatchD", "开始")
	defer func() {
		log.Info("startWatchD", "退出")
	}()
	ch = bili.Watch()
	for i := range ch {
		go func(extra.Package) {
			if i.T == extra.MIMEPlainText {
				url := filiterCpUrl(string(i.B))
				if url != "" {
					err := zenity.Question("检测到粘贴板有bilibili视频链接，是否开始下载?",
						zenity.Title("bilibili"))
					if err == nil {
						downloadVideo(url)
					}
				}
			}
		}(i)
	}

}
func stopWatchD() {
	bili.StopWatch(ch)
}

// 过滤粘贴板链接
func filiterCpUrl(src string) string {
	regexp, _ := regexp.Compile("https://www.bilibili.com.*")
	return regexp.FindString(src)
}
