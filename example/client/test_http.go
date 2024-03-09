package main

import (
	"fmt"
	"time"

	rawhttp "github.com/secoba/rawhttp"
)

func main() {
	raw := `POST /run HTTP/1.1
Host: your-ip:9999
Accept-Encoding: gzip, deflate
Accept: */*
Accept-Language: en
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.132 Safari/537.36
Connection: close
Content-Type: application/json
Content-Length: 365

{
  "jobId": 1,
  "executorHandler": "demoJobHandler",
  "executorParams": "demoJobHandler",
  "executorBlockStrategy": "COVER_EARLY",
  "executorTimeout": 0,
  "logId": 1,
  "logDateTime": 1586629003729,
  "glueType": "GLUE_SHELL",
  "glueSource": "start cmd",
  "glueUpdatetime": 1586699003758,
  "broadcastIndex": 0,
  "broadcastTotal": 0
}`
	options := rawhttp.DefaultOptions

	options.CustomRawBytes = []byte(raw)
	options.AutomaticHostHeader = true
	options.AutomaticContentLength = true
	urlStr := "http://127.0.0.1:9999/run"
	options.Timeout = time.Second * time.Duration(10)
	options.FollowRedirects = false
	c := rawhttp.NewClient(options)
	conn, err := c.CreateConnection(urlStr, options)
	fmt.Println(err)
	fmt.Println(conn)
	//req, res, err := c.DoRaw(conn, "", urlStr, "", nil, nil)
	//fmt.Println(string(req.RawBytes))
	//fmt.Println(hex.Dump(req.RawBytes))
	//r, _ := httputil.DumpResponse(res, true)
	//fmt.Println(string(r))
	//fmt.Println(err)
}
