package ginflux

import (
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
)

func Test_d1(t *testing.T) {
	ch := make(chan int, 100)
	for {
		select {
		case a := <-ch:
			fmt.Println(a)
		case <-time.After(time.Second * 2):
			fmt.Println("push to channel failed")
			return
		}
	}
}

func TestNewOPool(t *testing.T) {
	pl := NewOPool(Options{
		HttpConf: client.HTTPConfig{
			Addr:               "http://127.0.0.1:8086",
			Username:           "",
			Password:           "",
			UserAgent:          "",
			Timeout:            0,
			InsecureSkipVerify: false,
			TLSConfig:          nil,
			Proxy:              nil,
		},
		MinOpen: 10,
		MaxOpen: 20,
	})

	wg := sync.WaitGroup{}

	a := 0
	for a < 100 {
		a++
		wg.Add(1)

		go func() {
			defer wg.Done()

			fmt.Printf("current open clients count is %d\n", pl.CurrentOpen())
			oc, err := pl.Acquire()
			if err != nil {
				log.Println(err)
				return
			}
			defer oc.Release()
			a, _, _ := oc.GetInfluxClient().Ping(time.Second)
			fmt.Printf("网络延时:%s\n", a.String())
		}()
	}
	wg.Wait()
	fmt.Printf("current open clients count is %d\n", pl.CurrentOpen())
}
