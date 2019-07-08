package ginflux

import (
	"fmt"

	ic "github.com/influxdata/influxdb1-client/v2"
)

//Session ...
type Session struct {
	engine *Engine
	client *OClient
}

//Client ...
func (s *Session) Client() *OClient {
	return s.client
}

//SetClient ...
func (s *Session) SetClient(client *OClient) *Session {
	s.client = client
	return s
}

//NewDB ...
func (s *Session) NewDB(name string) (dbInstance *Database) {
	dbInstance = &Database{name: name}
	if s.client != nil {
		dbInstance.client = s.client
	}
	dbInstance.session = s
	return
}

func (s *Session) createDB(name string) error {
	qStr := fmt.Sprintf(`CREATE DATABASE "%s"`, name)
	qr := ic.NewQuery(qStr, name, "ns")
	_, err := s.client.Query(qr)
	return err
}

//CreateDB ...
func (s *Session) CreateDB(name string) error {
	return s.createDB(name)
}

//CreateRetentionPolicy ...
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
	_, err := s.client.Query(qr)
	if err != nil {
		return err
	}
	return nil
}

//Release ...
func (s *Session) Release() {
	s.client.Release()
}
