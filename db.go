package ginflux

import (
	"reflect"

	ic "github.com/influxdata/influxdb1-client/v2"
)

type db struct {
	name           string
	client         *oClient
	batchPointConf ic.BatchPointsConfig
	bpsFunc        bpsActionFunc
}

func (d *db) SetName(name string) *db {
	d.name = name
	return d
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
	d.client.Release()
	return
}
