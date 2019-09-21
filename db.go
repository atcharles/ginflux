package ginflux

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	ic "github.com/influxdata/influxdb1-client/v2"
)

//Database ...
type Database struct {
	session        *Session
	name           string
	client         *OClient
	batchPointConf ic.BatchPointsConfig
	bpsFunc        bpsActionFunc
}

//SetName ...
func (d *Database) SetName(name string) *Database {
	dNew := d.clone()
	dNew.name = name
	return dNew
}

type bpsActionFunc func(points ic.BatchPoints) error

//SetBpsFunc ...
func (d *Database) SetBpsFunc(bpsFunc bpsActionFunc) *Database {
	d.bpsFunc = bpsFunc
	return d
}

//SetBatchPointConf ...
func (d *Database) SetBatchPointConf(batchPointConf ic.BatchPointsConfig) *Database {
	d.batchPointConf = batchPointConf
	return d
}

func (d *Database) autoReleaseCallback(fn func(Database *Database) error) error {
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

func (d *Database) clone() *Database {
	return &Database{
		session: d.session,
		name:    d.name,
		client:  d.client,
	}
}

func (d *Database) scopeSearch() *Search {
	return &Search{Database: d.clone()}
}

//Query ...
func (d *Database) Query(str string, bean ...interface{}) (s *Search, err error) {
	err = d.autoReleaseCallback(func(Database *Database) error {
		//log.Println("------------------------!!!")
		s = Database.scopeSearch().queryDO(str).exec(bean...)
		if s.Err != nil {
			return s.Err
		}
		return s.Result.Error()
	})
	return
}

//Insert ...
func (d *Database) Insert(bean interface{}) error {
	return d.autoReleaseCallback(func(Database *Database) error {
		return Database.insert(bean)
	})
}

func (d *Database) insert(bean interface{}) (err error) {
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
	//ErrEmpty ...
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
	series := rp.Results[0].Series
	if len(series) == 0 {
		return ErrEmpty
	}
	var (
		columns []string
		rpVs    [][]interface{}
		tags    map[string]string
	)
	tags = rp.Results[0].Series[0].Tags
	columns = rp.Results[0].Series[0].Columns
	for i, column := range columns {
		indexMap[column] = i
	}
	for _, ss := range series {
		rpVs = append(rpVs, ss.Values...)
	}
	beans := reflect.MakeSlice(beanValue.Type(), 0, len(rpVs))
	vT := beanValue.Type().Elem()
	if vT.Kind() == reflect.Ptr {
		vT = vT.Elem()
	}
	for _, sl := range rpVs {
		b1 := reflect.New(vT)
		if err := bindBean(&b1, sl, indexMap, tags); err != nil {
			return err
		}
		beans = reflect.Append(beans, b1)
	}
	beanValue.Set(beans)
	return nil
}

func bindBean(item *reflect.Value, row []interface{}, indexMap map[string]int, tagsMap map[string]string) error {
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
			if err := bindBean(&fVal, row, indexMap, tagsMap); err != nil {
				return fmt.Errorf("inner bindBean error:%s", err.Error())
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
		//------------------------------ [START] ------------------------------
		//2019-08-21 17:05:    Author BY: charles
		//绑定tag
		for key, value := range tagsMap {
			if key == fieldName {
				if err := StringVal(value).Bind(&fVal); err != nil {
					return err
				}
			}
		}
		//------------------------------ [END] --------------------------------
		tagMap := tagMap(tags)
		if name, ok := tagMap[FieldName]; ok {
			fieldName = name
		}
		if _, ok := tagMap[TAGTime]; ok {
			fieldName = "time"
		}
		idx, ok := indexMap[fieldName]
		if !ok {
			continue
		}
		setVV := row[idx]
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
			ns, _ := strconv.ParseInt(ToStr(row[indexMap["time"]]), 10, 64)
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
