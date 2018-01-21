package botl

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func ExampleTransform() {
	rawjson, err := ioutil.ReadFile("test1_input.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	rawbotl, err := ioutil.ReadFile("test1.botl")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	var injson map[string]interface{}
	var inbotl map[string]interface{}
	json.Unmarshal(rawjson, &injson)
	json.Unmarshal(rawbotl, &inbotl)
	v, err := Transform(injson, inbotl)
	if err != nil {
		fmt.Errorf("Got error from Transform: %v", err)
	}
	fmt.Println(v)
}
