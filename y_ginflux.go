package ginflux

import (
	"fmt"
	"reflect"

	client2 "github.com/influxdata/influxdb1-client/v2"
)

type (
	//HTTPConfig alias
	HTTPConfig = client2.HTTPConfig
)

//RetentionPolicy ...
type RetentionPolicy struct {
	DBName        string
	RPName        string
	Duration      string
	Replication   int
	ShardDuration string
	Default       bool
}

//DefaultHTTPConfig ...get default config for dial
func DefaultHTTPConfig() HTTPConfig {
	return HTTPConfig{
		Addr:               "http://127.0.0.1:8086",
		Username:           "",
		Password:           "",
		UserAgent:          "ginflux/v2",
		Timeout:            0,
		InsecureSkipVerify: false,
		TLSConfig:          nil,
		Proxy:              nil,
	}
}

//NewEngine ...
func NewEngine() InterfaceGInflux {
	g := &newGInflux{}
	return g
}

type newGInflux struct {
	client client2.Client

	db string
}

//Client2 ...get raw client
func (g *newGInflux) Client2() client2.Client {
	return g.client
}

//Dial 全局只需要拨号一次
func (g *newGInflux) Dial(conf ...HTTPConfig) (err error) { return g.dial(conf...) }
func (g *newGInflux) dial(conf ...HTTPConfig) (err error) {
	var cf HTTPConfig
	if len(conf) > 0 {
		cf = conf[0]
	} else {
		cf = DefaultHTTPConfig()
	}
	g.client, err = client2.NewHTTPClient(cf)
	if err != nil {
		return
	}

	//interval ping
	/*go func() {
		tk := time.NewTicker(time.Second * 30)
		defer tk.Stop()
		for {
			select {
			case <-tk.C:
				_, _, _ = g.client.Ping(time.Second * 5)
			}
		}
	}()*/
	return
}

//Sync ...创建数据库,存储方案
func (g *newGInflux) Sync(db string, policy RetentionPolicy) (err error) {
	qStr := fmt.Sprintf(`CREATE DATABASE "%s"`, db)
	rp, err := g.client.Query(client2.NewQuery(qStr, db, ""))
	if err != nil {
		return
	}
	if err = rp.Error(); err != nil {
		return
	}
	policy.DBName = db
	return g.createRetentionPolicy(policy)
}

//createRetentionPolicy ...
func (g *newGInflux) createRetentionPolicy(rp RetentionPolicy) (err error) {
	//CREATE RETENTION POLICY <retention_policy_name>
	// ON <database_name> DURATION <duration> REPLICATION <n> [SHARD DURATION <duration>] [DEFAULT]
	if len(rp.RPName) == 0 {
		panic("must give a retention policy name")
	}
	if len(rp.Duration) == 0 {
		rp.Duration = "90d"
	}
	if len(rp.ShardDuration) == 0 {
		rp.ShardDuration = "1d"
	}
	qStr := fmt.Sprintf(
		`CREATE RETENTION POLICY %s ON %s DURATION %s REPLICATION 1 SHARD DURATION %s`,
		rp.RPName,
		rp.DBName,
		rp.Duration,
		rp.ShardDuration)
	if rp.Default {
		qStr += " DEFAULT"
	}
	qr := client2.NewQuery(qStr, rp.DBName, "ns")
	_, err = g.client.Query(qr)
	return
}

func (g *newGInflux) clone() *newGInflux {
	return &newGInflux{
		client: g.client,
	}
}

//DB ...
func (g *newGInflux) DB(dbName string) InterfaceGInflux {
	newIc := g.clone()
	newIc.db = dbName
	return newIc
}

//Insert ...
func (g *newGInflux) Insert(bean interface{}, bConf ...client2.BatchPointsConfig) (err error) {
	return g.insert(bean, bConf...)
}
func (g *newGInflux) insert(bean interface{}, bConf ...client2.BatchPointsConfig) (err error) {
	var (
		bps client2.BatchPoints
		bc  client2.BatchPointsConfig
	)
	if len(bConf) > 0 {
		bc = bConf[0]
	} else {
		bc = client2.BatchPointsConfig{
			Precision:        "ns",
			Database:         g.db,
			RetentionPolicy:  "",
			WriteConsistency: "",
		}
	}
	if bps, err = client2.NewBatchPoints(bc); err != nil {
		return
	}
	bVal := reflect.Indirect(reflect.ValueOf(bean))
	switch bVal.Kind() {
	case reflect.Struct:
		p := client2.NewPointFrom(NewStructBean(bean).Point())
		bps.AddPoint(p)
	case reflect.Slice:
		for i := 0; i < bVal.Len(); i++ {
			vBean := bVal.Slice(i, i+1).Index(0).Interface()
			p := client2.NewPointFrom(NewStructBean(vBean).Point())
			bps.AddPoint(p)
		}
	default:
		panic("gInflux:insert; unSupport insert type")
	}
	err = g.client.Write(bps)
	return
}

//Query ...
//if len(beans)>0 will exec bind
func (g *newGInflux) Query(sqlStr string, beans ...interface{}) (err error) {
	return g.query(sqlStr, beans...)
}
func (g *newGInflux) query(sqlStr string, beans ...interface{}) (err error) {
	queryStringAddTz(&sqlStr)
	cq := client2.NewQuery(sqlStr, g.db, "ns")
	rp, err := g.client.Query(cq)
	if err != nil {
		return
	}
	if err = rp.Error(); err != nil {
		return
	}
	if len(beans) > 0 {
		return bindSlice(rp, beans[0])
	}
	return
}
