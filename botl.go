// Package botl is a bOTL Object Transformation Language parser
// Go Port of https://github.com/emlynoregan/bOTL_js/
// See http://emlynoregan.github.io/bOTL_js/
// This is a Go implementation of bOTL v3.
package botl

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/gdey/jsonpath"
)

// Transform Maps, Arrays, Simple Values (MAS)
func Transform(aMASSource interface{}, abOTLTransform map[string]interface{}) (interface{}, error) {

	lscope := map[string]interface{}{
		"$": aMASSource,
		"@": aMASSource,
	}

	results, _, err := transform(lscope, abOTLTransform)
	return results, err
}

func transform(aScope interface{}, abOTLTransform interface{}) (interface{}, bool, error) {
	return getSectionFunction(abOTLTransform)(aScope)
}

//return the correct section func to use on the top level of this transform
//Section functions take a scope and return a results list, keep-nils and error
func getSectionFunction(abOTLTransform interface{}) func(interface{}) (interface{}, bool, error) {

	if isObject(abOTLTransform) {
		t := abOTLTransform.(map[string]interface{})
		if ltype, ok := t["_type"]; ok {
			if ltype == "#" {
				return func(aScope interface{}) (interface{}, bool, error) {
					return evalFullSection(aScope, t)
				}
			} else if ltype == "object" {
				return func(aScope interface{}) (interface{}, bool, error) {
					return evalObjectSection(aScope, t)
				}
			} else if ltype == "literal" {
				return func(aScope interface{}) (interface{}, bool, error) {
					return evalLiteralSection(aScope, t)
				}
			}
			return func(aScope interface{}) (interface{}, bool, error) {
				return evalObject(aScope, t)
			}

		}
		return func(aScope interface{}) (interface{}, bool, error) {
			return evalObject(aScope, t)
		}
	} else if isArray(abOTLTransform) {
		return func(aScope interface{}) (interface{}, bool, error) {
			return evalList(aScope, abOTLTransform.([]interface{}))
		}
	} else if isString(abOTLTransform) {
		t := abOTLTransform.(string)
		if t[:1] == "#" {
			return func(aScope interface{}) (interface{}, bool, error) {
				return evalFullSection(aScope, map[string]interface{}{
					"_type": "#",
					"path":  t[1:],
				})
			}
		}
		return func(aScope interface{}) (interface{}, bool, error) {
			return evalLiteral(abOTLTransform)
		}

	}
	return func(aScope interface{}) (interface{}, bool, error) {
		return evalLiteral(abOTLTransform)
	}

}

// eval functions must return
// {'results': <a list of results>, 'keepnils': <boolean>}
func evalLiteral(abOTLLiteral interface{}) (interface{}, bool, error) {
	return []interface{}{abOTLLiteral}, true, nil
}

func evalFullSection(aScope interface{}, abOTLSection map[string]interface{}) (interface{}, bool, error) {
	lpath, ok := abOTLSection["path"]
	if !ok {
		lpath = "@"
	}
	lscopeid, _ := abOTLSection["scope"]
	ltransform, _ := abOTLSection["transform"]
	lniltransform, _ := abOTLSection["niltransform"]
	lkeepnils, _ := abOTLSection["nils"]
	if lkeepnils == nil {
		lkeepnils = true
	}

	//1: eval the path
	var lscope interface{}
	typeoflpath := reflect.ValueOf(lpath).Kind()
	switch typeoflpath {
	case reflect.Uint8:
		lpath = strconv.FormatUint(uint64(lpath.(uint8)), 10)
	}
	lp := lpath.(string)
	if lp[:1] == "$" {
		lscope = aScope.(map[string]interface{})["$"]
	} else if lp[:1] == "@" {
		lscope = aScope.(map[string]interface{})
	} else {
		lscope = aScope
	}
	jpq, err := jsonpath.Parse(lp)
	if err != nil {
		fmt.Println("can't parse", lp)
		panic("too bad")
	}
	lselections, _ := jpq.Apply(lscope)

	var lresults []interface{}

	if isArray(lselections) && len(lselections.([]interface{})) > 0 {
		// here we've got some results
		if ltransform != nil {
			// need to transform all items in selection
			for aSelection := range lselections.([]interface{}) {
				// set up scope for transform, store
				// settings for restoration afterwards
				lPreviosAt := aScope.(map[string]interface{})["@"]
				var lPreviosScopeID interface{}
				aScope.(map[string]interface{})["@"] = lselections.([]interface{})[aSelection]
				if lscopeid != nil {
					lPreviosScopeID = aScope.(map[string]interface{})[lscopeid.(string)]
					aScope.(map[string]interface{})[lscopeid.(string)] = lselections.([]interface{})[aSelection]
				}

				lchildResults, _, _ := transform(aScope, ltransform.(map[string]interface{}))

				lresults = append(lresults, lchildResults)

				// now restore the scope
				aScope.(map[string]interface{})["@"] = lPreviosAt
				if lscopeid != nil {
					aScope.(map[string]interface{})[lscopeid.(string)] = lPreviosScopeID
				}
			}
		} else {
			lresults = lselections.([]interface{})
		}
	} else if isString(lselections) {
		lresults = []interface{}{lselections}
	} else {
		// here we've got no results
		if lniltransform != nil {
			v, _, err := transform(aScope, lniltransform.(map[string]interface{}))
			if err != nil {
				lresults = []interface{}{v}
			}
		} else {
			lresults = []interface{}{}
		}
	}

	if lkeepnils != true {
		var lfiltered []interface{}

		for aResult := range lresults {
			if lresults[aResult] != nil {
				lfiltered = append(lfiltered, lresults[aResult])
			}
		}

		lresults = lfiltered
	}

	return lresults, lkeepnils.(bool), err
}

func evalObjectSection(aScope interface{}, abOTLObjectSection interface{}) (interface{}, bool, error) {
	var retval []interface{}

	lvalue, _ := abOTLObjectSection.(map[string]interface{})["value"]

	if isObject(lvalue) {
		lresult := map[string]interface{}{}
		for lkey := range lvalue.(map[string]interface{}) {
			var litem interface{}
			if lkey == "_type" {
				litem = lvalue.(map[string]interface{})[lkey]
			} else {
				v, _, err := transform(aScope, lvalue.(map[string]interface{})[lkey].(map[string]interface{}))
				if err != nil {
					litem = v.([]interface{})[0]
				} else {
					litem = nil
				}
			}

			if litem != nil {
				lresult[lkey] = litem
			}
		}

		retval = []interface{}{lresult}
	}

	return retval, true, nil
}

func evalLiteralSection(aScope interface{}, abOTLObjectSection interface{}) (interface{}, bool, error) {
	var v map[string]interface{}
	if abOTLObjectSection.(map[string]interface{})["value"] != nil {
		v = map[string]interface{}{abOTLObjectSection.(map[string]interface{})["value"].(string): nil}
	}
	return v, true, nil
}

func evalObject(aScope interface{}, abOTLObject map[string]interface{}) (interface{}, bool, error) {
	ltransformedObject := map[string]interface{}{}
	for lkey := range abOTLObject {
		ltransformedValue, keepnils, _ := transform(aScope, abOTLObject[lkey])
		if ltransformedValue != nil {
			if isArray(ltransformedValue) && len(ltransformedValue.([]interface{})) > 0 {
				val := ltransformedValue.([]interface{})[0]
				if val != nil || keepnils {
					ltransformedObject[lkey] = val
				}
			} else if isObject(ltransformedValue) {
				val := ltransformedValue.(map[string]interface{})
				if val != nil || keepnils {
					ltransformedObject[lkey] = val
				}
			}
		}
	}

	return ltransformedObject, true, nil
}

func evalList(aScope interface{}, abOTLList []interface{}) (interface{}, bool, error) {
	lresults := []interface{}{}
	for aItem := range abOTLList {
		lresult, _, _ := transform(aScope, abOTLList[aItem])
		lresults = append(lresults, lresult)
	}
	return lresults, true, nil
}

func isObject(obj interface{}) bool {
	return reflect.ValueOf(obj).Kind() == reflect.Map
}

func isArray(obj interface{}) bool {
	return reflect.ValueOf(obj).Kind() == reflect.Slice
}

func isString(obj interface{}) bool {
	return reflect.ValueOf(obj).Kind() == reflect.String
}
