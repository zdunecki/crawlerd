package ql

import (
	"encoding/json"
	"fmt"
	"testing"
)

type RequestQueueListQL struct {
	RunID     QLString `json:"run_id,omitempty"`
	CreatedAt QLInt64  `json:"created_at,omitempty"`
}

func (ql *RequestQueueListQL) MarshalJSON() ([]byte, error) {
	// below code should be auto generated
	type RequestQueueListQL struct {
		RunID     string `json:"run_id,omitempty"`
		CreatedAt int64  `json:"created_at,omitempty"`
	}

	return json.Marshal(&RequestQueueListQL{
		RunID:     ql.RunID.Val(),
		CreatedAt: ql.CreatedAt.Val(),
	})
}

// TODO: custom marshaller/unmarshaller
// idea

/*

 */

func TestQL(t *testing.T) {
	//data := &RequestQueueListQL{
	//	RunID: String("123"),
	//}

	// TODO:
	// TODO: query language engine

	// GraphQL concept
	//raw := `query {
	//	requestQueueList(
	//		query: {
	//			bool: {
	//				must: [
	//						{
	//							"term": {
	//								run_id: "123"
	//
	//							}
	//						},
	//							{
	//							  range: {
	//								created: {
	//									gt: 300
	//								}
	//							  }
	//							}
	//					],
	//				]
	//			}
	//		}
	//	)
	//}`

	// Query Builder concept
	//fmt.Println(x.CreatedAt.Val())

	//{ $or: [ { quantity: { $lt: 20 } }, { price: 10 } ] }
	bq := NewBoolQuery()
	w1 := NewTermQuery(&RequestQueueListQL{
		RunID: String("123", "321"),
	})

	// { price: { $gt: 50 } }, { price: {$lte: 100} }
	w2 := NewRangeQuery().Gt(&RequestQueueListQL{
		CreatedAt: Int64(50),
	}).Lte(&RequestQueueListQL{
		CreatedAt: Int64(100),
	})

	// { $and: [{ price: { $gt: 50 } }, { price: {$lte: 100} }] }
	bq.Must(w2)
	bq.Should(w1)

	and := func(v interface{}) string {
		vB, _ := json.Marshal(v)

		return fmt.Sprintf("{$and:[%s]}", vB)
	}

	or := func(v interface{}) string {
		vB, _ := json.Marshal(v)

		return fmt.Sprintf("{$or:[%s]}", vB)
	}

	term := func(v interface{}) interface{} {
		vB, _ := json.Marshal(v)
		return string(vB)
	}

	packOperator := func(v interface{}, operator string) interface{} {
		var expressions map[string]interface{}

		vB, _ := json.Marshal(v)
		json.Unmarshal(vB, &expressions)

		for k, v := range expressions {
			expressions[k] = fmt.Sprintf("{$%s:%v}", operator, v)
		}

		return expressions
	}

	gt := func(v interface{}) interface{} {
		var expressions map[string]interface{}

		vB, _ := json.Marshal(v)
		json.Unmarshal(vB, &expressions)

		operator := "gt"
		for k, v := range expressions {
			expressions[k] = fmt.Sprintf("{$%s:%v}", operator, v)
		}

		return expressions
	}
	gte := func(v interface{}) interface{} {
		return packOperator(v, "gte")
	}

	lt := func(v interface{}) interface{} {
		return packOperator(v, "lt")
	}
	lte := func(v interface{}) interface{} {
		return packOperator(v, "lte")
	}

	data := &RequestQueueListQL{
		RunID:     String("123"),
		CreatedAt: Int64(100),
	}

	{
		g := term(data)
		andOut := and(g)
		orOut := or(g)
		t.Log(andOut, orOut)
	}

	t.Log("OOO")
	{
		g := gt(data)
		andOut := and(g)
		orOut := or(g)
		t.Log(andOut, orOut)
	}
	{
		g := gte(data)
		andOut := and(g)
		orOut := or(g)
		t.Log(andOut, orOut)
	}

	{
		g := lt(data)
		andOut := and(g)
		orOut := or(g)
		t.Log(andOut, orOut)
	}
	{
		g := lte(data)
		andOut := and(g)
		orOut := or(g)
		t.Log(andOut, orOut)
	}

	//dataB, _ := json.Marshal(data)
	//t.Log(string(dataB))
	//elastic.RangeGTE()
	//w2 := elastic.NewRangeQuery().Gte(&RequestQueueListQL{
	//	RunID:     StringPtr(nil),
	//	CreatedAt: Int64(time.Now().Unix()),
	//})

	//NewRangeQuery("created_at").Gt(100).Lt(500)
	//NewRangeQuery("money").Gt(150)

	//elastic.RangeGT(&RequestQueueListQL{
	//	CreatedAt: Int64(100),
	//	Money:     Int(150),
	//}).RangeLt(&RequestQueueListQL{
	//	CreatedAt: Int64(500),
	//})

	//bq.Must()
	//data.RunID
}
