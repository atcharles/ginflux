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

	PP *TPObj `influx:"field json"`
}

type TPObj struct {
	P1 string
	P2 int64
	P3 bool
}

func (m *TModel) Measurement() string {
	return "user"
}

type TEmbedded struct {
	UID     int64     `influx:"tag"`
	Created *JSONTime `influx:"time field name=created_field"`
	OK      bool      `influx:"field"`
	OBJ     struct {
		A string
		B int
		C float64
		D bool
	} `influx:"field"`
}

func Test_db_insert(t *testing.T) {
	now := time.Now()
	var beans []*TModel
	for i := 0; i < 1000; i++ {
		var b1 = &TModel{
			TEmbedded: TEmbedded{
				UID:     int64(i) + 1,
				Created: nil,
				OK:      true,
				OBJ: struct {
					A string
					B int
					C float64
					D bool
				}{
					A: "AA",
					B: 1,
					C: 2.111,
					D: false,
				},
			},
			IncrID: i,
			Name:   "model",
			Money:  rand.Float64() * 100,
			PP: &TPObj{
				P1: "PPP1",
				P2: 11222,
				P3: true,
			},
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
