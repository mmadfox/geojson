package geojson

import (
	"testing"

	"github.com/mmadfox/geojson/geometry"
	"github.com/tidwall/gjson"
)

func TestFeatureParse(t *testing.T) {
	p := expectJSON(t, `{"type":"Feature","geometry":{"type":"Point","coordinates":[1,2,3]},"properties":{}}`, nil)
	expect(t, p.Center() == P(1, 2))
	expectJSON(t, `{"type":"Feature"}`, errGeometryMissing)
	expectJSON(t, `{"type":"Feature","geometry":null}`, errDataInvalid)
	expectJSON(t, `{"type":"Feature","geometry":{"type":"Point","coordinates":[1,2,3]},"bbox":null,"properties":{}}`, nil)
	expectJSON(t, `{"type":"Feature","geometry":{"type":"Point","coordinates":[1,2]},"id":[4,true],"properties":{}}`, nil)
	expectJSON(t, `{"type":"Feature","geometry":{"type":"Point","coordinates":[1,2]},"id":"15","properties":{"a":"b"}}`, nil)
	expectJSON(t, `{"type":"Feature","geometry":{"type":"Point","coordinates":[1,2,3]},"bbox":[1,2,3,4],"properties":{}}`, nil)
	expectJSON(t, `{"type":"Feature","geometry":{"type":"Point","coordinates":[1,2],"bbox":[1,2,3,4]},"id":[4,true],"properties":{}}`, nil)
	json := `{"type":"Feature","rules":[{"id": "id", "spec":"spec"}],"geometry":{"type":"Point","coordinates":[1,2],"bbox":[1,2,3,4]},"id":[4,true],"properties":{"name": "name"}}`
	expectJSON(t, json, json)
}

func TestFeatureVarious(t *testing.T) {
	var g = expectJSON(t, `{"type":"Feature","id": "A", "geometry":{"type":"Point","coordinates":[1,2,3]},"properties":{}}`, nil)
	expect(t, string(g.AppendJSON(nil)) == `{"type":"Feature","geometry":{"type":"Point","coordinates":[1,2,3]},"properties":{}}`)
	expect(t, g.Rect() == R(1, 2, 1, 2))
	expect(t, g.Center() == P(1, 2))
	expect(t, !g.Empty())

	g = expectJSONOpts(t,
		`{"type":"Feature","geometry":{"type":"Point","coordinates":[1,2,3]},"bbox":[1,2,3,4],"properties":{}}`,
		nil, nil)
	expect(t, !g.Empty())
	expect(t, g.Rect() == R(1, 2, 1, 2))
	expect(t, g.Center() == P(1, 2))

}

func TestFeatureProperties(t *testing.T) {
	obj, err := Parse(`{"type":"Feature","geometry":{"type":"Point","coordinates":[1,2]}}`, nil)
	if err != nil {
		t.Fatal(err)
	}
	json := obj.JSON()
	if !gjson.Valid(json) {
		t.Fatal("invalid json")
	}
	if !gjson.Get(json, "properties").Exists() {
		t.Fatal("expected 'properties' member")
	}

	obj, err = Parse(`{"type":"Feature","geometry":{"type":"Point","coordinates":[1,2]},"properties":true}`, nil)
	if err != nil {
		t.Fatal(err)
	}
	json = obj.JSON()
	if !gjson.Valid(json) {
		t.Fatal("invalid json")
	}
	if gjson.Get(json, "properties").Type != gjson.True {
		t.Fatal("expected 'properties' member to be 'true'")
	}

	obj, err = Parse(`{"type":"Feature","geometry":{"type":"Point","coordinates":[1,2]},"id":{}}`, nil)
	if err != nil {
		t.Fatal(err)
	}
	json = obj.JSON()
	if !gjson.Valid(json) {
		t.Fatal("invalid json")
	}
	if !gjson.Get(json, "properties").Exists() {
		t.Fatal("expected 'properties' member")
	}
	if gjson.Get(json, "id").String() != "{}" {
		t.Fatal("expected 'id' member")
	}

}

// https://github.com/tidwall/tile38/issues/529
func TestIssue529(t *testing.T) {
	o, err := Parse(`{"type":"LineString","coordinates":[[0,0],[0,1]]}`, nil)
	if err != nil {
		t.Fatal(err)
	}
	ls1 := o.(*LineString)
	o, err = Parse(` {"type":"Feature","geometry":{"type":"LineString","coordinates":[[0,0],[0,1]]},"properties":{}}`, nil)
	if err != nil {
		t.Fatal(err)
	}
	ls2 := o.(*Feature)
	circ := NewCircle(geometry.Point{X: 0, Y: 0.5}, 5000, 64)
	expect(t, ls1.Intersects(circ))
	expect(t, circ.Intersects(ls1))
	expect(t, ls2.Intersects(circ))
	expect(t, circ.Intersects(ls2))
}
