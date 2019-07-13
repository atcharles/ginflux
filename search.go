package ginflux

import (
	ic "github.com/influxdata/influxdb1-client/v2"
)

type search struct {
	db         *db
	conditions []Map

	query ic.Query
}

/**
*	innerConditions()
**/

const (
	queryLang = "query"
	argsLang  = "args"
	//quoteReplaceStr = "?"
)

func (s *search) innerConditions(beans ...interface{}) *search {
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

func (s *search) queryDO(str string) *search {
	s.query = ic.NewQuery(str, s.db.name, "ns")
	return s
}

func (s *search) exec(bean ...interface{}) error {
	r, e := s.db.client.Query(s.query)
	if e != nil {
		return e
	}
	s.db.client.Release()
	if len(bean) > 0 {
		return bindSlice(r, bean[0])
	}
	return nil
}
