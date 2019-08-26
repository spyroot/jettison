package adts

import (
	"fmt"
	"reflect"
)

// tag example
type Credential struct {
	Username string `required max: "100"`
	Password string
}

func StructTutorials() {
	// get type and access field username and get it tag
	t := reflect.TypeOf(Credential{})
	field, _ := t.FieldByName("Username")
	fmt.Println(field.Tag)
}
