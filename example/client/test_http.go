package main

import (
	"fmt"
	"github.com/secoba/rawhttp"
	"net/http"
	"time"
)

func main() {
	options := rawhttp.DefaultOptions
	//if headers != nil && len(headers) > 0 {
	//	for k, v := range headers {
	//		options.CustomHeaders = append(options.CustomHeaders, client.Header{
	//			Key:   k,
	//			Value: v,
	//		})
	//	}
	//}
	//if len(socks5) > 0 {
	//	options.Proxy = socks5
	//} else {
	//	options.Proxy = ""
	//}
	//if timeout > 0 {
	//	options.Timeout = time.Second * time.Duration(timeout)
	//}
	start := time.Now()
	options.Timeout = time.Second * 6
	options.FollowRedirects = false
	options.ProxyDialTimeout = time.Second * 6
	//options.Proxy = "socks5://127.0.0.1:1080"
	c := rawhttp.NewClient(rawhttp.DefaultOptions)
	var (
		err error
		res *http.Response
	)
	cc, e := c.CreateConnection("https://www.csdn.com/", "", nil, options)
	if e != nil {
		fmt.Println(e)
	} else {
		go func() {
			time.Sleep(time.Second)
			fmt.Println(cc.Close())
		}()
		_, res, err = c.Get(cc, "https://docs.google.com/")

		fmt.Println(res)
		fmt.Println(err)
		fmt.Println(time.Now().Unix() - start.Unix())
	}

}
