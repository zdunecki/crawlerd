package abc

import (
	"encoding/json"
	"fmt"
	"testing"

	"crawlerd/pkg/store/ql"
)

// TODO: own marshaller/unmarshaller

type Abc struct {
	ID    string `json:"id,omitempty"`
	Value string `json:"value,omitempty"`
}

type qlString struct {
	s        *string
	nillable bool
}

func String(s string) qlString {
	return qlString{
		s: ql.String(s),
	}
}

func NilString(s *string) qlString {
	return qlString{
		s:        s,
		nillable: true,
	}
}

func (q qlString) MarshalJSON() ([]byte, error) {
	r := "{}"

	if !q.nillable {
		if q.s != nil {
			r = fmt.Sprintf(`{"value": "%s"}`, *q.s)
		}
	} else {
		if q.s == nil {
			r = `{"value": null}`
		} else {
			r = fmt.Sprintf(`{"value": "%s"}`, *q.s)
		}
	}

	return []byte(r), nil
}

type Abc2 struct {
	ID    qlString `json:"id,omitempty"`
	Value qlString `json:"value,omitempty"`
}

//func (q *Abc2) MarshalJSON() ([]byte, error) {
//	// TODO: marshall generator for ql structs
//	return []byte(`{}`), nil
//}

func TestAbc(t *testing.T) {
	{
		v := &Abc{
			ID: "abc",
		}

		b, _ := json.Marshal(v)

		t.Log(string(b))
	}

	{
		v := &Abc{
			ID:    "abc",
			Value: "",
		}

		b, _ := json.Marshal(v)

		t.Log(string(b))
	}

	{
		v := &Abc2{
			ID: String(""),
			//Value: NilString(nil),
			//ID:   "",
			//Value: NillableStrig,
			//ID:    "abc",
		}

		b, _ := json.Marshal(v)

		t.Log(string(b))
	}

}
