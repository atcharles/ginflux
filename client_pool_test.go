package ginflux

import (
	"fmt"
	"sync"
	"testing"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
)

func TestNewOPool(t *testing.T) {
	pl := NewOPool(Options{
		httpConf: client.HTTPConfig{
			Addr:               "http://127.0.0.1:8086",
			Username:           "",
			Password:           "",
			UserAgent:          "",
			Timeout:            0,
			InsecureSkipVerify: false,
			TLSConfig:          nil,
			Proxy:              nil,
		},
		minOpen: 10,
		maxOpen: 300,
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
				panic(err)
			}
			defer oc.Release()
			a, _, _ := oc.GetInfluxClient().Ping(time.Second)
			fmt.Printf("网络延时:%s\n", a.String())
		}()
	}
	wg.Wait()
	fmt.Printf("current open clients count is %d\n", pl.CurrentOpen())
}
