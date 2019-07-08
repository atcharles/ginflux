package ginflux

import (
	"fmt"
	"reflect"
	"testing"
)

func TestStringVal_Bind(t *testing.T) {
	str := "123"
	var a int
	v := reflect.Indirect(reflect.ValueOf(&a))
	_ = StringVal(str).Bind(&v)
	fmt.Printf("a is %d\n", a)
}
