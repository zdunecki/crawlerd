package client

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"testing"
)

type TestStruct struct {
	Key string `json:"key"`
}

func TestDecodeJSON(t *testing.T) {
	//var s []*TestStruct

	//s := []*TestStruct{}
	s := make([]TestStruct, 0)
	//s[0] = &TestStruct{}
	//var s *TestStruct

	err := decode2(&s)
	if err != nil {
		t.Error(err)
		return
	}

	//if s.Key == "" {
	//	t.Error("stuct is nil")
	//}

	if s[0].Key != "value" {
		t.Error("stuct is nil")
	}
}

func GetBody(r io.Reader) ([]byte, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func decode(outPtr interface{}) error {
	return json.NewDecoder(bytes.NewReader([]byte(`{"key": "value"}`))).Decode(&outPtr)
	//rawData, err := GetBody(bytes.NewReader([]byte(`{"key": "value"}`)))
	//if err != nil {
	//	return err
	//}
	//
	//
	//return json.Unmarshal(rawData, outPtr)
}


func decode2(outPtr interface{}) error {
	return json.NewDecoder(bytes.NewReader([]byte(`[{"key": "value"}]`))).Decode(outPtr)
}
