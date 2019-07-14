package ginflux

import (
	"fmt"
	"testing"
	"time"
)

func Test_obj2Time(t *testing.T) {
	tt := JSONTime(time.Now())
	fmt.Printf("%s\n", obj2Time(tt).String())
}
