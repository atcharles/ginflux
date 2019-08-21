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

type (
	//GRoomOrder 注单表
	GRoomOrder struct {
		//influx 时间字段
		Created *JSONTime `json:"created,omitempty" influx:"time"`
		RoomOrder
		UUID string `json:"uuid,omitempty" influx:"tag"`
	}
	RoomOrder struct {
		UID      int64   `json:"uid" influx:"tag"`
		RoomID   int64   `json:"room_id" influx:"tag"`
		Gid      int64   `json:"gid" influx:"tag"`
		Issue    string  `json:"issue" influx:"tag"`
		Status   int     `json:"status" influx:"tag"`
		OpenCode string  `json:"open_code" influx:"field"`
		BetCount int     `json:"bet_count" influx:"field"`
		BetMoney int64   `json:"bet_money" influx:"field"`
		Bonus    float64 `json:"bonus" influx:"field"`
		WinMoney float64 `json:"win_money" influx:"field"`

		Contents *OrderContent `json:"contents" influx:"json"`
	}

	//BetContent 投注内容单条结构,需要存入数据库中,并进行序列化
	BetContent struct {
		//Name  string  `json:"name"`
		Money int     `json:"money"`
		Odds  float64 `json:"odds"`
		Win   bool    `json:"win"`
		Bonus float64 `json:"bonus"`
	}
	//PlayContents 玩法投注内容映射
	PlayContents map[string]*BetContent
	//OrderContent 玩家投注的内容,key是具体的玩法(比如:重庆时时彩->第一球("1"))
	//value 为投注内容,针对一个玩法可以投注很多内容
	//OrderContent map[string][]*BetContent
	OrderContent map[string]PlayContents
)

func Test_01(t *testing.T) {
	beans := make([]*GRoomOrder, 0)
	sqlStr := `select sum("bet_money") as bet_money,
sum("bonus") as bonus,sum("win_money") as win_money from "g_room_order"
 where  "room_id" = '1'  group by "room_id","gid","uid" slimit 100`
	s, err := testGlobalEngine.DB("BusinessDB").Query(sqlStr, &beans)
	fmt.Printf("%#v\n,%#v\n", s, err)
	for _, b := range beans {
		fmt.Printf("%#v\n", b)
		fmt.Printf("%#v\n", b.Contents)
		fmt.Printf("%#v\n", b.Created.String())
	}
}
