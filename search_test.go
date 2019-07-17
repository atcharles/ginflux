package ginflux

import (
	"fmt"
	"log"
	"sync"
	"testing"
)

func Test_Query(t *testing.T) {
	str := `SELECT * FROM "user" ORDER BY "time" DESC LIMIT 2 OFFSET 0`
	var beans []*TModel
	_, err := testGlobalEngine.DB("db1").Query(str, &beans)
	if err != nil {
		t.Error(err)
	}
	for _, val := range beans {
		fmt.Printf("bean value is :%#v\n", val)
		b, _ := json.MarshalIndent(val, "", "  ")
		fmt.Printf("%s\n", b)
	}
}

func Test_concurrent_Query(t *testing.T) {
	ats := 200
	wg := new(sync.WaitGroup)
	for i := 0; i < ats; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			//log.Println("current num", testGlobalEngine.pool.CurrentOpen())
			//str := `DROP SERIES FROM "user"`
			str := `SELECT * FROM "user" ORDER BY "time" DESC LIMIT 2 OFFSET 0`
			_, err := testGlobalEngine.DB("db1").Query(str)
			if err != nil {
				log.Println(err)
				return
			}
		}()
	}
	wg.Wait()
	log.Println("G current num", testGlobalEngine.pool.CurrentOpen())
}

func Test_Drop(t *testing.T) {
	str := `DROP SERIES FROM "user"`
	s, err := testGlobalEngine.DB("db1").Query(str)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("result is %#v\n", s.Result)
}
