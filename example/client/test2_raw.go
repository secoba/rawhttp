package main

import (
	"fmt"
	rawhttp "github.com/secoba/rawhttp"
	"github.com/secoba/rawhttp/client"
	"io"
	"net/http"
	"time"
)

func DoRaw(c *rawhttp.Client, conn rawhttp.Conn, urlStr string, rawBuffer []byte) (*client.Request, *http.Response, error) {
	return c.DoRaw(conn, "", urlStr, "", nil, nil, rawBuffer)
}

func CreateConn(urlStr string, headers map[string]string, socks5 string, timeout int, redirect bool, raw []byte, fixHost, fixLength bool) (rawhttp.Conn, *rawhttp.Client, error) {
	options := &rawhttp.Options{
		Timeout:                30 * time.Second,
		FollowRedirects:        true,
		MaxRedirects:           10,
		AutomaticHostHeader:    true,
		AutomaticContentLength: true,
	}

	options.CustomRawBytes = raw
	options.AutomaticHostHeader = fixHost
	options.AutomaticContentLength = fixLength

	if headers != nil && len(headers) > 0 {
		for k, v := range headers {
			options.CustomHeaders = append(options.CustomHeaders, client.Header{
				Key:   k,
				Value: v,
			})
		}
	}
	if len(socks5) > 0 {
		options.Proxy = socks5
	} else {
		options.Proxy = ""
	}
	if timeout > 0 {
		options.Timeout = time.Second * time.Duration(timeout)
	}
	// options.ForceReadAllBody = true
	options.FollowRedirects = redirect
	c := rawhttp.NewClient(options)
	conn, err := c.CreateConnection(urlStr, options)
	return conn, c, err
}

func main() {
	pack := `GET /phoenix/web/v1/get-latest-comment HTTP/1.1
Host: blog.csdn.net
Accept: application/json, text/javascript, */*; q=0.01
Accept-Language: zh-CN,zh;q=0.9
Content-Type: application/x-www-form-urlencoded; charset=utf-8
Cookie: uuid_tt_dd=10_37080423800-1695025116718-293300; c_segment=4; SESSION=0670500b-01c2-4513-aa2e-fd60d31fa1f2; UserName=u014196376; UserInfo=01f8fa011d6d428e8b071a0e6c11a081; UserToken=01f8fa011d6d428e8b071a0e6c11a081; UserNick=eror234; AU=4CA; UN=u014196376; BT=1696917721258; p_uid=U010000; Hm_up_6bcd52f51e9b3dce32bec4a3997715ac=%7B%22islogin%22%3A%7B%22value%22%3A%221%22%2C%22scope%22%3A1%7D%2C%22isonline%22%3A%7B%22value%22%3A%221%22%2C%22scope%22%3A1%7D%2C%22isvip%22%3A%7B%22value%22%3A%220%22%2C%22scope%22%3A1%7D%2C%22uid_%22%3A%7B%22value%22%3A%22u014196376%22%2C%22scope%22%3A1%7D%7D; c_pref=https%3A//www.bing.com/; c_ref=https%3A//blog.csdn.net/lz710117239/article/details/73391703; c_segment=4; log_Id_pv=99; log_Id_view=3011; log_Id_click=43; __gpi=UID=00000c49b0b00a9d:T=1695025117:RT=1699856292:S=ALNI_MaWGeeMQAkYjQ3ue7-RvGs961gWtQ; _ga_7W1N0GEY1P=deleted; Hm_lvt_e5ef47b9f471504959267fd614d579cd=1699866652,1701847456; Hm_lpvt_e5ef47b9f471504959267fd614d579cd=1701916718; ssxmod_itna=Qqfx0DnDRD9DgjKDXiG7ma0=IfLRRnfi70UfDBL+i4iNDnD8x7YDvm+rD02ib2i2x1Oi17bvsfQ0TpwmtQBDPwxe577oKoD84i7DKqibDCqD1D3qDkbnoxiicDCeDIDWeDiDG+kD02NZOD0R10r5zk1FO1p91DYcOwxDOvxGCa4GteS5Gy6aPyxDWW40kq4iawx0Cl19HDmKDIgSOrz5DFaOfUtS7Dm432=SADCKvmZPDUAHsV1FX/GD33GGeZAikeARDaAqeKQDqWVhte0DxHZ=KqEqaYmGFZChxGFrN2dD; ssxmod_itna2=Qqfx0DnDRD9DgjKDXiG7ma0=IfLRRnfi70fD8MCDxxGNDC8DFESZazcciKAp38Pq86kv45qp6whi+Dkgdo63aiUk8a8q7GeSbj=WC407K8x7=D+OiD==; c_dl_prid=-; c_dl_rid=1709012761679_499984; c_dl_fref=https://blog.csdn.net/rui754220732/article/details/120748657; c_dl_fpage=/download/navy102019/2878919; c_dl_um=-; log_Id_click=44; log_Id_pv=100; log_Id_view=3012; dc_sid=46407ebe960f8e788ef46c2c9676a57d; Hm_lvt_6bcd52f51e9b3dce32bec4a3997715ac=1709101320; _ga=GA1.2.1931716709.1697096726; _ga_7W1N0GEY1P=GS1.1.1709484724.30.1.1709485972.60.0.0; FCNEC=%5B%5B%22AKsRol_lfIzW_W26CULwBwApOoIcvmZFA7liM0MTDeymriMx4LS3hyQc-_bfw0yvnDEsKVWAq48KJ99l04x_AhjbK6hbv7U2cQuvZ8i3EBCjsInOjGY7yQ85poVIzTxK_Dcssz_Hz9j7hlRfYlCBp4VwZ26Ze2kwHg%3D%3D%22%5D%5D; https_waf_cookie=2a5fc57c-6f9f-4d18da966273f012ca3fcabd9322bc6a7387; firstDie=1; is_advert=1; __gads=ID=53f4f473210d29e7-22ceec2edee300a5:T=1695025117:RT=1710128791:S=ALNI_MZL7hVV2Hv6ksmZXTSl69uKQ0qofA; __eoi=ID=fc9dd111ffcc462f:T=1706685335:RT=1710128791:S=AA-Afja15MDYsqd2M2yxzPhjtF-s; dc_session_id=10_1710146895258.562661; c_pref=default; c_ref=default; c_first_ref=default; c_first_page=https%3A//blog.csdn.net/xiaoweite1/article/details/135258123; creativeSetApiNew=%7B%22toolbarImg%22%3A%22https%3A//img-home.csdnimg.cn/images/20230921102607.png%22%2C%22publishSuccessImg%22%3A%22https%3A//img-home.csdnimg.cn/images/20230920034826.png%22%2C%22articleNum%22%3A59%2C%22type%22%3A2%2C%22oldUser%22%3Atrue%2C%22useSeven%22%3Afalse%2C%22oldFullVersion%22%3Atrue%2C%22userName%22%3A%22u014196376%22%7D; Hm_lpvt_6bcd52f51e9b3dce32bec4a3997715ac=1710146902; waf_captcha_marker=46f7b46f24a68e4efd4310cbf7d0e3d0e4fe743b4ed461f255ef640a79d572d0; c_dsid=11_1710146924251.606356; c_page_id=default; dc_tos=sa6dt8; creative_btn_mp=3
Referer: https://blog.csdn.net/xiaoweite1/article/details/135258123
Sec-Ch-Ua: "Chromium";v="122", "Not(A:Brand";v="24", "Google Chrome";v="122"
Sec-Ch-Ua-Mobile: ?0
Sec-Ch-Ua-Platform: "macOS"
Sec-Fetch-Dest: empty
Sec-Fetch-Mode: cors
Sec-Fetch-Site: same-origin
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36
X-Requested-With: XMLHttpRequest

`
	urlStr := "https://blog.csdn.net:443/phoenix/web/v1/get-latest-comment"
	cnn, cli, err := CreateConn(urlStr, nil, "", 10, false, []byte(pack), false, true)
	fmt.Println(err)
	req, res, err := DoRaw(cli, cnn, urlStr, []byte(pack))
	fmt.Println(string(req.RawBytes))
	fmt.Println(err)
	buf, _ := io.ReadAll(res.Body)
	fmt.Println(string(buf))

}
