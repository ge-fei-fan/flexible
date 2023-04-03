package bili

import (
	"encoding/json"
	"errors"
	"fmt"
	"fyne.io/systray"
	log "github.com/ge-fei-fan/gefflog"
	"github.com/lxn/walk"
	qrcode "github.com/skip2/go-qrcode"
	"io"
	"net/http"
	"time"
)

type GetQrResp struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    QRCode `json:"data"`
}

type QRCode struct {
	Url       string `json:"url"`        // 二维码内容url
	QrcodeKey string `json:"qrcode_key"` // 扫码登录秘钥
}
type LoginQrResp struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    struct {
		RefreshToken string `json:"refresh_token"`
		Code         int64  `json:"code"`
		Message      string `json:"message"`
	} `json:"data"`
}

var GETQRURL = "https://passport.bilibili.com/x/passport-login/web/qrcode/generate"
var CHECKQR = "https://passport.bilibili.com/x/passport-login/web/qrcode/poll?qrcode_key=%s"

// GetQRCode 申请二维码URL及扫码密钥
// 保存二维码到temp.png
func GetQRCode() (*GetQrResp, error) {

	resp, err := http.Get(GETQRURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var gqr GetQrResp
	err = json.Unmarshal(body, &gqr)
	if err != nil {
		return nil, err
	}
	if gqr.Message != "0" {
		return nil, errors.New(gqr.Message)
	}
	if gqr.Data.Url == "" {
		return nil, errors.New("请求二维码链接为空")
	}
	err = qrcode.WriteFile(gqr.Data.Url, qrcode.Medium, 256, "qr.png")
	if err != nil {
		return nil, err
	}
	return &gqr, nil
}

func CheckQrLogin(gqr *GetQrResp, done chan struct{}) error {
	if gqr == nil {
		walk.MsgBox(nil, "错误", "登录二维码获取失败", walk.MsgBoxIconInformation)
		return errors.New("传入的*GetQrResp为空")
	}
	if gqr.Data.QrcodeKey == "" {
		walk.MsgBox(nil, "错误", "登录二维码获取失败", walk.MsgBoxIconInformation)
		return errors.New("传入QrcodeKey为空")
	}
	u := fmt.Sprintf(CHECKQR, gqr.Data.QrcodeKey)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36")
	go func(d chan struct{}) {
		timer := time.NewTicker(time.Second)
		defer func() {
			timer.Stop()
		}()
		for {
			select {
			case <-timer.C:
				exit := func(r *http.Request) bool {
					resp, err := http.DefaultClient.Do(r)
					if err != nil {
						log.Err("get CHECKQR err:", err)
						return false
					}
					defer resp.Body.Close()
					body, _ := io.ReadAll(resp.Body)

					var lqr LoginQrResp
					json.Unmarshal(body, &lqr)
					if lqr.Message != "0" {
						log.Err("get CHECKQR err:", err)
						return false
					}
					switch lqr.Data.Code {
					case 0: //登录成功
						//设置sessdata refresh_token
						conf.RefreshToken = lqr.Data.RefreshToken
						for _, ck := range resp.Cookies() {
							if ck.Name == "SESSDATA" {
								conf.Sessdata = ck.Value
								break
							}
						}
						bili.SaveConfig(conf)
						return true
					case 86038: //二维码已失效
						//不用通知应该也没事
						//fmt.Println("二维码已失效")
						return true
					case 86090: //二维码已扫码未确认
						//fmt.Println("二维码已扫码未确认")
						return false
					case 86101: //未扫码
						//fmt.Println("未扫码")
						return false
					default:
						return false
					}
				}(req)
				if exit {
					biliTab.userInfo.CheckLoginHandle()       //获取一下个人信息
					biliTab.qrLoginDlg.Close(walk.DlgCmdNone) //把二维码弹窗关闭
					return
				}
			case <-d:
				//fmt.Println("关闭窗口退出")
				return
			}
		}
	}(done)

	return nil

}

// 检查account是否能获取到
type account struct {
	Code    int16  `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Mid    int64  `json:"mid"`
		Uname  string `json:"uname"`
		UserId string `json:"userid"`
		Rank   string `json:"rank"`
	} `json:"data"`
}

const accountUrl = "https://api.bilibili.com/x/member/web/account"

func checkLogin(item *systray.MenuItem) {
	req, err := http.NewRequest(http.MethodGet, accountUrl, nil)
	if err != nil {
		log.Err(err)
		return
	}
	cookie := fmt.Sprintf("SESSDATA=%s", conf.Sessdata)
	req.Header.Set("cookie", cookie)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Err(err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Err(err)
		return
	}
	var a account
	err = json.Unmarshal(body, &a)
	if err != nil {
		log.Err(err)
		return
	}
	if a.Code == 0 {
		item.SetTitle(a.Data.Uname + ":已登陆")
		biliTab.userInfo.Name = a.Data.Uname
		biliTab.exitPB.SetVisible(true)
		biliTab.nameTE.SetText(fmt.Sprintf("%s: 已登陆", biliTab.userInfo.Name))
	} else if a.Code == -101 {
		item.SetTitle(a.Message)
		biliTab.userInfo.Name = a.Message
		biliTab.nameTE.SetText(biliTab.userInfo.Name)
		biliTab.exitPB.SetVisible(false)
	}

}
