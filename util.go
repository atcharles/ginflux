package ginflux

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"unsafe"
)

// BytesToString convert []byte type to string type.
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// StringToBytes convert string type to []byte type.
// NOTE: panic if modify the member value of the []byte.
func StringToBytes(s string) []byte {
	sp := *(*[2]uintptr)(unsafe.Pointer(&s))
	bp := [3]uintptr{sp[0], sp[1], sp[1]}
	return *(*[]byte)(unsafe.Pointer(&bp))
}

// SnakeString converts the accepted string to a snake string (XxYy to xx_yy)
func SnakeString(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	for _, d := range StringToBytes(s) {
		if d >= 'A' && d <= 'Z' {
			if j {
				data = append(data, '_')
				j = false
			}
		} else if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	return strings.ToLower(BytesToString(data))
}

// CamelString converts the accepted string to a camel string (xx_yy to XxYy)
func CamelString(s string) string {
	data := make([]byte, 0, len(s))
	j := false
	k := false
	num := len(s) - 1
	for i := 0; i <= num; i++ {
		d := s[i]
		if k == false && d >= 'A' && d <= 'Z' {
			k = true
		}
		if d >= 'a' && d <= 'z' && (j || k == false) {
			d = d - 32
			j = false
			k = true
		}
		if k && d == '_' && num > i && s[i+1] >= 'a' && s[i+1] <= 'z' {
			j = true
			continue
		}
		data = append(data, d)
	}
	return string(data[:])
}

// ObjectName gets the type name of the object
func ObjectName(obj interface{}) string {
	v := reflect.ValueOf(obj)
	t := v.Type()
	switch t.Kind() {
	case reflect.Ptr:
		return v.Elem().Type().Name()
	case reflect.Struct:
		return t.Name()
	case reflect.Func:
		return runtime.FuncForPC(v.Pointer()).Name()
	}
	return t.String()
}

func splitTag(tag string) (tags []string) {
	tag = strings.TrimSpace(tag)
	var hasQuote = false
	var lastIdx = 0
	for i, t := range tag {
		if t == '\'' {
			hasQuote = !hasQuote
		} else if t == ' ' {
			if lastIdx < i && !hasQuote {
				tags = append(tags, strings.TrimSpace(tag[lastIdx:i]))
				lastIdx = i + 1
			}
		}
	}
	if lastIdx < len(tag) {
		tags = append(tags, strings.TrimSpace(tag[lastIdx:]))
	}
	return
}

// IsSliceContainsStr returns true if the string exists in given slice, ignore case.
func IsSliceContainsStr(sl []string, str string) bool {
	str = strings.ToLower(str)
	for _, s := range sl {
		if strings.ToLower(s) == str {
			return true
		}
	}
	return false
}

type argInt []int

func (a argInt) Get(i int, args ...int) (r int) {
	if i >= 0 && i < len(a) {
		r = a[i]
	} else if len(args) > 0 {
		r = args[0]
	}
	return
}

//Conversion ...
type Conversion interface {
	FromDB(data []byte) error
	ToDB() (b []byte, err error)
}

//ToStr Convert any type to string.
func ToStr(value interface{}, args ...int) (s string) {
	switch v := value.(type) {
	case bool:
		s = strconv.FormatBool(v)
	case float32:
		s = strconv.FormatFloat(float64(v), 'f', argInt(args).Get(0, -1), argInt(args).Get(1, 32))
	case float64:
		s = strconv.FormatFloat(v, 'f', argInt(args).Get(0, -1), argInt(args).Get(1, 64))
	case int:
		s = strconv.FormatInt(int64(v), argInt(args).Get(0, 10))
	case int8:
		s = strconv.FormatInt(int64(v), argInt(args).Get(0, 10))
	case int16:
		s = strconv.FormatInt(int64(v), argInt(args).Get(0, 10))
	case int32:
		s = strconv.FormatInt(int64(v), argInt(args).Get(0, 10))
	case int64:
		s = strconv.FormatInt(v, argInt(args).Get(0, 10))
	case uint:
		s = strconv.FormatUint(uint64(v), argInt(args).Get(0, 10))
	case uint8:
		s = strconv.FormatUint(uint64(v), argInt(args).Get(0, 10))
	case uint16:
		s = strconv.FormatUint(uint64(v), argInt(args).Get(0, 10))
	case uint32:
		s = strconv.FormatUint(uint64(v), argInt(args).Get(0, 10))
	case uint64:
		s = strconv.FormatUint(v, argInt(args).Get(0, 10))
	case string:
		s = v
	case []byte:
		s = string(v)
	case Conversion:
		b, _ := v.ToDB()
		s = string(b)
	default:
		if v == nil {
			return ""
		}
		b, _ := json.Marshal(v)
		s = BytesToString(b)
	}
	return s
}

//StringVal ...
type StringVal string

//Bind ...
func (s StringVal) Bind(fieldValue *reflect.Value) (err error) {
	fieldType := fieldValue.Type()
	if !fieldValue.CanAddr() {
		return errors.New("StringVal:Bind:fieldValue:cannot take address")
	}
	switch fieldType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var i int
		if i, err = strconv.Atoi(string(s)); err != nil {
			return
		}
		fieldValue.SetInt(int64(i))
	case reflect.Bool:
		var b bool
		if b, err = strconv.ParseBool(string(s)); err != nil {
			return
		}
		fieldValue.SetBool(b)
	case reflect.Float32, reflect.Float64:
		var v float64
		if v, err = strconv.ParseFloat(string(s), 64); err != nil {
			return
		}
		fieldValue.SetFloat(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var v uint64
		if v, err = strconv.ParseUint(string(s), 10, 64); err != nil {
			return
		}
		fieldValue.SetUint(v)
	case reflect.String:
		fieldValue.SetString(string(s))
	default:
		err = unmarshalJSON(string(s), fieldValue)
	}
	return
}

//Deprecated
//MapInterfaceToStruct ...
func (s StringVal) MapInterfaceToStruct(dstVal *reflect.Value) (err error) {
	vv := dstVal.Interface()
	if err = json.Unmarshal(StringToBytes(string(s)), &vv); err != nil {
		return
	}
	switch vvBean := vv.(type) {
	case map[string]interface{}:
		for key, value := range vvBean {
			fieldValue := dstVal.FieldByName(key)
			if err = StringVal(ToStr(value)).Bind(&fieldValue); err != nil {
				return
			}
		}
	case interface{}:
		return
	default:
		return fmt.Errorf("unsupported type: %s", dstVal.Type().String())
	}
	return
}

func unmarshalJSON(str string, fVal *reflect.Value) (err error) {
	val := reflect.Indirect(*fVal)
	b1 := reflect.New(val.Type()).Interface()
	err = json.Unmarshal([]byte(str), b1)
	if err != nil {
		return err
	}
	val.Set(reflect.ValueOf(b1).Elem())
	return
}
