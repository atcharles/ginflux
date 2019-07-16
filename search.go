package ginflux

import (
	ic "github.com/influxdata/influxdb1-client/v2"
)

//Search ...
type Search struct {
	Database   *Database
	conditions []Map

	query  ic.Query
	Result *ic.Response
	Err    error
}

const (
	queryLang = "query"
	argsLang  = "args"
	//quoteReplaceStr = "?"
)

func (s *Search) innerConditions(beans ...interface{}) *Search {
	if len(beans) == 0 {
		return s
	}
	conditionsMap := Map{queryLang: beans[0]}
	if len(beans) > 1 {
		conditionsMap[argsLang] = beans[1:]
	}
	s.conditions = append(s.conditions, conditionsMap)
	return s
}

func (s *Search) queryDO(str string) *Search {
	s.query = ic.NewQuery(str, s.Database.name, "ns")
	return s
}

func (s *Search) exec(bean ...interface{}) *Search {
	r, e := s.Database.client.Query(s.query)
	if e != nil {
		s.Err = e
		return s
	}
	s.Result = r
	s.Database.client.Release()
	if len(bean) > 0 {
		s.Err = bindSlice(r, bean[0])
	}
	return s
}
