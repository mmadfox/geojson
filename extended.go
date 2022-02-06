package geojson

import "github.com/tidwall/gjson"

type Lookuper interface {
	Lookup(id string) (Object, bool)
}

type Extended interface {
	ForEachRule(iter func(rule *Rule) bool) bool
}

var _ = []Extended{&FeatureCollection{}, &Feature{}}
var _ = []Lookuper{&FeatureCollection{}, &Feature{}}

// Rule ...
type Rule struct {
	id   string
	spec string
}

func (r Rule) ID() string {
	return r.id
}

func (r Rule) Spec() string {
	return r.spec
}

func WalkRule(root Object, fn func(r *Rule, l Lookuper) bool) bool {
	switch typ := root.(type) {
	default:
		return false
	case *Feature:
		return typ.ForEachRule(func(rule *Rule) bool { return fn(rule, typ) })
	case *FeatureCollection:
		return typ.ForEachRule(func(rule *Rule) bool { return fn(rule, typ) })
	}
}

func parseRules(keys *parseKeys) (rules []*Rule, err error) {
	if !keys.rules.Exists() {
		return nil, nil
	}
	rules = make([]*Rule, 0)
	keys.rules.ForEach(func(key, value gjson.Result) bool {
		if value.Type != gjson.JSON {
			err = errRulesInvalid
			return false
		}
		var rule Rule
		value.ForEach(func(key, value gjson.Result) bool {
			if !value.Exists() {
				err = errRulesInvalid
				return false
			}
			switch key.Str {
			case "id":
				rule.id = value.Str
			case "spec":
				rule.spec = value.Str
			default:
				err = errRulesInvalid
				return false
			}
			return true
		})
		if err != nil {
			return false
		}
		if len(rule.spec) == 0 {
			err = errRulesInvalid
			return false
		}
		if len(rule.id) == 0 {
			err = errRuleIDInvalid
			return false
		}
		rules = append(rules, &rule)
		return true
	})
	return rules, err
}
