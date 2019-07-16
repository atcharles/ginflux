package ginflux

import (
	"go/ast"
	"reflect"
	"strings"
	"time"

	"github.com/influxdata/influxdb1-client/models"
)

const TAGKey = "influx"

type (
	Map       map[string]interface{}
	MapString map[string]string
)

//keywords
const (
	TAG       = "tag"
	Field     = "field"
	FieldName = "name"
	TAGTime   = "time"
	FieldJSON = "json"
)

type StructBean struct {
	bean interface{}
}

type (
	InterfaceTable interface {
		Measurement() string
	}

	InterfaceTime interface {
		Time() time.Time
	}
)

func obj2Time(bean interface{}) time.Time {
	if bean == nil {
		return time.Time{}
	}
	if tt, ok := bean.(InterfaceTime); ok {
		return tt.Time()
	}
	timeType := reflect.TypeOf(time.Time{})
	v1 := reflect.Indirect(reflect.ValueOf(bean))
	if v1.Type().ConvertibleTo(timeType) {
		return v1.Convert(timeType).Interface().(time.Time)
	}
	return time.Time{}
}

func NewStructBean(bean interface{}) *StructBean {
	return &StructBean{bean: bean}
}

func (s *StructBean) Point() models.Point {
	ip := new(innerPoint)
	ip.parseBeanTags(s.bean)
	mName := LintGonicMapper.Obj2Table(ObjectName(s.bean))
	if nm1, ok := s.bean.(InterfaceTable); ok {
		mName = nm1.Measurement()
	}
	return models.MustNewPoint(mName, models.NewTags(ip.tagMap), ip.fieldsMap, ip.tagTime)
}

type innerPoint struct {
	tagMap    MapString
	fieldsMap models.Fields
	tagTime   time.Time
}

func (ip *innerPoint) def() {
	if ip == nil {
		panic("innerPoint: nil")
	}
	if ip.tagMap == nil {
		ip.tagMap = make(MapString)
	}
	if ip.fieldsMap == nil {
		ip.fieldsMap = make(models.Fields)
	}
}

func (ip *innerPoint) parseBeanTags(bean interface{}) {
	ip.def()
	v := reflect.Indirect(reflect.ValueOf(bean))
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		if !ast.IsExported(field.Name) {
			continue
		}
		fieldName := LintGonicMapper.Obj2Table(field.Name)
		tStr := field.Tag.Get(TAGKey)
		vv := reflect.Indirect(v.Field(i))
		if (vv.Kind() == reflect.Struct && len(tStr) == 0) || field.Anonymous {
			ip.parseBeanTags(vv.Interface())
			continue
		}
		if len(tStr) == 0 {
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
		if _, ok := tagMap[TAG]; ok {
			ip.tagMap[fieldName] = ToStr(v.Field(i).Interface())
		}
		if _, ok := tagMap[Field]; ok {
			ip.fieldsMap[fieldName] = fieldValue(v.Field(i).Interface())
		}
		if _, ok := tagMap[TAGTime]; ok {
			ip.tagTime = obj2Time(v.Field(i).Interface())
		}
		if _, ok := tagMap[FieldJSON]; ok {
			bts, _ := json.Marshal(v.Field(i).Interface())
			ip.fieldsMap[fieldName] = BytesToString(bts)
			continue
		}
	}
	if ip.tagTime.IsZero() {
		ip.tagTime = time.Now()
	}
}

func tagMap(tags []string) MapString {
	mp := make(MapString)
	for _, value := range tags {
		sl := strings.Split(strings.TrimSpace(value), "=")
		if len(sl) > 1 {
			mp[sl[0]] = sl[1]
		} else {
			mp[sl[0]] = "TRUE"
		}
	}
	return mp
}

func fieldValue(field interface{}) (rs interface{}) {
	switch val := field.(type) {
	case float32, float64, int, int8, int16, int32,
		int64, uint, uint8, uint16, uint32, uint64, string:
		rs = val
	case []byte:
		rs = BytesToString(val)
	case Conversion:
		b, _ := val.ToDB()
		rs = BytesToString(b)
	default:
		b, _ := json.Marshal(val)
		rs = BytesToString(b)
	}
	return rs
}
