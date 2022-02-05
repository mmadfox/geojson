package geojson

import (
	"github.com/tidwall/geojson/geometry"
	"strings"
	"testing"
)

func TestWalkRuleFeatureCollection(t *testing.T) {
	json := `{"type":"FeatureCollection","rules":[{"id":"id", "name":"0","spec":"spec"}], "features":[
		{"type":"Feature","rules":[{"id":"id", "name":"1","spec":"spec"}],"id":"A","geometry":{"type":"Point","coordinates":[1,2]},"properties":{}},
		{"type":"Feature","rules":[{"id":"id", "name":"2","spec":"spec"}],"id":"B","geometry":{"type":"Point","coordinates":[3,4]},"properties":{}},
		{"type":"Feature","rules":[{"id":"id", "name":"3","spec":"spec"}],"id":"C","geometry":{"type":"Point","coordinates":[5,6]},"properties":{}},
		{"type":"Feature","rules":[{"id":"id", "name":"4","spec":"spec"}],"id":"D","geometry":{"type":"Point","coordinates":[7,8]},"properties":{}}
	]}`

	g, _ := Parse(json, nil)
	names := make([]string, 0)
	WalkRule(g, geometry.Rect{}, func(rule Rule) bool {
		names = append(names, rule.Name)
		return true
	})

	if have, want := strings.Join(names, ""), "01234"; have != want {
		t.Fatalf("have %s, want %s", have, want)
	}
}

func TestWalkRuleFeature(t *testing.T) {
	json := `{"type":"Feature","rules":[{"id": "id", "name": "0", "spec":"spec"}],"geometry":{"type":"Point","coordinates":[1,2],"bbox":[1,2,3,4]},"id":[4,true],"properties":{}}`

	g, _ := Parse(json, nil)
	names := make([]string, 0)
	WalkRule(g, geometry.Rect{}, func(rule Rule) bool {
		names = append(names, rule.Name)
		return true
	})

	if have, want := strings.Join(names, ""), "0"; have != want {
		t.Fatalf("have %s, want %s", have, want)
	}
}
