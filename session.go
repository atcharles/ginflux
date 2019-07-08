package ginflux

import (
	"fmt"

	ic "github.com/influxdata/influxdb1-client/v2"
)

type Session struct {
	engine *Engine
}

func (s *Session) NewDB(name string) (dbInstance *db, err error) {
	dbInstance = &db{name: name}
	var cl *oClient
	cl, err = s.engine.Acquire()
	if err != nil {
		return
	}
	dbInstance.client = cl
	return
}

func (s *Session) CreateRetentionPolicy(rp RetentionPolicy) error {
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
	qr := ic.NewQuery(qStr, rp.DBName, "ns")
	client, err := s.engine.Acquire()
	if err != nil {
		return err
	}
	_, err = client.Query(qr)
	return err
}
