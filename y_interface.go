package ginflux

import (
	client "github.com/influxdata/influxdb1-client/v2"
)

//InterfaceGInflux ...接口
type InterfaceGInflux interface {
	Dial(conf ...HTTPConfig) (err error)
	Client2() client.Client
	Sync(db string, policy RetentionPolicy) (err error)
	DB(dbName string) InterfaceGInflux
	Insert(bean interface{}, bConf ...client.BatchPointsConfig) (err error)
	Query(sqlStr string, beans ...interface{}) (rp *client.Response, err error)
	QueryRaw(sqlStr string, beans ...interface{}) (rp *client.Response, err error)
}
