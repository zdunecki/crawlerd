package ql

import "fmt"

type Query interface {
	// Source returns the JSON-serializable query request.
	Source() (interface{}, error)
}

type TermQuery struct {
	name            string
	value           interface{}
	boost           *float64
	caseInsensitive *bool
	queryName       string
}

// NewTermQuery creates and initializes a new TermQuery.
//func NewTermQuery(name string, value interface{}) *TermQuery {
//	return &TermQuery{name: name, value: value}
//}

// NewTermQuery creates and initializes a new TermQuery.
func NewTermQuery(value interface{}) *TermQuery {
	return &TermQuery{value: value}
}

// Boost sets the boost for this query.
func (q *TermQuery) Boost(boost float64) *TermQuery {
	q.boost = &boost
	return q
}

func (q *TermQuery) CaseInsensitive(caseInsensitive bool) *TermQuery {
	q.caseInsensitive = &caseInsensitive
	return q
}

// QueryName sets the query name for the filter that can be used
// when searching for matched_filters per hit
func (q *TermQuery) QueryName(queryName string) *TermQuery {
	q.queryName = queryName
	return q
}

// Source returns JSON for the query.
func (q *TermQuery) Source() (interface{}, error) {
	// {"term":{"name":"value"}}
	source := make(map[string]interface{})
	tq := make(map[string]interface{})
	source["term"] = tq

	if q.boost == nil && q.caseInsensitive == nil && q.queryName == "" {
		tq[q.name] = q.value
	} else {
		subQ := make(map[string]interface{})
		subQ["value"] = q.value
		if q.boost != nil {
			subQ["boost"] = *q.boost
		}
		if q.caseInsensitive != nil {
			subQ["case_insensitive"] = *q.caseInsensitive
		}
		if q.queryName != "" {
			subQ["_name"] = q.queryName
		}
		tq[q.name] = subQ
	}
	return source, nil
}

type RangeQuery struct {
	name         string
	from         interface{}
	to           interface{}
	timeZone     string
	includeLower bool
	includeUpper bool
	boost        *float64
	queryName    string
	format       string
	relation     string
}

// NewRangeQuery creates and initializes a new RangeQuery.
//func NewRangeQuery(name string) *RangeQuery {
//	return &RangeQuery{name: name, includeLower: true, includeUpper: true}
//}

func NewRangeQuery() *RangeQuery {
	return &RangeQuery{includeLower: true, includeUpper: true}
}

func NewRangeQueryGT(name interface{}) *RangeQuery {
	return &RangeQuery{includeLower: true, includeUpper: true}
}

// From indicates the from part of the RangeQuery.
// Use nil to indicate an unbounded from part.
func (q *RangeQuery) From(from interface{}) *RangeQuery {
	q.from = from
	return q
}

// Gt indicates a greater-than value for the from part.
// Use nil to indicate an unbounded from part.
func (q *RangeQuery) Gt(from interface{}) *RangeQuery {
	q.from = from
	q.includeLower = false
	return q
}

// Gte indicates a greater-than-or-equal value for the from part.
// Use nil to indicate an unbounded from part.
func (q *RangeQuery) Gte(from interface{}) *RangeQuery {
	q.from = from
	q.includeLower = true
	return q
}

// To indicates the to part of the RangeQuery.
// Use nil to indicate an unbounded to part.
func (q *RangeQuery) To(to interface{}) *RangeQuery {
	q.to = to
	return q
}

// Lt indicates a less-than value for the to part.
// Use nil to indicate an unbounded to part.
func (q *RangeQuery) Lt(to interface{}) *RangeQuery {
	q.to = to
	q.includeUpper = false
	return q
}

// Lte indicates a less-than-or-equal value for the to part.
// Use nil to indicate an unbounded to part.
func (q *RangeQuery) Lte(to interface{}) *RangeQuery {
	q.to = to
	q.includeUpper = true
	return q
}

// IncludeLower indicates whether the lower bound should be included or not.
// Defaults to true.
func (q *RangeQuery) IncludeLower(includeLower bool) *RangeQuery {
	q.includeLower = includeLower
	return q
}

// IncludeUpper indicates whether the upper bound should be included or not.
// Defaults to true.
func (q *RangeQuery) IncludeUpper(includeUpper bool) *RangeQuery {
	q.includeUpper = includeUpper
	return q
}

// Boost sets the boost for this query.
func (q *RangeQuery) Boost(boost float64) *RangeQuery {
	q.boost = &boost
	return q
}

// QueryName sets the query name for the filter that can be used when
// searching for matched_filters per hit.
func (q *RangeQuery) QueryName(queryName string) *RangeQuery {
	q.queryName = queryName
	return q
}

// TimeZone is used for date fields. In that case, we can adjust the
// from/to fields using a timezone.
func (q *RangeQuery) TimeZone(timeZone string) *RangeQuery {
	q.timeZone = timeZone
	return q
}

// Format is used for date fields. In that case, we can set the format
// to be used instead of the mapper format.
func (q *RangeQuery) Format(format string) *RangeQuery {
	q.format = format
	return q
}

// Relation is used for range fields. which can be one of
// "within", "contains", "intersects" (default) and "disjoint".
func (q *RangeQuery) Relation(relation string) *RangeQuery {
	q.relation = relation
	return q
}

// Source returns JSON for the query.
func (q *RangeQuery) Source() (interface{}, error) {
	source := make(map[string]interface{})

	rangeQ := make(map[string]interface{})
	source["range"] = rangeQ

	params := make(map[string]interface{})
	rangeQ[q.name] = params

	params["from"] = q.from
	params["to"] = q.to
	if q.timeZone != "" {
		params["time_zone"] = q.timeZone
	}
	if q.format != "" {
		params["format"] = q.format
	}
	if q.relation != "" {
		params["relation"] = q.relation
	}
	if q.boost != nil {
		params["boost"] = *q.boost
	}
	params["include_lower"] = q.includeLower
	params["include_upper"] = q.includeUpper

	if q.queryName != "" {
		rangeQ["_name"] = q.queryName
	}

	return source, nil
}

type BoolQuery struct {
	Query
	mustClauses        []Query
	mustNotClauses     []Query
	filterClauses      []Query
	shouldClauses      []Query
	boost              *float64
	minimumShouldMatch string
	adjustPureNegative *bool
	queryName          string
}

// Creates a new bool query.
func NewBoolQuery() *BoolQuery {
	return &BoolQuery{
		mustClauses:    make([]Query, 0),
		mustNotClauses: make([]Query, 0),
		filterClauses:  make([]Query, 0),
		shouldClauses:  make([]Query, 0),
	}
}

func (q *BoolQuery) Must(queries ...Query) *BoolQuery {
	q.mustClauses = append(q.mustClauses, queries...)
	return q
}

func (q *BoolQuery) MustNot(queries ...Query) *BoolQuery {
	q.mustNotClauses = append(q.mustNotClauses, queries...)
	return q
}

func (q *BoolQuery) Filter(filters ...Query) *BoolQuery {
	q.filterClauses = append(q.filterClauses, filters...)
	return q
}

func (q *BoolQuery) Should(queries ...Query) *BoolQuery {
	q.shouldClauses = append(q.shouldClauses, queries...)
	return q
}

func (q *BoolQuery) Boost(boost float64) *BoolQuery {
	q.boost = &boost
	return q
}

func (q *BoolQuery) MinimumShouldMatch(minimumShouldMatch string) *BoolQuery {
	q.minimumShouldMatch = minimumShouldMatch
	return q
}

func (q *BoolQuery) MinimumNumberShouldMatch(minimumNumberShouldMatch int) *BoolQuery {
	q.minimumShouldMatch = fmt.Sprintf("%d", minimumNumberShouldMatch)
	return q
}

func (q *BoolQuery) AdjustPureNegative(adjustPureNegative bool) *BoolQuery {
	q.adjustPureNegative = &adjustPureNegative
	return q
}

func (q *BoolQuery) QueryName(queryName string) *BoolQuery {
	q.queryName = queryName
	return q
}

// Creates the query source for the bool query.
func (q *BoolQuery) Source() (interface{}, error) {
	// {
	//	"bool" : {
	//		"must" : {
	//			"term" : { "user" : "kimchy" }
	//		},
	//		"must_not" : {
	//			"range" : {
	//				"age" : { "from" : 10, "to" : 20 }
	//			}
	//		},
	//    "filter" : [
	//      ...
	//    ]
	//		"should" : [
	//			{
	//				"term" : { "tag" : "wow" }
	//			},
	//			{
	//				"term" : { "tag" : "elasticsearch" }
	//			}
	//		],
	//		"minimum_should_match" : 1,
	//		"boost" : 1.0
	//	}
	// }

	query := make(map[string]interface{})

	boolClause := make(map[string]interface{})
	query["bool"] = boolClause

	// must
	if len(q.mustClauses) == 1 {
		src, err := q.mustClauses[0].Source()
		if err != nil {
			return nil, err
		}
		boolClause["must"] = src
	} else if len(q.mustClauses) > 1 {
		var clauses []interface{}
		for _, subQuery := range q.mustClauses {
			src, err := subQuery.Source()
			if err != nil {
				return nil, err
			}
			clauses = append(clauses, src)
		}
		boolClause["must"] = clauses
	}

	// must_not
	if len(q.mustNotClauses) == 1 {
		src, err := q.mustNotClauses[0].Source()
		if err != nil {
			return nil, err
		}
		boolClause["must_not"] = src
	} else if len(q.mustNotClauses) > 1 {
		var clauses []interface{}
		for _, subQuery := range q.mustNotClauses {
			src, err := subQuery.Source()
			if err != nil {
				return nil, err
			}
			clauses = append(clauses, src)
		}
		boolClause["must_not"] = clauses
	}

	// filter
	if len(q.filterClauses) == 1 {
		src, err := q.filterClauses[0].Source()
		if err != nil {
			return nil, err
		}
		boolClause["filter"] = src
	} else if len(q.filterClauses) > 1 {
		var clauses []interface{}
		for _, subQuery := range q.filterClauses {
			src, err := subQuery.Source()
			if err != nil {
				return nil, err
			}
			clauses = append(clauses, src)
		}
		boolClause["filter"] = clauses
	}

	// should
	if len(q.shouldClauses) == 1 {
		src, err := q.shouldClauses[0].Source()
		if err != nil {
			return nil, err
		}
		boolClause["should"] = src
	} else if len(q.shouldClauses) > 1 {
		var clauses []interface{}
		for _, subQuery := range q.shouldClauses {
			src, err := subQuery.Source()
			if err != nil {
				return nil, err
			}
			clauses = append(clauses, src)
		}
		boolClause["should"] = clauses
	}

	if q.boost != nil {
		boolClause["boost"] = *q.boost
	}
	if q.minimumShouldMatch != "" {
		boolClause["minimum_should_match"] = q.minimumShouldMatch
	}
	if q.adjustPureNegative != nil {
		boolClause["adjust_pure_negative"] = *q.adjustPureNegative
	}
	if q.queryName != "" {
		boolClause["_name"] = q.queryName
	}

	return query, nil
}
