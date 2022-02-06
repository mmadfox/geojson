package geojson

import (
	"strings"

	"github.com/tidwall/gjson"
)

// FeatureCollection ...
type FeatureCollection struct {
	collection
	rules      []*Rule
	indexByIDs map[string]Object
}

// NewFeatureCollection ...
func NewFeatureCollection(features []Object) *FeatureCollection {
	g := new(FeatureCollection)
	g.children = features
	g.parseInitRectIndex(DefaultParseOptions)
	return g
}

// AppendJSON appends the GeoJSON reprensentation to dst
func (g *FeatureCollection) AppendJSON(dst []byte) []byte {
	dst = append(dst, `{"type":"FeatureCollection","features":[`...)
	for i := 0; i < len(g.children); i++ {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = g.children[i].AppendJSON(dst)
	}
	dst = append(dst, ']')
	if g.extra != nil {
		dst = g.extra.appendJSONExtra(dst, false)
	}
	dst = appendJSONRules(dst, g.rules)
	dst = append(dst, '}')
	strings.Index("", " ")
	return dst
}

// Lookup ...
func (g *FeatureCollection) Lookup(id string) (Object, bool) {
	f, ok := g.indexByIDs[id]
	if ok {
		return f, true
	}
	return nil, false
}

// ForEachRule ...
func (g *FeatureCollection) ForEachRule(iter func(rule *Rule) bool) bool {
	if len(g.rules) == 0 {
		return true
	}
	for i := 0; i < len(g.rules); i++ {
		if ok := iter(g.rules[i]); !ok {
			return false
		}
	}
	return true
}

// String ...
func (g *FeatureCollection) String() string {
	return string(g.AppendJSON(nil))
}

// JSON ...
func (g *FeatureCollection) JSON() string {
	return string(g.AppendJSON(nil))
}

// MarshalJSON ...
func (g *FeatureCollection) MarshalJSON() ([]byte, error) {
	return g.AppendJSON(nil), nil
}

func parseJSONFeatureCollection(
	keys *parseKeys, opts *ParseOptions,
) (Object, error) {
	var g FeatureCollection
	if !keys.rFeatures.Exists() {
		return nil, errFeaturesMissing
	}
	if !keys.rFeatures.IsArray() {
		return nil, errFeaturesInvalid
	}
	var err error
	keys.rFeatures.ForEach(func(key, value gjson.Result) bool {
		var f Object
		f, err = Parse(value.Raw, opts)
		if err != nil {
			return false
		}
		feature, ok := f.(*Feature)
		if ok && len(feature.id) > 0 {
			if g.indexByIDs == nil {
				g.indexByIDs = make(map[string]Object)
			}
			g.indexByIDs[feature.id] = f
		}
		g.children = append(g.children, f)
		return true
	})
	if err != nil {
		return nil, err
	}
	if err := parseBBoxAndExtras(&g.extra, keys, opts); err != nil {
		return nil, err
	}
	rules, err := parseRules(keys)
	if err != nil {
		return nil, err
	}
	g.parseInitRectIndex(opts)
	g.rules = rules
	return &g, nil
}
