package main

import (
	"flag"
	"fmt"
	"net/http/httputil"

	"github.com/secoba/rawhttp"
)

var (
	url   string
	short bool
)

func main() {
	flag.StringVar(&url, "url", "https://scanme.sh", "URL to fetch")
	flag.BoolVar(&short, "short", false, "Skip printing http response body")
	flag.Parse()

	client := rawhttp.NewClient(rawhttp.DefaultOptions)
	cc, _ := client.CreateConnection("", "", nil, nil)
	_, resp, err := client.Get(cc, url)
	if err != nil {
		panic(err)
	}
	bin, err := httputil.DumpResponse(resp, !short)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bin))
}
