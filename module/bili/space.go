package bili

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/ge-fei-fan/gefflog"
	"github.com/google/go-querystring/query"
	"io"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"
)

var getViedoInterval = 300 * time.Second

type Video struct {
	Aid          int    `json:"aid"`            // 稿件avid
	Author       string `json:"author"`         // 视频UP主，不一定为目标用户（合作视频）
	Bvid         string `json:"bvid"`           // 稿件bvid
	Comment      int    `json:"comment"`        // 视频评论数
	Copyright    string `json:"copyright"`      // 空，作用尚不明确
	Created      int64  `json:"created"`        // 投稿时间戳
	Description  string `json:"description"`    // 视频简介
	HideClick    bool   `json:"hide_click"`     // 固定值false，作用尚不明确
	IsPay        int    `json:"is_pay"`         // 固定值0，作用尚不明确
	IsUnionVideo int    `json:"is_union_video"` // 是否为合作视频，0：否，1：是
	Length       string `json:"length"`         // 视频长度，MM:SS
	Mid          int    `json:"mid"`            // 视频UP主mid，不一定为目标用户（合作视频）
	Pic          string `json:"pic"`            // 视频封面
	Play         int    `json:"play"`           // 视频播放次数
	Review       int    `json:"review"`         // 固定值0，作用尚不明确
	Subtitle     string `json:"subtitle"`       // 固定值空，作用尚不明确
	Title        string `json:"title"`          // 视频标题
	Typeid       int    `json:"typeid"`         // 视频分区tid
	VideoReview  int    `json:"video_review"`   // 视频弹幕数
}

type GetUserVideosData struct {
	List struct { // 列表信息
		Tlist map[int]struct { // 投稿视频分区索引
			Count int    `json:"count"` // 投稿至该分区的视频数
			Name  string `json:"name"`  // 该分区名称
			Tid   int    `json:"tid"`   // 该分区tid
		} `json:"tlist"`
		Vlist []Video `json:"vlist"` // 投稿视频列表
	} `json:"list"`
	Page struct { // 页面信息
		Count int `json:"count"` // 总计稿件数
		Pn    int `json:"pn"`    // 当前页码
		Ps    int `json:"ps"`    // 每页项数
	} `json:"page"`
	EpisodicButton struct { // “播放全部“按钮
		Text string `json:"text"` // 按钮文字
		Uri  string `json:"uri"`  // 全部播放页url
	} `json:"episodic_button"`
}

// 查询用户投稿视频响应
type GetUserVideosResult struct {
	Code    int64             `json:"code"`
	Message string            `json:"message"`
	Data    GetUserVideosData `json:"data"`
}

type arcParams struct {
	Mid     string `url:"mid"`
	Order   string `url:"order"`
	Tid     int    `url:"tid,omitempty"`
	Keyword string `url:"keyword,omitempty"`
	Pn      int    `url:"pn"`
	Ps      int    `url:"ps"`
}

const (
	OrderPubDate string = "pubdate"
	OrderClick   string = "click"
	OrderStow    string = "stow"
	ARCURL       string = "https://api.bilibili.com/x/space/wbi/arc/search"
)

func NewArcParams(tid, ps, pn int, mid, order, keyword string) *arcParams {
	return &arcParams{
		Mid:     mid,
		Order:   order,
		Tid:     tid,
		Keyword: keyword,
		Pn:      pn,
		Ps:      ps,
	}
}
func NewDefaultArcParams(mid string) *arcParams {
	return &arcParams{
		Mid:   mid,
		Order: OrderPubDate,
		Ps:    10,
		Pn:    1,
	}
}

type User struct {
	Mid       string
	lastVideo map[string]Video
}

func NewUser(mid string) *User {
	return &User{
		Mid:       mid,
		lastVideo: make(map[string]Video),
	}
}
func (u *User) GetUserVideo(input any) (*GetUserVideosResult, error) {
	req, err := http.NewRequest("GET", ARCURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36")
	if input != nil {
		vals, _ := query.Values(input)
		params := vals.Encode()
		if params != "" {
			req.URL.RawQuery = params
		}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var guvr GetUserVideosResult
	err = json.Unmarshal(body, &guvr)
	if err != nil {
		return nil, err
	}
	return &guvr, err
}

// 过滤最新视频
func (u *User) CheckVideo() {
	//一次只请求最新的10个投稿
	userVideoInfos, err := u.GetUserVideo(NewDefaultArcParams(u.Mid))
	if err != nil {
		log.Err("GetUserVideo err:", err)
		return
	}
	if userVideoInfos.Message != "0" {
		log.Err("GetUserVideo err:", userVideoInfos.Message)
		return
	}
	start, end := GetDateparse()
	for _, v := range userVideoInfos.Data.List.Vlist {
		if v.Created > start && v.Created < end {
			//今天的视频，查看是否已经存在
			_, has := u.lastVideo[v.Bvid]
			if !has { //视频不存在
				//下载视频
				fmt.Println(v.Description)
				go downloadVideoByBvid(v.Bvid)
				//bvid记录一下
				u.lastVideo[v.Bvid] = v
			}
		} else if v.Created < start {
			break
		}
	}
	//把今天以前的视频都删除记录
	for k, v := range u.lastVideo {
		if v.Created < start {
			fmt.Println(v.Created)
			delete(u.lastVideo, k)
		}
	}
}

type UserHub struct {
	sync.RWMutex
	configName string
	Users      map[string]*User
	done       chan struct{}
}

func NewUserHub() *UserHub {
	return &UserHub{
		configName: "bili.json",
		Users:      make(map[string]*User),
		done:       make(chan struct{}),
	}
}
func (uh *UserHub) Start() {
	log.Info("进入bili最新视频监控")
	videoTicker := time.NewTicker(getViedoInterval)
	defer func() {
		videoTicker.Stop()
		log.Info("退出bili监控")
	}()
	for {
		select {
		case <-videoTicker.C:
			VideoSchedule(uh.Users)
		case <-uh.done:
			return
		}

	}
}
func VideoSchedule(m map[string]*User) {
	for _, user := range m {
		//查询最新视频，并比对
		go user.CheckVideo()
	}
}
func (uh *UserHub) DelUser(mid string) error {
	_, has := uh.Users[mid]
	if has {
		delete(uh.Users, mid)
		err := uh.save()
		if err != nil {
			return err
		}
	}
	return nil
}
func (uh *UserHub) AddUser(mid string) error {
	uh.Lock()
	defer uh.Unlock()
	_, has := uh.Users[mid]
	if !has {
		uh.Users[mid] = NewUser(mid)
		err := uh.save()
		if err != nil {
			return err
		}
	}
	return nil
}

// 将userhub数据写入conf
func (uh *UserHub) save() error {
	marshal, err := json.Marshal(uh)
	if err != nil {
		return err
	}
	return os.WriteFile(uh.configName, marshal, 0644)
}

// 读取conf
func (uh *UserHub) load() error {
	if uh.configName == "" {
		return errors.New("配置文件名为空")
	}

	file, err := os.ReadFile(uh.configName)
	if err != nil {
		return err
	}
	return json.Unmarshal(file, uh)
}

// 过滤mid
func FiliterMid(src string) string {
	regexp, _ := regexp.Compile("\\d+$")
	return regexp.FindString(src)
}

// 日期当天0点时间戳和23时59分时间戳
func GetDateparse() (int64, int64) {
	//获取当前时区
	loc, _ := time.LoadLocation("Local")
	date := time.Now().Format("2006-01-02")
	//日期当天0点时间戳(拼接字符串)
	startDate := date + "_00:00:00"
	startTime, _ := time.ParseInLocation("2006-01-02_15:04:05", startDate, loc)

	//日期当天23时59分时间戳
	endDate := date + "_23:59:59"
	end, _ := time.ParseInLocation("2006-01-02_15:04:05", endDate, loc)

	//返回当天0点和23点59分的时间戳
	return startTime.Unix(), end.Unix()
}
