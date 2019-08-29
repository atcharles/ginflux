package ginflux

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"testing"
)

func Test_Query(t *testing.T) {
	str := `SELECT * FROM "user" ORDER BY "time" DESC LIMIT 2 OFFSET 0`
	var beans []*TModel
	_, err := testGlobalEngine.DB("db1").Query(str, &beans)
	if err != nil {
		t.Error(err)
	}
	for _, val := range beans {
		fmt.Printf("bean value is :%#v\n", val)
		b, _ := json.MarshalIndent(val, "", "  ")
		fmt.Printf("%s\n", b)
	}
}

func Test_concurrent_Query(t *testing.T) {
	ats := 10
	wg := new(sync.WaitGroup)
	for i := 0; i < ats; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			//log.Println("current num", testGlobalEngine.pool.CurrentOpen())
			//str := `DROP SERIES FROM "user"`
			str := `SELECT * FROM "user" ORDER BY "time" DESC LIMIT 2 OFFSET 0`
			_, err := testGlobalEngine.DB("db1").Query(str)
			if err != nil {
				log.Println(err)
				return
			}
		}()
	}
	wg.Wait()
	log.Println("G current num", testGlobalEngine.pool.CurrentOpen())
}

func Test_Drop(t *testing.T) {
	str := `DROP SERIES FROM "user"`
	s, err := testGlobalEngine.DB("db1").Query(str)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("result is %#v\n", s.Result)
}

func Test_Query_1(t *testing.T) {
	str := `SELECT * FROM "g_open_data" WHERE "gid"='1' AND "issue"='20190725047' LIMIT 1`
	var beans []*GOpenData
	_, err := testGlobalEngine.DB("open_data_db").Query(str, &beans)
	if err != nil {
		t.Error(err)
		return
	}
	for _, val := range beans {
		//fmt.Printf("bean value is :%#v\n", val)
		b, _ := json.MarshalIndent(val, "", "  ")
		fmt.Printf("%s\n", b)
	}
}

type (
	Base struct {
		ID      int64     `json:"id" xorm:"autoincr pk" influx:"tag"`
		Version int64     `json:"version" xorm:"version notnull index"`
		Created *JSONTime `json:"created,omitempty" xorm:"created notnull index"`
		Updated *JSONTime `json:"updated,omitempty" xorm:"updated notnull index"`
	}
	//GOpenData 开奖数据
	GOpenData struct {
		Base     `xorm:"extends"`
		OpenData `xorm:"extends"`
	}
	//OpenData ...
	OpenData struct {
		LotName   string `json:"lot_name" influx:"tag"`
		LotNameCN string `json:"lot_name_cn" influx:"field"`

		Gid int64 `json:"gid" influx:"tag"`
		//开奖期号
		Issue string `json:"issue" influx:"tag"`
		//开奖号码
		OpenCode string `json:"open_code" influx:"field"`
		//开奖时间
		OpenTime *JSONTime `json:"open_time" influx:"field"`
		//开奖时间戳
		OpenTimestamp int64 `json:"open_timestamp" influx:"tag"`
		//下期开奖期号
		NextIssue string `json:"next_issue" influx:"field"`

		NextOpenTime *JSONTime `json:"next_open_time" influx:"field"`

		NextOpenTimestamp int64 `json:"next_open_timestamp" influx:"tag"`

		lt *GLotteries
		//必须实现 Conversion 接口
		OtherData interface{} `json:"other_data,omitempty"`
	}
)

type (
	//GLotteries 全局彩种列表结构体
	GLotteries struct {
		Base      `xorm:"extends"`
		Lotteries `xorm:"extends"`
	}
	//Lotteries ...
	Lotteries struct {
		Gid int64 `json:"gid" xorm:"notnull unique comment(彩种ID)" validate:"required"`
		//类型
		TypeID   int    `json:"type_id" xorm:"notnull index comment(彩种类型)" validate:"required"`
		TypeName string `json:"type_name" xorm:"varchar(20) notnull index"`
		//英文简称
		Name string `json:"name" xorm:"varchar(10) notnull unique comment(彩种简称)" validate:"required"`
		//中文名称
		NameCN string `json:"name_cn" xorm:"varchar(100) notnull" validate:"required"`
		//平台是否开启
		Enable bool `json:"enable" xorm:"notnull index default(1) comment(彩种开关)"`
		//彩种图标本地路径 ./assets/
		Icon string `json:"icon,omitempty" xorm:"notnull text comment(彩种图标)"`
		//开奖视频地址
		VideoURL string `json:"video_url,omitempty" xorm:"notnull text comment(彩种动画地址)"`
		//期号是否连续+1
		IssueAddOne bool           `json:"issue_add_one,omitempty" xorm:"notnull default(0) comment(彩种期号+1)"`
		Calibrator  *LotCalibrator `json:"calibrator,omitempty" xorm:"json comment(彩种校对器)"`
		//开奖时间计划
		TimePlans []*LOTOpenTimePlan `json:"time_plans,omitempty" xorm:"notnull json longtext comment(彩种开奖计划)"`
	}
	//LOTOpenTimePlan 彩种开奖时间计划
	LOTOpenTimePlan struct {
		Start, Stop  string
		PeriodSecond int
	}
	//LotCalibrator 开奖期号校准器
	//如果不是期号+1的彩种,为nil
	LotCalibrator struct {
		IssueNO  int      `json:"issue_no"`
		OpenTime JSONTime `json:"open_time"`
	}
)
