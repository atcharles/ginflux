package ginflux

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	ic "github.com/influxdata/influxdb1-client/v2"
)

type db struct {
	session        *Session
	name           string
	client         *oClient
	batchPointConf ic.BatchPointsConfig
	bpsFunc        bpsActionFunc
}

func (d *db) SetName(name string) *db {
	dNew := d.clone()
	dNew.name = name
	return dNew
}

type bpsActionFunc func(points ic.BatchPoints) error

func (d *db) SetBpsFunc(bpsFunc bpsActionFunc) *db {
	d.bpsFunc = bpsFunc
	return d
}

func (d *db) SetBatchPointConf(batchPointConf ic.BatchPointsConfig) *db {
	d.batchPointConf = batchPointConf
	return d
}

func (d *db) autoReleaseCallback(fn func(db *db) error) error {
	d1 := d.clone()
	oc, err := d1.session.engine.Acquire()
	if err != nil {
		return err
	}
	defer oc.Release()
	d1.client = oc
	if err = fn(d1); err != nil {
		return err
	}
	return nil
}

func (d *db) clone() *db {
	return &db{
		session: d.session,
		name:    d.name,
		client:  d.client,
	}
}

func (d *db) scopeSearch() *search {
	return &search{db: d.clone()}
}

func (d *db) Query(str string, bean ...interface{}) error {
	return d.scopeSearch().queryDO(str).exec(bean...)
}

func (d *db) Insert(bean interface{}) error {
	return d.autoReleaseCallback(func(db *db) error {
		return db.insert(bean)
	})
}

func (d *db) insert(bean interface{}) (err error) {
	if len(d.batchPointConf.Database) == 0 {
		d.batchPointConf.Database = d.name
	}
	var (
		bps ic.BatchPoints
	)
	if bps, err = ic.NewBatchPoints(d.batchPointConf); err != nil {
		return
	}
	if d.bpsFunc != nil {
		if err = d.bpsFunc(bps); err != nil {
			return
		}
	}
	bVal := reflect.Indirect(reflect.ValueOf(bean))
	switch bVal.Kind() {
	case reflect.Struct:
		p := ic.NewPointFrom(NewStructBean(bean).Point())
		bps.AddPoint(p)
	case reflect.Slice:
		for i := 0; i < bVal.Len(); i++ {
			vBean := bVal.Slice(i, i+1).Index(0).Interface()
			p := ic.NewPointFrom(NewStructBean(vBean).Point())
			bps.AddPoint(p)
		}
	default:
		panic("gInflux:insert; unSupport insert type")
	}
	if err = d.client.Write(bps); err != nil {
		return
	}
	return
}

var (
	ErrEmpty = errors.New("empty")
)

func bindSlice(rp *ic.Response, bean interface{}) error {
	if rp.Error() != nil {
		return rp.Error()
	}
	vv := reflect.ValueOf(bean)
	if vv.Kind() != reflect.Ptr {
		panic("need pointer for bind bean")
	}
	beanValue := reflect.Indirect(vv)
	var indexMap = make(map[string]int)
	if len(rp.Results) == 0 {
		return fmt.Errorf("没有返回的数据:服务端错误 %s", rp.Error().Error())
	}
	if len(rp.Results[0].Series) == 0 {
		return ErrEmpty
	}
	columns := rp.Results[0].Series[0].Columns
	for i, column := range columns {
		indexMap[column] = i
	}
	rpVs := rp.Results[0].Series[0].Values
	beans := reflect.MakeSlice(beanValue.Type(), 0, len(rpVs))
	vT := beanValue.Type().Elem()
	if vT.Kind() == reflect.Ptr {
		vT = vT.Elem()
	}
	for _, sl := range rpVs {
		b1 := reflect.New(vT)
		if err := bindBean(&b1, sl, indexMap); err != nil {
			return err
		}
		beans = reflect.Append(beans, b1)
	}
	beanValue.Set(beans)
	return nil
}

func bindBean(item *reflect.Value, row []interface{}, indexMap map[string]int) error {
	v := reflect.Indirect(*item)
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		var fVal reflect.Value
		if field.Type.Kind() == reflect.Ptr {
			v.Field(i).Set(reflect.New(field.Type.Elem()))
		}
		fVal = v.Field(i)
		fieldName := LintGonicMapper.Obj2Table(field.Name)
		tStr := field.Tag.Get(TAGKey)
		if (reflect.Indirect(fVal).Kind() == reflect.Struct && len(tStr) == 0) || field.Anonymous {
			if err := bindBean(&fVal, row, indexMap); err != nil {
				return fmt.Errorf("inner bindBean error:%s", err.Error())
			}
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
		setVV := row[indexMap[fieldName]]
		if !fVal.CanSet() {
			//fmt.Printf("field:%s cannot set value\n", field.Name)
			continue
		}
		if _, ok := tagMap[FieldJSON]; ok {
			fVal = reflect.Indirect(fVal)
			if err := json.Unmarshal(StringToBytes(setVV.(string)), &fVal); err != nil {
				return err
			}
			continue
		}
		if _, ok := tagMap[TAG]; ok {
			fVal = reflect.Indirect(fVal)
			if err := StringVal(setVV.(string)).Bind(&fVal); err != nil {
				return err
			}
			continue
		}
		if _, ok := tagMap[TAGTime]; ok {
			fVal = reflect.Indirect(fVal)
			timeVal := row[indexMap["time"]]
			timeStr := ToStr(timeVal)
			ns, _ := strconv.ParseInt(timeStr, 10, 64)
			tm1 := time.Unix(int64(time.Duration(ns)/time.Second), int64(time.Duration(ns)%time.Second))
			fVal.Set(reflect.ValueOf(tm1).Convert(fVal.Type()))
			//fmt.Printf("%#v\n", fVal.Interface().(JSONTime).String())
			continue
		}
		switch val := fVal.Interface().(type) {
		case Conversion:
			if err := val.FromDB(StringToBytes(setVV.(string))); err != nil {
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
