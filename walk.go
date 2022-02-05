package geojson

import "github.com/tidwall/geojson/geometry"

func WalkRule(root Object, bbox geometry.Rect, fn func(Rule) bool) bool {
	// root
	if ok := root.ForEachRule(fn); !ok {
		return false
	}
	skipIndex := bbox.Min.X == 0 && bbox.Min.Y == 0 &&
		bbox.Max.X == 0 && bbox.Max.Y == 0
	walkFunc := func(geom Object) bool {
		if !geom.ForEachRule(fn) {
			return false
		}
		return true
	}
	// children
	switch typ := root.(type) {
	case *FeatureCollection:
		if skipIndex {
			return typ.ForEach(walkFunc)
		} else {
			typ.Search(bbox, walkFunc)
		}
	case *GeometryCollection:
		if skipIndex {
			return typ.ForEach(walkFunc)
		} else {
			typ.Search(bbox, walkFunc)
		}
	case *MultiPolygon:
		if skipIndex {
			return typ.ForEach(walkFunc)
		} else {
			typ.Search(bbox, walkFunc)
		}
	case *MultiPoint:
		if skipIndex {
			return typ.ForEach(walkFunc)
		} else {
			typ.Search(bbox, walkFunc)
		}
	case *MultiLineString:
		if skipIndex {
			return typ.ForEach(walkFunc)
		} else {
			typ.Search(bbox, walkFunc)
		}
	}
	return true
}
