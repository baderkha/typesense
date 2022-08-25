package reflection

import "reflect"

// GetTypeName : gets name of the type
func GetTypeName(myvar interface{}) (name string, isPointer bool) {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		return t.Elem().Name(), true
	} else {
		return t.Name(), false
	}
}
