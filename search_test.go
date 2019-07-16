package ginflux

import (
	"fmt"
	"testing"
)

func Test_Query(t *testing.T) {
	db, _ := testGlobalEngine.NewDB("db1")
	str := `SELECT * FROM "user" ORDER BY "time" DESC LIMIT 2 OFFSET 0`
	var beans []*TModel
	if err := db.Query(str, &beans); err != nil {
		t.Error(err)
	}
	for _, val := range beans {
		fmt.Printf("bean value is :%#v\n", val)
		b, _ := json.MarshalIndent(val, "", "  ")
		fmt.Printf("%s\n", b)

	}
}
