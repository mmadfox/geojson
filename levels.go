package geojson

import "github.com/tidwall/gjson"

func resIsArray(res gjson.Result) bool {
	if res.Type == gjson.JSON {
		for i := 0; i < len(res.Raw); i++ {
			if res.Raw[i] == '[' {
				return true
			}
			if res.Raw[i] <= ' ' {
				continue
			}
			break
		}
	}
	return false
}

////////////////////////////////
// level 1
////////////////////////////////

func fillLevel1Map(json string) (
	coordinates Position, bbox *BBox, err error,
) {
	coords := gjson.Get(json, "coordinates")
	switch coords.Type {
	default:
		err = errInvalidCoordinates
		return
	case gjson.Null:
		err = errCoordinatesRequired
		return
	case gjson.JSON:
		if !resIsArray(coords) {
			err = errInvalidCoordinates
			return
		}
		coordinates, err = fillPosition(coords)
		if err != nil {
			return
		}
	}
	bbox, err = fillBBox(json)
	return
}

func level1CalculatedBBox(coordinates Position, bbox *BBox) BBox {
	if bbox != nil {
		return *bbox
	}
	return BBox{
		Min: coordinates,
		Max: coordinates,
	}
}

func level1PositionCount(coordinates Position, bbox *BBox) int {
	if bbox != nil {
		return 3
	}
	return 1
}

func level1Weight(coordinates Position, bbox *BBox) int {
	return level1PositionCount(coordinates, bbox) * sizeofPosition
}

func appendLevel1JSON(
	json []byte, name string,
	coordinates Position,
	bbox *BBox, bboxDefined bool,
) []byte {
	if bbox != nil && !bboxDefined {
		bbox = nil
	}
	isCordZ := level1IsCoordZDefined(coordinates, bbox)
	json = append(json, `{"type":"`...)
	json = append(json, name...)
	json = append(json, `","coordinates":[`...)
	json = appendPositionJSON(json, coordinates, isCordZ)
	json = append(json, ']')
	if bboxDefined {
		json = appendBBoxJSON(json, bbox)
	}
	return append(json, '}')
}

func level1IsCoordZDefined(coordinates Position, bbox *BBox) bool {
	if bbox.isCordZDefined() {
		return true
	}
	return coordinates.Z != nilz
}

////////////////////////////////
// level 2
////////////////////////////////

func fillLevel2Map(json string) (
	coordinates []Position, bbox *BBox, err error,
) {
	coords := gjson.Get(json, "coordinates")
	switch coords.Type {
	default:
		err = errInvalidCoordinates
		return
	case gjson.Null:
		err = errCoordinatesRequired
		return
	case gjson.JSON:
		if !resIsArray(coords) {
			err = errInvalidCoordinates
			return
		}
		v := coords.Array()
		coordinates = make([]Position, len(v))
		for i, coords := range v {
			if !resIsArray(coords) {
				err = errInvalidCoordinates
				return
			}
			var p Position
			p, err = fillPosition(coords)
			if err != nil {
				return
			}
			coordinates[i] = p
		}
	}
	bbox, err = fillBBox(json)
	return
}

func level2CalculatedBBox(coordinates []Position, bbox *BBox) BBox {
	if bbox != nil {
		return *bbox
	}
	_, bbox2 := positionBBox(0, BBox{}, coordinates)
	return bbox2
}

func level2PositionCount(coordinates []Position, bbox *BBox) int {
	if bbox != nil {
		return 2 + len(coordinates)
	}
	return len(coordinates)
}

func level2Weight(coordinates []Position, bbox *BBox) int {
	return level2PositionCount(coordinates, bbox) * sizeofPosition
}

func appendLevel2JSON(
	json []byte, name string,
	coordinates []Position,
	bbox *BBox, bboxDefined bool,
) []byte {
	if bbox != nil && !bboxDefined {
		bbox = nil
	}
	isCordZ := level2IsCoordZDefined(coordinates, bbox)
	json = append(json, `{"type":"`...)
	json = append(json, name...)
	json = append(json, `","coordinates":[`...)
	for i, p := range coordinates {
		if i > 0 {
			json = append(json, ',')
		}
		json = append(json, '[')
		json = appendPositionJSON(json, p, isCordZ)
		json = append(json, ']')
	}
	json = append(json, ']')
	if bboxDefined {
		json = appendBBoxJSON(json, bbox)
	}
	json = append(json, '}')
	return json
}

func level2IsCoordZDefined(coordinates []Position, bbox *BBox) bool {
	if bbox.isCordZDefined() {
		return true
	}
	for _, p := range coordinates {
		if p.Z != nilz {
			return true
		}
	}
	return false
}

////////////////////////////////
// level 3
////////////////////////////////

func fillLevel3Map(json string) (
	coordinates [][]Position, bbox *BBox, err error,
) {
	coords := gjson.Get(json, "coordinates")
	switch coords.Type {
	default:
		err = errInvalidCoordinates
		return
	case gjson.Null:
		err = errCoordinatesRequired
		return
	case gjson.JSON:
		if !resIsArray(coords) {
			err = errInvalidCoordinates
			return
		}
		v := coords.Array()
		coordinates = make([][]Position, len(v))
		for i, coords := range v {
			if !resIsArray(coords) {
				err = errInvalidCoordinates
				return
			}
			v := coords.Array()
			ps := make([]Position, len(v))
			for i, coords := range v {
				if !resIsArray(coords) {
					err = errInvalidCoordinates
					return
				}
				var p Position
				p, err = fillPosition(coords)
				if err != nil {
					return
				}
				ps[i] = p
			}
			coordinates[i] = ps
		}
	}
	bbox, err = fillBBox(json)
	return
}

func level3CalculatedBBox(
	coordinates [][]Position, bbox *BBox, isPolygon bool,
) BBox {
	if bbox != nil {
		return *bbox
	}
	var bbox2 BBox
	var i = 0
	for _, ps := range coordinates {
		i, bbox2 = positionBBox(i, bbox2, ps)
		if isPolygon {
			break // only the exterior ring should be calculated for a polygon
		}
	}
	return bbox2
}

func level3Weight(coordinates [][]Position, bbox *BBox) int {
	return level3PositionCount(coordinates, bbox) * sizeofPosition
}

func level3PositionCount(coordinates [][]Position, bbox *BBox) int {
	var res int
	for _, p := range coordinates {
		res += len(p)
	}
	if bbox != nil {
		return 2 + res
	}
	return res
}

func appendLevel3JSON(
	json []byte, name string,
	coordinates [][]Position,
	bbox *BBox, bboxDefined bool,
) []byte {
	if bbox != nil && !bboxDefined {
		bbox = nil
	}
	isCordZ := level3IsCoordZDefined(coordinates, bbox)
	json = append(json, `{"type":"`...)
	json = append(json, name...)
	json = append(json, `","coordinates":[`...)
	for i, p := range coordinates {
		if i > 0 {
			json = append(json, ',')
		}
		json = append(json, '[')
		for i, p := range p {
			if i > 0 {
				json = append(json, ',')
			}
			json = append(json, '[')
			json = appendPositionJSON(json, p, isCordZ)
			json = append(json, ']')
		}
		json = append(json, ']')
	}
	json = append(json, ']')
	if bboxDefined {
		json = appendBBoxJSON(json, bbox)
	}
	return append(json, '}')
}

func level3IsCoordZDefined(coordinates [][]Position, bbox *BBox) bool {
	if bbox.isCordZDefined() {
		return true
	}
	for _, p := range coordinates {
		for _, p := range p {
			if p.Z != nilz {
				return true
			}
		}
	}
	return false
}

////////////////////////////////
// level 4
////////////////////////////////

func fillLevel4Map(json string) (
	coordinates [][][]Position, bbox *BBox, err error,
) {
	coords := gjson.Get(json, "coordinates")
	switch coords.Type {
	default:
		err = errInvalidCoordinates
		return
	case gjson.Null:
		err = errCoordinatesRequired
		return
	case gjson.JSON:
		if !resIsArray(coords) {
			err = errInvalidCoordinates
			return
		}
		v := coords.Array()
		coordinates = make([][][]Position, len(v))
		for i, coords := range v {
			if !resIsArray(coords) {
				err = errInvalidCoordinates
				return
			}
			v := coords.Array()
			ps := make([][]Position, len(v))
			for i, coords := range v {
				if !resIsArray(coords) {
					err = errInvalidCoordinates
					return
				}
				v := coords.Array()
				pss := make([]Position, len(v))
				for i, coords := range v {
					if !resIsArray(coords) {
						err = errInvalidCoordinates
						return
					}
					var p Position
					p, err = fillPosition(coords)
					if err != nil {
						return
					}
					pss[i] = p
				}
				ps[i] = pss
			}
			coordinates[i] = ps
		}
	}
	bbox, err = fillBBox(json)
	return
}

func level4Weight(coordinates [][][]Position, bbox *BBox) int {
	return level4PositionCount(coordinates, bbox) * sizeofPosition
}

func level4PositionCount(coordinates [][][]Position, bbox *BBox) int {
	var res int
	for _, p := range coordinates {
		for _, p := range p {
			res += len(p)
		}
	}
	if bbox != nil {
		return 2 + res
	}
	return res
}

func appendLevel4JSON(
	json []byte, name string,
	coordinates [][][]Position,
	bbox *BBox, bboxDefined bool,
) []byte {
	if bbox != nil && !bboxDefined {
		bbox = nil
	}
	isCordZ := level4IsCoordZDefined(coordinates, bbox)
	json = append(json, `{"type":"`...)
	json = append(json, name...)
	json = append(json, `","coordinates":[`...)
	for i, p := range coordinates {
		if i > 0 {
			json = append(json, ',')
		}
		json = append(json, '[')
		for i, p := range p {
			if i > 0 {
				json = append(json, ',')
			}
			json = append(json, '[')
			for i, p := range p {
				if i > 0 {
					json = append(json, ',')
				}
				json = append(json, '[')
				json = appendPositionJSON(json, p, isCordZ)
				json = append(json, ']')
			}
			json = append(json, ']')
		}
		json = append(json, ']')
	}
	json = append(json, ']')
	if bboxDefined {
		json = appendBBoxJSON(json, bbox)
	}
	return append(json, '}')
}

func level4IsCoordZDefined(coordinates [][][]Position, bbox *BBox) bool {
	if bbox.isCordZDefined() {
		return true
	}
	for _, p := range coordinates {
		for _, p := range p {
			for _, p := range p {
				if p.Z != nilz {
					return true
				}
			}
		}
	}
	return false
}
