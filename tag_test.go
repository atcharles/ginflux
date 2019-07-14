package ginflux

import (
	"fmt"
	"testing"
	"time"
)

func Test_obj2Time(t *testing.T) {
	var tIc InterfaceTime
	tt := JSONTime(time.Now())

	tIc = &tt
	fmt.Printf("ok %#v\n", tIc.Time().String())
	fmt.Printf("ok %#v\n", tt.String())
}
