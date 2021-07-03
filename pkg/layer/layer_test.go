package layer

import (
	"encoding/json"
	"os"
	"testing"
)

func TestNewLayerEngine(t *testing.T) {
	engine := NewLayerEngine()
	resp, err := engine.httpRequest("https://livesession.io/")
	if err != nil {
		t.Error(err)
	}

	root, err := NewHTMLReader(resp.Body).Encode()
	if err != nil {
		t.Error(err)
	}

	f, err := os.Create("page.json")
	if err != nil {
		t.Error(err)
	}
	b, err := json.Marshal(root)
	if err != nil {
		t.Error(err)
	}
	if _, err := f.Write(b); err != nil {
		t.Error(err)
	}

	diff, err := engine.Diff(resp.Body, resp.Body)
	if err != nil {
		t.Error(err)
	}

	t.Log(diff)
}

func TestNewLayerEngine2(t *testing.T) {
	engine := NewLayerEngine()

	diff, err := engine.Diff(nil, nil)
	if err != nil {
		t.Error(err)
	}

	t.Log(diff)
}
