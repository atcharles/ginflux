package ginflux

import (
	"testing"
	"time"
)

type TModel struct {
	TEmbedded
	IncrID int    `influx:"tag"`
	Name   string `influx:"field"`
}

func (m *TModel) Measurement() string {
	return "user"
}

type TEmbedded struct {
	UID     int64     `influx:"tag"`
	Created *JSONTime `influx:"time field name=created_field"`
}

func Test_db_insert(t *testing.T) {
	var beans []*TModel
	for i := 0; i < 100; i++ {
		b1 := &TModel{
			IncrID: i,
			Name:   "model",
		}
		b1.UID = int64(i) + 1
		now := JSONTime(time.Now())
		b1.Created = &now
		beans = append(beans, b1)
	}
	db, err := testGlobalEngine.NewDB("db1")
	if err != nil {
		t.Error(err)
	}
	if err = db.insert(beans); err != nil {
		t.Error(err)
	}
}
