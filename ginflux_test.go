package ginflux

import (
	"fmt"
	"testing"
)

var (
	testGlobalEngine *Engine
)

func init() {
	eg, err := NewEngine(DefOpts)
	if err != nil {
		panic(err)
	}
	testGlobalEngine = eg
}

func TestEngine_SyncDB(t *testing.T) {
	err := testGlobalEngine.SyncDB([]RetentionPolicy{
		{
			DBName:   "db1",
			RPName:   "rp_3d",
			Duration: "3d",
			Default:  false,
		},
		{
			DBName:   "db2",
			RPName:   "rp_6d",
			Duration: "6d",
			Default:  false,
		},
		{
			DBName:        "db1",
			RPName:        "rp_30d",
			Duration:      "30d",
			Replication:   10,
			ShardDuration: "10d",
			Default:       true,
		},
	}...)
	if err != nil {
		t.Error(err)
	}
}

func TestNewEngine(t *testing.T) {
	eg, err := NewEngine(DefOpts)
	fmt.Printf("%#v %#v\n", eg, err)
}

func mustGetClient() *OClient {
	client, err := testGlobalEngine.Acquire()
	if err != nil {
		panic(err)
	}
	defer client.Release()
	return client
}

func TestEngine_Acquire(t *testing.T) {
	client := mustGetClient()
	fmt.Printf("%#v\n", client)
}
