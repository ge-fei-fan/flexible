package bili

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/ge-fei-fan/gefflog"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"syscall"
)

const (
	//获取视频cid
	viewUrl = "https://api.bilibili.com/x/web-interface/view"
	//获取音视频下载链接
	playUrl = "https://api.bilibili.com/x/player/playurl"
)

// 视频信息
type VideoInfo struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Bvid   string `json:"bvid"`
		Aid    int64  `json:"aid"`
		Videos int64  `json:"videos"`
		Title  string `json:"title"`
		//子分区名称
		Tname string `json:"tname"`
		Pages []Page `json:"pages"`
	} `json:"data"`
}
type Page struct {
	Cid int64 `json:"cid"`
	//分批序号
	Page int16 `json:"page"`
	//分P标题
	Part string `json:"part"`
}

// 视频音频信息
type PlayInfo struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    struct {
		AcceptDescription []string `json:"accept_description"`
		AcceptQuality     []int16  `json:"accept_quality"`
		Dash              struct {
			Video []struct {
				Id        int16    `json:"id"`
				BaseUrl   string   `json:"baseUrl"`
				BackupUrl []string `json:"backupUrl"`
			} `json:"video"`
			Audio []struct {
				Id        int16    `json:"id"`
				BaseUrl   string   `json:"baseUrl"`
				BackupUrl []string `json:"backupUrl"`
			} `json:"audio"`
		} `json:"dash"`
	} `json:"data"`
}

func parseUrl(url string) error {
	u := filiterShareUrl(url)
	if u == "" {
		log.Err("过滤分享链接为空")
		return errors.New("过滤分享链接为空")
	}
	resp, err := http.Get(u)
	if err != nil {
		log.Err("解析分析链接失败", err)
		return err
	}
	defer resp.Body.Close()
	go downloadVideo(resp.Request.URL.String())
	return nil
}
func downloadVideo(s string) {
	u, err := url.Parse(s)
	if err != nil {
		conf.commonChan <- NewfinishNotify("地址错误，获取bvid失败", err)
		return
	}
	bvid := path.Base(u.Path)
	if bvid == "" {
		conf.commonChan <- NewfinishNotify("地址错误，获取bvid失败", errors.New("地址错误，获取bvid失败"))
		return
	}
	downloadVideoByBvid(bvid)
}
func downloadVideoByBvid(bvid string) {
	if bvid == "" {
		conf.commonChan <- NewfinishNotify("bvid为空", errors.New("bvid为空"))
		return
	}

	v, err := getCid(bvid, 0)
	if err != nil {
		conf.commonChan <- NewfinishNotify("获取cid失败", err)
		return
	}
	if v.Message != "0" {
		errStr := fmt.Sprintf("%s 获取cid失败: %s", bvid, v.Message)
		conf.commonChan <- NewfinishNotify(v.Message, errors.New(errStr))
		return
	}
	v.start()
}

// 获取cid
func getCid(bvid string, aid int64) (*VideoInfo, error) {
	req, err := http.NewRequest("GET", viewUrl, nil)
	if err != nil {
		return nil, err
	}
	//设置参数
	params := req.URL.Query()
	params.Add("bvid", bvid)
	params.Add("aid", strconv.Itoa(int(aid)))
	req.URL.RawQuery = params.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var v VideoInfo
	err = json.Unmarshal(body, &v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (vi VideoInfo) start() {
	//视频存储路径
	var fileP string
	fileP = conf.Path
	if vi.Data.Videos > 1 {
		title := FiliterFilename(vi.Data.Title)
		fileP = filepath.Join(conf.Path, title)
	}
	for i, _ := range vi.Data.Pages {
		go vi.saveVideo(fileP, i)
	}
}

func (vi VideoInfo) saveVideo(fileP string, i int) {
	part := FiliterFilename(vi.Data.Pages[i].Part)
	//只有一个视频的话，视频名字就是title，多p视频名称是part名称
	if vi.Data.Videos == 1 {
		part = FiliterFilename(vi.Data.Title)
	}
	if isExist(filepath.Join(fileP, part+".mp4")) {
		conf.commonChan <- NewfinishNotify(part+"【已存在】", os.ErrExist)
		conf.commonChan <- newVideoInfo(part, "", "已存在")
		return
	}
	p, err := getVideoInfo(vi.Data.Bvid, 0, vi.Data.Pages[i].Cid)
	if err != nil {
		msg := fmt.Sprintf("%s获取视频信息失败", part)
		conf.commonChan <- NewfinishNotify(msg, err)
		conf.commonChan <- newVideoInfo(part, "", "请求视频信息失败")
		return
	}
	if p.Message != "0" {
		errStr := fmt.Sprintf("%s 获取视频信息失败: %s", part, p.Message)
		conf.commonChan <- NewfinishNotify(p.Message, errors.New(errStr))
		conf.commonChan <- newVideoInfo(part, "", "获取视频信息失败")
		return
	}

	//下载视频
	log.Info(part + "视频开始下载，清晰度为：" + getVideoQuality(p.Data.Dash.Video[0].Id))
	conf.commonChan <- newVideoInfo(part, getVideoQuality(p.Data.Dash.Video[0].Id), "正在下载视频")
	err = bilidownload(p.Data.Dash.Video[0].BaseUrl, fileP, part+".video")
	if err != nil {
		err = bilidownload(p.Data.Dash.Video[0].BackupUrl[0], fileP, part+".video")
		if err != nil {
			conf.commonChan <- NewfinishNotify(part+"下载视频失败", err)
			conf.commonChan <- newVideoInfo(part, getVideoQuality(p.Data.Dash.Video[0].Id), "下载视频失败")
			return
		}
	}
	log.Info(part + "视频下载完成")
	//下载音频
	log.Info(part + "音频开始下载")
	conf.commonChan <- newVideoInfo(part, getVideoQuality(p.Data.Dash.Video[0].Id), "正在下载音频")
	err = bilidownload(p.Data.Dash.Audio[0].BaseUrl, fileP, part+".audio")
	if err != nil {
		err = bilidownload(p.Data.Dash.Audio[0].BackupUrl[0], fileP, part+".audio")
		if err != nil {
			conf.commonChan <- NewfinishNotify(part+"下载音频失败", err)
			conf.commonChan <- newVideoInfo(part, getVideoQuality(p.Data.Dash.Video[0].Id), "下载音频失败")
			return
		}
	}
	log.Info(part + "音频下载完成")
	//合并音视频
	log.Info(part + "开始合并音视频")
	conf.commonChan <- newVideoInfo(part, getVideoQuality(p.Data.Dash.Video[0].Id), "正在合并音视频")
	err = mergeVideo(fileP, part)
	if err != nil {
		err = mergeVideo(fileP, part)
		if err != nil {
			conf.commonChan <- NewfinishNotify(part+"合并音视频失败", err)
			conf.commonChan <- newVideoInfo(part, getVideoQuality(p.Data.Dash.Video[0].Id), "合并音视频失败")
			return
		}
	}
	log.Info(part + "合并音视频完成")
	conf.commonChan <- NewfinishNotify(part+"【下载完成】", nil)
	conf.commonChan <- newVideoInfo(part, getVideoQuality(p.Data.Dash.Video[0].Id), "下载完成")
}

func getVideoInfo(bvid string, aid, cid int64) (*PlayInfo, error) {
	req, err := http.NewRequest("GET", playUrl, nil)
	if err != nil {
		return nil, err
	}
	//设置参数
	params := req.URL.Query()
	if bvid != "" {
		params.Add("bvid", bvid)
	} else if aid != 0 {
		params.Add("aid", strconv.Itoa(int(aid)))
	} else if bvid == "" && aid == 0 {
		return nil, errors.New("bvid和aid都为空")
	}
	if cid == 0 {
		return nil, errors.New("cid为0")
	}
	params.Add("cid", strconv.Itoa(int(cid)))
	params.Add("qn", "0")
	params.Add("fnver", "0")
	params.Add("fnval", "208")
	params.Add("fourk", "1")
	req.URL.RawQuery = params.Encode()

	cookie := fmt.Sprintf("SESSDATA=%s", conf.Sessdata)
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36")
	req.Header.Set("cookie", cookie)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var p PlayInfo
	err = json.Unmarshal(body, &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
func getVideoQuality(q int16) string {
	switch q {
	case 6:
		return "240P 极速"
	case 16:
		return "240P 极速"
	case 32:
		return "480P 清晰"
	case 64:
		return "720P 高清"
	case 72:
		return "720P60 高帧率"
	case 80:
		return "1080P 高清"
	case 112:
		return "1080P+ 高码率"
	case 116:
		return "1080P60 高帧率"
	case 120:
		return "4K 超清"
	case 125:
		return "HDR 真彩色"
	case 126:
		return "杜比视界"
	case 127:
		return "8K 超高清"
	default:
		return "未知"
	}
}
func bilidownload(url, path, filename string) error {
	err := os.MkdirAll(path, 0666)
	if err != nil {
		return err
	}
	n := filepath.Join(path, filename)
	if isExist(n) {
		log.Info(filename, ":已存在")
		return nil
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("referer", "https://www.bilibili.com")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.Create(n)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}
	return nil

}
func mergeVideo(path, name string) error {
	videoPath, _ := filepath.Abs(filepath.Join(path, name+".video"))
	audioPath, _ := filepath.Abs(filepath.Join(path, name+".audio"))
	outPath, _ := filepath.Abs(filepath.Join(path, name+".mp4"))
	cmdArguments := []string{"-i", videoPath, "-i", audioPath,
		"-c:v", "copy", "-c:a", "copy", "-f", "mp4", outPath}

	cmd := exec.Command("ffmpeg", cmdArguments...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	//删除临时文件
	err = os.Remove(videoPath)
	if err != nil {
		return err
	}
	err = os.Remove(audioPath)
	if err != nil {
		return err
	}

	return nil
}
func isExist(filepath string) bool {
	_, err := os.Stat(filepath)
	return err == nil || os.IsExist(err)
}

// 过滤创建文件夹时的非法字符
func FiliterFilename(src string) string {
	regexp, _ := regexp.Compile(`[/ : * ? " < > | \\]`)
	return regexp.ReplaceAllString(src, " ")
}

// 过滤分享链接
func filiterShareUrl(src string) string {
	regexp, _ := regexp.Compile("https://.*")
	return regexp.FindString(src)
}
