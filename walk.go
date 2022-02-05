package geojson

import "github.com/tidwall/geojson/geometry"

func WalkRule(root Object, bbox geometry.Rect, fn func(*Rule, Object) bool) bool {
	skipIndex := bbox.Min.X == 0 && bbox.Min.Y == 0 &&
		bbox.Max.X == 0 && bbox.Max.Y == 0

	nexts := func(geom Object) bool {
		return walk(geom, fn)
	}
	nextc := func(geom Object) bool {
		return walkCollection(root, geom, fn)
	}
	switch typ := root.(type) {
	case *SimplePoint:
		return typ.ForEach(nexts)
	case *Point:
		typ.ForEach(nexts)
	case *Polygon:
		typ.ForEach(nexts)
	case *LineString:
		typ.ForEach(nexts)
	case *Rect:
		typ.ForEach(nexts)
	case *Circle:
		typ.ForEach(nexts)
	case *Feature:
		typ.ForEach(nexts)
	case *FeatureCollection:
		if skipIndex {
			if !typ.ForEach(nextc) {
				return false
			}
		} else {
			typ.Search(bbox, nextc)
		}
	case *GeometryCollection:
		if skipIndex {
			if !typ.ForEach(nextc) {
				return false
			}
		} else {
			typ.Search(bbox, nextc)
		}
	case *MultiPolygon:
		if skipIndex {
			if !typ.ForEach(nextc) {
				return false
			}
		} else {
			typ.Search(bbox, nextc)
		}
	case *MultiPoint:
		if skipIndex {
			if !typ.ForEach(nextc) {
				return false
			}
		} else {
			typ.Search(bbox, nextc)
		}
	case *MultiLineString:
		if skipIndex {
			if !typ.ForEach(nextc) {
				return false
			}
		} else {
			typ.Search(bbox, nextc)
		}
	}
	return true
}

func walk(geom Object, fn func(*Rule, Object) bool) bool {
	iter := func(rule *Rule) bool {
		if !fn(rule, geom) {
			return false
		}
		return true
	}
	if !geom.ForEachRule(iter) {
		return false
	}
	return true
}

func walkCollection(root, geom Object, fn func(*Rule, Object) bool) bool {
	iter := func(rule *Rule) bool {
		if !fn(rule, geom) {
			return false
		}
		return true
	}
	if ok := root.ForEachRule(iter); !ok {
		return false
	}
	if !geom.ForEachRule(iter) {
		return false
	}
	return true
}
