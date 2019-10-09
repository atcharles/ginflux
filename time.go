package ginflux

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

const (
	//TimeZone ...
	TimeZone = "Asia/Shanghai"
	//Custom ...
	Custom = "2006-01-02 15:04:05"

	queryTz = "tz('Asia/Shanghai')"
)

//queryStringAddTz 查询字符串添加时区
func queryStringAddTz(str *string) {
	s := *str
	if !strings.Contains(s, queryTz) {
		s = s + " " + queryTz
	}
	str = &s
}

//SetTimeZone ...
func SetTimeZone() {
	lc, err := time.LoadLocation(TimeZone)
	if err == nil {
		time.Local = lc
	}
}

//JSONTime ...
type JSONTime time.Time

//FromDB ...
func (p *JSONTime) FromDB(data []byte) error {
	timeStd, _ := time.ParseInLocation(Custom, string(data), time.Local)
	*p = JSONTime(timeStd)
	return nil
}

//ToDB ...
func (p *JSONTime) ToDB() (b []byte, err error) {
	b = []byte(p.String())
	return
}

//SetByTime ...
func (p *JSONTime) SetByTime(timeVal time.Time) {
	*p = JSONTime(timeVal)
}

//Time ...
func (p *JSONTime) Time() time.Time {
	return p.Convert2Time()
}

//Convert2Time ...
func (p *JSONTime) Convert2Time() time.Time {
	return time.Time(*p).Local()
}

// Value insert timestamp into mysql need this function.
func (p JSONTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	var ti = (&p).Convert2Time()
	if ti.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return ti, nil
}

// Scan valueof time.Time
func (p *JSONTime) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*p = JSONTime(value)
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

// GobEncode implements the gob.GobEncoder interface.
func (p JSONTime) GobEncode() ([]byte, error) {
	return (&p).Convert2Time().MarshalBinary()
}

// GobDecode implements the gob.GobDecoder interface.
func (p *JSONTime) GobDecode(data []byte) error {
	s := p.Convert2Time()
	p1 := &s
	return p1.UnmarshalBinary(data)
}

//MarshalJSON ...
func (p JSONTime) MarshalJSON() ([]byte, error) {
	if time.Time(p).IsZero() {
		return []byte(`""`), nil
	}
	data := make([]byte, 0)
	data = append(data, '"')
	data = (&p).Convert2Time().AppendFormat(data, Custom)
	data = append(data, '"')
	return data, nil
}

//UnmarshalJSON ...
func (p *JSONTime) UnmarshalJSON(data []byte) error {
	local, _ := time.ParseInLocation(`"`+Custom+`"`, string(data), time.Local)
	*p = JSONTime(local)
	return nil
}

//String ...
func (p JSONTime) String() string {
	return (&p).Convert2Time().Format(Custom)
}

//Addr ...
func (p JSONTime) Addr() *JSONTime {
	return &p
}

//ToDatetime ...
func ToDatetime(in string) (JSONTime, error) {
	out, err := time.ParseInLocation(Custom, in, time.Local)
	return JSONTime(out), err
}

//Must2JSONTimeAddr maybe panic
func Must2JSONTimeAddr(in string) *JSONTime {
	j, err := ToDatetime(in)
	if err != nil {
		panic(err)
	}
	return &j
}
