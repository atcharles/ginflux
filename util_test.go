package ginflux

import (
	"fmt"
	"reflect"
	"testing"
)

func TestStringVal_Bind(t *testing.T) {
	bean := new(TBA)
	vv := reflect.ValueOf(bean).Elem()
	for i := 0; i < vv.NumField(); i++ {
		fv := vv.Field(i)
		fmt.Println(vv.Type().Field(i).Name)
		fmt.Println(fv.String())
		fmt.Println(fv.Type().String())
		fmt.Println(fv.Type().Name())
	}

}

type TBA struct {
	AA string
	BB int
}
