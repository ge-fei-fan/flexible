package bili

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ncruces/zenity"
	"golang.design/x/clipboard"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"path"
	"testing"
	"time"
)

func TestB(t *testing.T) {
	client := new(http.Client) //初始化一个http客户端结构体
	url := "https://api.bilibili.com/x/player/playurl?bvid=BV1p54y1w7Ti&cid=1014102446&qn=0&fnver=0&fnval=208&fourk=1"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36")
	req.Header.Set("cookie", "SESSDATA=7eece3eb%2C1682341057%2C427b5%2Aa2")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	var v PlayInfo
	body, _ := io.ReadAll(resp.Body)
	//fmt.Println(string(body))
	err = json.Unmarshal(body, &v)
	fmt.Println(err)

}

func TestPath(t *testing.T) {
	u, _ := url.Parse("https://www.bilibili.com/video/BV14s4y1a7Mb/?spm_id_from=333.1007.tianma.1-1-1.click")
	bvid := path.Base(u.Path)
	fmt.Println(bvid)
	fmt.Println(errors.New(bvid + ":" + "test"))
}

func TestP(t *testing.T) {
	err := exec.Command("cmd", "/c", "start", "info.log").Run()
	if err != nil {
		fmt.Println(err)
	}
	time.Sleep(10 * time.Second)
	//zenity.SelectFile()
}

type test struct {
	i []int
}

func TestTime(t *testing.T) {
	//url := "【【顔美】摆尾 Freaky 我们肉肉学姐这么性感的嘛？-哔哩哔哩】 https://b23.tv/gRSOPdz"
	//fmt.Println(filiterShareUrl(url))
	var ctx context.Context
	ctx, cancel := context.WithCancel(context.Background())

	var ch <-chan []byte
	ch = clipboard.Watch(ctx, clipboard.FmtText)

	go func() {
		defer fmt.Println("go退出")
		for i := range ch {
			fmt.Println(string(i))
		}

	}()
	time.Sleep(5 * time.Second)
	cancel()
	time.Sleep(5 * time.Second)
	fmt.Println("cancal again")
	cancel()
	//fmt.Println("done")
	time.Sleep(5 * time.Second)
	//fmt.Println("退出")
	//fmt.Println("again")
	//ctx, cancel = context.WithCancel(context.Background())
	//ch = clipboard.Watch(ctx, clipboard.FmtText)
	//go func() {
	//	defer fmt.Println("go退出")
	//	for i := range ch {
	//		fmt.Println(string(i))
	//	}
	//
	//}()
	//time.Sleep(10 * time.Second)
	//cancel()
}

func TestMap(t *testing.T) {
	//var a map[int]int
	//a = make(map[int]int)
	//a[1] = 1
	err := zenity.Question("Are you sure you want to proceed?",
		zenity.Title("Question"),
		zenity.QuestionIcon)
	if err != nil {
		fmt.Println(err)
	}
}

func TestCheck(t *testing.T) {
	//u := NewUser("1575111")
	//u.CheckVideo()
	fmt.Println(newVideoInfo("test", "4k", "123"))
}
