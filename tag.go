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
)

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

func (ip *innerPoint) parseBeanTags(bean interface{}) {
	if ip == nil {
		return
	}
	if ip.tagMap == nil {
		ip.tagMap = make(MapString)
	}
	if ip.fieldsMap == nil {
		ip.fieldsMap = make(models.Fields)
	}
	v := reflect.Indirect(reflect.ValueOf(bean))
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		if !ast.IsExported(field.Name) {
			continue
		}
		tStr := field.Tag.Get(TAGKey)
		vv := reflect.Indirect(v.Field(i))
		if (vv.Kind() == reflect.Struct && len(tStr) == 0) || field.Anonymous {
			ip.parseBeanTags(vv.Interface())
			continue
		}
		fieldName := LintGonicMapper.Obj2Table(field.Name)
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
			ip.fieldsMap[fieldName] = v.Field(i).Interface()
		}
		if _, ok := tagMap[TAGTime]; ok {
			ip.tagTime, _ = v.Field(i).Interface().(time.Time)
		}
		if _, ok := tagMap[FieldJSON]; ok {
			bts, _ := json.Marshal(v.Field(i).Interface())
			ip.fieldsMap[fieldName] = BytesToString(bts)
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
