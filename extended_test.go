package geojson

import (
	"strings"
	"testing"
)

func TestWalkRuleFeatureCollection(t *testing.T) {
	json := `{"type":"FeatureCollection","rules":[{"id":"1", "spec":"A"}, {"id":"2", "spec":"D"}], "features":[
		{"type":"Feature","id":"A","geometry":{"type":"Point","coordinates":[1,2]},"properties":{}},
		{"type":"Feature","id":"B","geometry":{"type":"Point","coordinates":[3,4]},"properties":{}},
		{"type":"Feature","id":"C","geometry":{"type":"Point","coordinates":[5,6]},"properties":{}},
		{"type":"Feature","id":"D","geometry":{"type":"Point","coordinates":[7,8]},"properties":{}}
	]}`

	g, _ := Parse(json, nil)
	ids := make([]string, 0)
	WalkRule(g, func(r *Rule, l Lookuper) bool {
		_, ok := l.Lookup(r.Spec())
		if ok {
			ids = append(ids, r.Spec())
		}
		return true
	})
	if have, want := strings.Join(ids, ""), "AD"; have != want {
		t.Fatalf("have %s, want %s", have, want)
	}
}
