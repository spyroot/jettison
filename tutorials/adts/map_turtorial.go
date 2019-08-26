package adts

import (
	"encoding/json"
	"fmt"
	"log"
)

type Pair struct {
	first  string `required max: "100"`
	second string `required max: "100"`
}

func MapTutorialMain() {
	checkMapKey()
	checkStructAsVal()
}

func checkMapKey() {
	log.Println("checkMapKey")
	statePopulation := map[string]int{
		"California": 10000000,
		"Texas":      20000000,
	}

	// valid key
	val, ok := statePopulation["Texas"]
	fmt.Println(val, ok)

	// invalid key // should return 0 and false
	val, ok = statePopulation["New York"]
	fmt.Println(val, ok)
}

func jsonPrint(val map[string]Pair) {
	jsonString, err := json.MarshalIndent(val, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Print(string(jsonString))
}

/**
json serialize and print
*/
func jsonPairPrint(val Pair) {
	jsonString, err := json.MarshalIndent(val, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Print(string(jsonString))
}

/**
store struct in a map
*/
func checkStructAsVal() {

	log.Println("checkStructAsVal")

	pairmap := map[string]Pair{
		"California": {
			"California population",
			"California Capital"},
		"Texas": {
			"first val",
			"second val"},
	}

	val, _ := pairmap["California"]
	jsonPairPrint(val)
	jsonPrint(pairmap)
}
