package demo

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/leeseika/cv-demo/pkg/jsonx"
)

//go:embed input.json
var inputJSON []byte

type Basket struct {
	Fruits []jsonx.JSONValue `json:"fruits"`
}

type Apple struct {
	Type   string `json:"type"`
	Color  string `json:"color"`
	Peeled bool   `json:"peeled"`
}

type Watermelon struct {
	Type     string `json:"type"`
	Sliced   bool   `json:"sliced"`
	HasSeeds bool   `json:"has_seeds"`
}

func TestLoadFruitsFromJSON(t *testing.T) {
	var basket Basket
	if err := json.Unmarshal(inputJSON, &basket); err != nil {
		t.Fatalf("failed to unmarshal input JSON: %v", err)
	}

	for _, fruit := range basket.Fruits {
		switch typ := fruit.Result().Get("type").String(); typ {
		case "apple":
			var apple Apple
			if err := json.Unmarshal(fruit.RawMessage, &apple); err != nil {
				t.Errorf("failed to unmarshal apple: %v", err)
				continue
			}
			t.Logf("Loaded Apple: %+v", apple)
		case "watermelon":
			var watermelon Watermelon
			if err := json.Unmarshal(fruit.RawMessage, &watermelon); err != nil {
				t.Errorf("failed to unmarshal watermelon: %v", err)
				continue
			}
			t.Logf("Loaded Watermelon: %+v", watermelon)
		default:
			t.Errorf("unknown fruit type: %s", typ)
		}
	}
}

func TestMarshalFruitsToJSON(t *testing.T) {
	apple := Apple{
		Type:   "apple",
		Color:  "red",
		Peeled: false,
	}
	watermelon := Watermelon{
		Type:     "watermelon",
		Sliced:   true,
		HasSeeds: false,
	}

	appleJSON, err := jsonx.Marshal(apple)
	if err != nil {
		t.Fatalf("failed to marshal apple: %v", err)
	}

	watermelonJSON, err := jsonx.Marshal(watermelon)
	if err != nil {
		t.Fatalf("failed to marshal watermelon: %v", err)
	}

	basket := Basket{
		Fruits: []jsonx.JSONValue{*appleJSON, *watermelonJSON},
	}

	basketJSON, err := json.Marshal(basket)
	if err != nil {
		t.Fatalf("failed to marshal basket: %v", err)
	}

	t.Logf("Marshalled basket JSON:\n%s", string(basketJSON))
}
