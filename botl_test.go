package botl

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/clbanning/mxj"
)

type Map map[string]interface{}

// TestBotlTransform is my only test, but given the limited set of test examples
func TestBotlTransform(t *testing.T) {
	rawjson, err := ioutil.ReadFile("testdata/test1_input.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	var injson map[string]interface{}
	json.Unmarshal(rawjson, &injson)

	rawbotl, err := ioutil.ReadFile("testdata/test1.botl")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	var inbotl map[string]interface{}
	json.Unmarshal(rawbotl, &inbotl)

	rawresult, err := ioutil.ReadFile("testdata/test1_output.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	var outjson map[string]interface{}
	json.Unmarshal(rawresult, &outjson)

	v, err := Transform(injson, inbotl)
	if err != nil {
		fmt.Errorf("Got error from Transform: %v", err)
	}

	fixedDeepEqual(t, v, outjson)

}

func fixedDeepEqual(t *testing.T, a interface{}, b interface{}) {
	if reflect.ValueOf(a).Kind() != reflect.ValueOf(b).Kind() {
		t.Fatal("Not same types")
	}
	ta := fmt.Sprintf("%T", a)
	tb := fmt.Sprintf("%T", b)
	switch ta {
	case "mxj.Map":
		ma := a.(mxj.Map)
		var mb Map
		if tb == "map[string]interface {}" {
			mb = Map(b.(map[string]interface{}))
		} else {
			mb = Map(b.(mxj.Map))
		}
		for k, v := range ma {
			if _, ok := mb[k]; !ok {
				t.Fatal("doesn't have the same keys")
			}
			fixedDeepEqual(t, v, mb[k])
		}
		for k := range mb {
			if _, ok := ma[k]; !ok {
				t.Fatal("doesn't have the same keys")
			}
		}
		break
	case "map[string]interface {}":
		ma := Map(a.(map[string]interface{}))
		var mb Map
		if tb == "map[string]interface {}" {
			mb = Map(b.(map[string]interface{}))
		} else {
			mb = Map(b.(mxj.Map))
		}
		for k, v := range ma {
			if _, ok := mb[k]; !ok {
				t.Fatal("doesn't have the same keys")
			}
			fixedDeepEqual(t, v, mb[k])
		}
		for k := range mb {
			if _, ok := ma[k]; !ok {
				t.Fatal("doesn't have the same keys")
			}
		}
		break
	case "[]interface {}":
		ma := a.([]interface{})
		mb := b.([]interface{})
		if ma == nil && mb == nil {
			t.Fatal("one is nil other not")
		}
		if ma == nil || mb == nil {
			return
		}
		if len(ma) != len(mb) {
			t.Fatal("not equal lenght slices")
		}
		for i := range ma {
			fixedDeepEqual(t, ma[i], mb[i])
		}
		break
	default:
		if !reflect.DeepEqual(a, b) {
			t.Fatalf("not equal %+v %+v\n", a, b)
		}
	}
}
