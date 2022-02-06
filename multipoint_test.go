package geojson

import "testing"

func TestMultiPoint(t *testing.T) {
	expectJSON(t, `{"type":"MultiPoint", "properties": {"prop0": "value0"}, "coordinates":[[1,2]]}`, `{"type":"MultiPoint", "properties": {"prop0": "value0"}, "coordinates":[[1,2]]}`)
	p := expectJSON(t, `{"type":"MultiPoint","coordinates":[[1,2,3]]}`, nil)
	expect(t, p.Center() == P(1, 2))
	expectJSON(t, `{"type":"MultiPoint","coordinates":[1,null]}`, errCoordinatesInvalid)
	expectJSON(t, `{"type":"MultiPoint","coordinates":[[1,2]],"bbox":null}`, nil)
	expectJSON(t, `{"type":"MultiPoint"}`, errCoordinatesMissing)
	expectJSON(t, `{"type":"MultiPoint","coordinates":null}`, errCoordinatesInvalid)
	expectJSON(t, `{"type":"MultiPoint","coordinates":[[1,2,3],[4,5,6]],"bbox":[1,2,3,4]}`, nil)
}
