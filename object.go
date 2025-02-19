package geojson

import (
	"errors"
	"fmt"
	"github.com/mmadfox/geojson/geo"
	"github.com/mmadfox/geojson/geometry"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
	"strconv"
)

var (
	fmtErrTypeIsUnknown         = "type '%s' is unknown"
	errDataInvalid              = errors.New("invalid data")
	errTypeInvalid              = errors.New("invalid type")
	errTypeMissing              = errors.New("missing type")
	errCoordinatesInvalid       = errors.New("invalid coordinates")
	errRulesInvalid             = errors.New("invalid rules")
	errRuleIDInvalid            = errors.New("invalid rule id")
	errCoordinatesMissing       = errors.New("missing coordinates")
	errGeometryInvalid          = errors.New("invalid geometry")
	errGeometryMissing          = errors.New("missing geometry")
	errFeaturesMissing          = errors.New("missing features")
	errFeaturesInvalid          = errors.New("invalid features")
	errGeometriesMissing        = errors.New("missing geometries")
	errGeometriesInvalid        = errors.New("invalid geometries")
	errCircleRadiusUnitsInvalid = errors.New("invalid circle radius units")
)

// Object is a GeoJSON type
type Object interface {
	Empty() bool
	Valid() bool
	Rect() geometry.Rect
	Center() geometry.Point
	Contains(other Object) bool
	Within(other Object) bool
	Intersects(other Object) bool
	AppendJSON(dst []byte) []byte
	JSON() string
	String() string
	Distance(obj Object) float64
	NumPoints() int
	ForEach(iter func(geom Object) bool) bool
	Spatial() Spatial
	MarshalJSON() ([]byte, error)
}

var _ = []Object{
	&Point{}, &LineString{}, &Polygon{}, &Feature{},
	&MultiPoint{}, &MultiLineString{}, &MultiPolygon{},
	&GeometryCollection{}, &FeatureCollection{},
	&Rect{}, &Circle{}, &SimplePoint{},
}

// Collection is a searchable collection type.
type Collection interface {
	Children() []Object
	Indexed() bool
	Search(rect geometry.Rect, iter func(child Object) bool)
}

var _ = []Collection{
	&MultiPoint{}, &MultiLineString{}, &MultiPolygon{},
	&FeatureCollection{}, &GeometryCollection{},
}

type extra struct {
	dims   byte      // number of extra coordinate values, 1 or 2
	values []float64 // extra coordinate values
	// valid json object that includes extra members such as
	// "bbox", "id", "properties", and foreign members
	members string
}

// ParseOptions ...
type ParseOptions struct {
	// IndexChildren option will cause the object to indexByIDs their children
	// objects when the number of children is greater than or equal to the
	// provided value. Setting this value to 0 will disable indexing.
	// The default is 64.
	IndexChildren int
	// IndexGeometry option will cause the object to indexByIDs it's geometry
	// when the number of points in it's base polygon or linestring is greater
	// that or equal to the provided value. Setting this value to 0 will
	// disable indexing.
	// The default is 64.
	IndexGeometry int
	// IndexGeometryKind is the kind of indexByIDs implementation.
	// Default is QuadTreeCompressed
	IndexGeometryKind geometry.IndexKind
	// RequireValid option cause parse to fail when a geojson object is invalid.
	RequireValid bool
	// AllowSimplePoints options will force to parse to return the SimplePoint
	// type when a geojson point only consists of an 2D x/y coord and no extra
	// json members.
	AllowSimplePoints bool
	// DisableCircleType disables the special Circle syntax that is unique to
	// only Tile38.
	DisableCircleType bool
	// AllowRects options will force to parse and return the Rect type when a
	// geojson polygon only consists of a perfect rectangle, where there are
	// exactly 5 points with the first point being the min x/y and the
	// following point winding counter clockwise creating a closed rectangle.
	AllowRects bool
}

// DefaultParseOptions ...
var DefaultParseOptions = &ParseOptions{
	IndexChildren:     64,
	IndexGeometry:     64,
	IndexGeometryKind: geometry.QuadTree,
	RequireValid:      false,
	AllowSimplePoints: false,
	DisableCircleType: false,
	AllowRects:        false,
}

// Parse a GeoJSON object
func Parse(data string, opts *ParseOptions) (Object, error) {
	if opts == nil {
		// opts should never be nil
		opts = DefaultParseOptions
	}
	// look at the first byte
	for i := 0; ; i++ {
		if len(data) == 0 {
			return nil, errDataInvalid
		}
		switch data[0] {
		default:
			// well-known text is not supported yet
			return nil, errDataInvalid
		case 0, 1:
			if i > 0 {
				// 0x00 or 0x01 must be the first bytes
				return nil, errDataInvalid
			}
			// well-known binary is not supported yet
			return nil, errDataInvalid
		case ' ', '\t', '\n', '\r':
			// strip whitespace
			data = data[1:]
			continue
		case '{':
			return parseJSON(data, opts)
		}
	}
}

func toGeometryOpts(opts *ParseOptions) geometry.IndexOptions {
	var gopts geometry.IndexOptions
	if opts == nil {
		gopts = *geometry.DefaultIndexOptions
	} else {
		gopts.Kind = opts.IndexGeometryKind
		gopts.MinPoints = opts.IndexGeometry
	}
	return gopts
}

type parseKeys struct {
	rCoordinates gjson.Result
	rGeometries  gjson.Result
	rGeometry    gjson.Result
	rFeatures    gjson.Result
	members      string // a valid payload with all extra members
	rules        gjson.Result
}

func parseJSON(data string, opts *ParseOptions) (Object, error) {
	if !gjson.Valid(data) {
		return nil, errDataInvalid
	}
	var keys parseKeys
	var fmembers []byte
	var rType gjson.Result
	gjson.Parse(data).ForEach(func(key, val gjson.Result) bool {
		switch key.String() {
		case "type":
			rType = val
		case "coordinates":
			keys.rCoordinates = val
		case "geometries":
			keys.rGeometries = val
		case "geometry":
			keys.rGeometry = val
		case "features":
			keys.rFeatures = val
		case "rules":
			keys.rules = val
		default:
			if len(fmembers) == 0 {
				fmembers = append(fmembers, '{')
			} else {
				fmembers = append(fmembers, ',')
			}
			fmembers = append(fmembers, pretty.UglyInPlace([]byte(key.Raw))...)
			fmembers = append(fmembers, ':')
			fmembers = append(fmembers, pretty.UglyInPlace([]byte(val.Raw))...)
		}
		return true
	})
	if len(fmembers) > 0 {
		fmembers = append(fmembers, '}')
		keys.members = string(fmembers)
	}
	if !rType.Exists() {
		return nil, errTypeMissing
	}
	if rType.Type != gjson.String {
		return nil, errTypeInvalid
	}
	switch rType.String() {
	default:
		return nil, fmt.Errorf(fmtErrTypeIsUnknown, rType.String())
	case "Point":
		return parseJSONPoint(&keys, opts)
	case "LineString":
		return parseJSONLineString(&keys, opts)
	case "Polygon":
		return parseJSONPolygon(&keys, opts)
	case "Feature":
		return parseJSONFeature(&keys, opts)
	case "MultiPoint":
		return parseJSONMultiPoint(&keys, opts)
	case "MultiLineString":
		return parseJSONMultiLineString(&keys, opts)
	case "MultiPolygon":
		return parseJSONMultiPolygon(&keys, opts)
	case "GeometryCollection":
		return parseJSONGeometryCollection(&keys, opts)
	case "FeatureCollection":
		return parseJSONFeatureCollection(&keys, opts)
	}
}

func parseBBoxAndExtras(ex **extra, keys *parseKeys, opts *ParseOptions) error {
	if keys.members == "" {
		return nil
	}
	if *ex == nil {
		*ex = new(extra)
	}
	(*ex).members = keys.members
	return nil
}

func appendJSONPoint(dst []byte, point geometry.Point, ex *extra, idx int) []byte {
	dst = append(dst, '[')
	dst = strconv.AppendFloat(dst, point.X, 'f', -1, 64)
	dst = append(dst, ',')
	dst = strconv.AppendFloat(dst, point.Y, 'f', -1, 64)
	if ex != nil {
		dims := int(ex.dims)
		for i := 0; i < dims; i++ {
			dst = append(dst, ',')
			dst = strconv.AppendFloat(
				dst, ex.values[idx*dims+i], 'f', -1, 64,
			)
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendJSONRules(dst []byte, rules []*Rule) []byte {
	if len(rules) == 0 {
		return dst
	}
	dst = append(dst, `,"rules":`...)
	dst = append(dst, '[')
	for i := 0; i < len(rules); i++ {
		dst = append(dst, '{')
		dst = append(dst, `"id":"`...)
		dst = append(dst, rules[i].id...)
		dst = append(dst, `"`...)
		dst = append(dst, `,"spec":"`...)
		dst = append(dst, rules[i].spec...)
		dst = append(dst, `"`...)
		dst = append(dst, '}')
		if i < len(rules)-1 {
			dst = append(dst, ',')
		}
	}
	dst = append(dst, ']')
	return dst
}

func (ex *extra) appendJSONExtra(dst []byte, propertiesRequired bool) []byte {
	if ex != nil && ex.members != "" {
		dst = append(dst, ',')
		dst = append(dst, ex.members[1:len(ex.members)-1]...)
		if propertiesRequired {
			if !gjson.Get(ex.members, "properties").Exists() {
				dst = append(dst, `,"properties":{}`...)
			}
		}
	} else if propertiesRequired {
		dst = append(dst, `,"properties":{}`...)
	}

	return dst
}

func appendJSONSeries(
	dst []byte, series geometry.Series, ex *extra, pidx int,
) (ndst []byte, npidx int) {
	dst = append(dst, '[')
	nPoints := series.NumPoints()
	for i := 0; i < nPoints; i++ {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = appendJSONPoint(dst, series.PointAt(i), ex, pidx)
		pidx++
	}
	dst = append(dst, ']')
	return dst, pidx
}

func unionRects(a, b geometry.Rect) geometry.Rect {
	if b.Min.X < a.Min.X {
		a.Min.X = b.Min.X
	}
	if b.Max.X > a.Max.X {
		a.Max.X = b.Max.X
	}
	if b.Min.Y < a.Min.Y {
		a.Min.Y = b.Min.Y
	}
	if b.Max.Y > a.Max.Y {
		a.Max.Y = b.Max.Y
	}
	return a
}

func geoDistancePoints(a, b geometry.Point) float64 {
	return geo.DistanceTo(a.Y, a.X, b.Y, b.X)
}
