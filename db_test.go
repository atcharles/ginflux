package ginflux

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

type TModel struct {
	TEmbedded
	IncrID int     `influx:"tag"`
	Name   string  `influx:"field"`
	Money  float64 `influx:"field"`
}

func (m *TModel) Measurement() string {
	return "user"
}

type TEmbedded struct {
	UID     int64     `influx:"tag"`
	Created *JSONTime `influx:"time field name=created_field"`
}

func Test_db_insert(t *testing.T) {
	now := time.Now()
	var beans []*TModel
	for i := 0; i < 1000; i++ {
		b1 := &TModel{
			IncrID: i,
			Name:   "model",
			Money:  rand.Float64() * 100,
		}
		b1.UID = int64(i) + 1
		now := JSONTime(time.Now())
		b1.Created = &now
		beans = append(beans, b1)
	}
	if err := testGlobalEngine.DB("db1").Insert(beans); err != nil {
		t.Error(err)
	}
	fmt.Printf("end time; use %s\n", time.Now().Sub(now).String())
}
