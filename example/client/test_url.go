package main

import (
	"fmt"
	urlutil "github.com/projectdiscovery/utils/url"
)

func main() {
	urlStr := "http://bai.com/setup/setup-s/%u002e%u002e/%u002e%u002e/user-groups.jsp"
	u, e := urlutil.ParseURL(urlStr, true)
	fmt.Println(e)
	fmt.Println(u.Host)
	fmt.Println(u.Path)
	fmt.Println(u.URL.Host)
	fmt.Println(u.URL.String())
	fmt.Println(u.URL.Path)
}
