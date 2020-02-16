package ginflux

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unsafe"

	client2 "github.com/influxdata/influxdb1-client/v2"
)

//BytesToString convert []byte type to string type.
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

//StringToBytes convert string type to []byte type.
// NOTE: panic if modify the member value of the []byte.
func StringToBytes(s string) []byte {
	sp := *(*[2]uintptr)(unsafe.Pointer(&s))
	bp := [3]uintptr{sp[0], sp[1], sp[1]}
	return *(*[]byte)(unsafe.Pointer(&bp))
}

//SnakeString converts the accepted string to a snake string (XxYy to xx_yy)
/*func SnakeString(s string) string {
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
}*/

//CamelString converts the accepted string to a camel string (xx_yy to XxYy)
/*func CamelString(s string) string {
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
}*/

//ObjectName gets the type name of the object
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

//IsSliceContainsStr returns true if the string exists in given slice, ignore case.
/*func IsSliceContainsStr(sl []string, str string) bool {
	str = strings.ToLower(str)
	for _, s := range sl {
		if strings.ToLower(s) == str {
			return true
		}
	}
	return false
}*/

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
		vv := reflect.Indirect(reflect.ValueOf(v))
		switch vv.Type().Kind() {
		case reflect.Struct, reflect.Slice, reflect.Map, reflect.Interface:
			b, e := json.Marshal(v)
			if e != nil {
				panic("写入influx: json 序列化错误: " + e.Error())
			}
			s = BytesToString(b)
		default:
			s = fmt.Sprintf("%v", v)
		}
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

func unmarshalJSON(str string, fVal *reflect.Value) (err error) {
	if len(str) == 0 {
		return
	}
	val := reflect.Indirect(*fVal)
	b1 := reflect.New(val.Type()).Interface()
	err = json.Unmarshal([]byte(str), b1)
	if err != nil {
		return err
	}
	val.Set(reflect.ValueOf(b1).Elem())
	return
}

var (
	//ErrEmpty ...
	ErrEmpty = errors.New("empty")
)

func bindSlice(rp *client2.Response, bean interface{}) error {
	if rp.Error() != nil {
		return rp.Error()
	}
	vv := reflect.ValueOf(bean)
	if vv.Kind() != reflect.Ptr {
		panic("need pointer for bind bean")
	}
	beanValue := reflect.Indirect(vv)
	if len(rp.Results) == 0 {
		return fmt.Errorf("没有返回的数据:服务端错误 %s", rp.Error().Error())
	}
	series := rp.Results[0].Series
	if len(series) == 0 {
		return ErrEmpty
	}
	//将series分割成数组
	valArr := make([]map[string]interface{}, 0)
	for _, sv := range series {
		for _, vs := range sv.Values {
			var maps = make(map[string]interface{})
			for i1, col := range sv.Columns {
				maps[col] = vs[i1]
			}
			for key, value := range sv.Tags {
				maps[key] = value
			}
			valArr = append(valArr, maps)
		}
	}
	beans := reflect.MakeSlice(beanValue.Type(), 0, len(valArr))
	vT := beanValue.Type().Elem()
	if vT.Kind() == reflect.Ptr {
		vT = vT.Elem()
	}
	for _, value := range valArr {
		b1 := reflect.New(vT)
		beans = reflect.Append(beans, b1)
		if err := bindBean(&b1, value); err != nil {
			return err
		}
	}
	beanValue.Set(beans)
	return nil
}

func bindBean(item *reflect.Value, row map[string]interface{}) error {
	v := reflect.Indirect(*item)
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		fVal := v.Field(i)
		if !fVal.CanSet() {
			continue
		}

		if field.Type.Kind() == reflect.Ptr {
			fVal.Set(reflect.New(field.Type.Elem()))
		}

		fieldName := LintGonicMapper.Obj2Table(field.Name)
		tStr := field.Tag.Get(TAGKey)
		if (reflect.Indirect(fVal).Kind() == reflect.Struct && len(tStr) == 0) || field.Anonymous {
			if err := bindBean(&fVal, row); err != nil {
				return fmt.Errorf("inner bindBean error:%s;hints:[%s]", err.Error(), field.Name)
			}
			continue
		}
		tags := splitTag(tStr)
		if len(tags) == 0 {
			continue
		}
		if tags[0] == "-" {
			continue
		}
		tagMap := tagMap(tags)
		if name, ok := tagMap[FieldName]; ok {
			fieldName = name
		}
		if _, ok := tagMap[TAGTime]; ok {
			fieldName = "time"
		}
		setVV, ok := row[fieldName]
		if !ok {
			continue
		}
		if _, ok := tagMap[FieldJSON]; ok {
			if err := unmarshalJSON(ToStr(setVV), &fVal); err != nil {
				return err
			}
			continue
		}
		if _, ok := tagMap[TAG]; ok {
			fVal = reflect.Indirect(fVal)
			if err := StringVal(ToStr(setVV)).Bind(&fVal); err != nil {
				return err
			}
			continue
		}
		if _, ok := tagMap[TAGTime]; ok {
			fVal = reflect.Indirect(fVal)
			ns, _ := strconv.ParseInt(ToStr(row["time"]), 10, 64)
			tm1 := time.Unix(int64(time.Duration(ns)/time.Second), int64(time.Duration(ns)%time.Second))
			fVal.Set(reflect.ValueOf(tm1.Local()).Convert(fVal.Type()))
			continue
		}
		switch val := fVal.Interface().(type) {
		case Conversion:
			if err := val.FromDB(StringToBytes(ToStr(setVV))); err != nil {
				return err
			}
		default:
			fVal = reflect.Indirect(fVal)
			if err := StringVal(ToStr(setVV)).Bind(&fVal); err != nil {
				return err
			}
		}
	}
	return nil
}
