package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/remeh/sizedwaitgroup"
	rawhttp "github.com/secoba/rawhttp"
)

func main() {
	swg := sizedwaitgroup.New(25)
	pipeOptions := rawhttp.DefaultPipelineOptions
	pipeOptions.Host = "scanme.sh"
	pipeOptions.MaxConnections = 1
	pipeclient := rawhttp.NewPipelineClient(context.Background(), pipeOptions)
	for i := 0; i < 100; i++ {
		swg.Add()
		go func(swg *sizedwaitgroup.SizedWaitGroup) {
			defer swg.Done()
			req, err := http.NewRequest("GET", "http://scanme.sh/headers", nil)
			if err != nil {
				log.Printf("Error sending request to API endpoint. %+v", err)
				return
			}
			req.Host = "scanme.sh"
			req.Header.Set("Host", "scanme.sh")

			//pipeclient.DoRaw("", "", "", nil, nil, nil)
			//pipeclient.DoRaw()

			_, resp, err := pipeclient.Do(req)
			if err != nil {
				log.Printf("Error sending request to API endpoint. %+v", err)
				return
			}
			fmt.Println(i)
			//log.Printf("%+v\n", resp)
			_ = resp
		}(&swg)
	}

	swg.Wait()
}
