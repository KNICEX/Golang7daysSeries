package schema

import (
	"fmt"
	"geeorm/dialect"
	"reflect"
	"testing"
)

type User struct {
	Name string `geeorm:"primary key"`
	Age  int
}

var TestDial, _ = dialect.GetDialect("sqlite3")

func TestParse(t *testing.T) {
	schema := Parse("&User{}", TestDial)
	if schema.Name != "User" || len(schema.Fields) != 2 {
		t.Fatal("failed to parse User struct")
	}

	if schema.GetField("Name").Tag != "primary key" {
		t.Fatal("failed to parse tag")
	}
}

func TestReflect(t *testing.T) {
	user := User{Name: "Alice", Age: 100}
	userPtr := &User{Name: "Chtholly", Age: 17}
	userValue := reflect.ValueOf(user)
	slice := []User{{Name: "Arcueid", Age: 300}, {Name: "Hello", Age: 19}}
	slicePtr := &slice
	sliceValue := reflect.ValueOf(slice)
	printValue("sliceValue", sliceValue)
	slicePtrValue := reflect.ValueOf(slicePtr)
	printValue("slicePtrValue", slicePtrValue)
	sliceType := reflect.TypeOf(slice)
	printType("sliceType", sliceType)
	slicePtrType := reflect.TypeOf(slicePtr)
	printType("slicePtrType", slicePtrType)

	printValue("userValue", userValue)
	userPtrValue := reflect.ValueOf(userPtr)
	printValue("userPtrValue", userPtrValue)
	userType := reflect.TypeOf(user)
	printType("userType", userType)
	userPtrType := reflect.TypeOf(userPtr)
	printType("userPtrType", userPtrType)
}

func printType(pre string, p reflect.Type) {
	fmt.Printf("%s type: %v kind: %v elem: %v name: %v \n",
		pre, p.String(), p.Kind(), transElem4Type(p), p.Name())
}

func printValue(pre string, v reflect.Value) {
	fmt.Printf("%s type: %v value: %v kind: %v elem: %v\n",
		pre, v.Type(), v.Interface(), v.Kind(), transElem4Value(v))
}

func transElem4Value(v reflect.Value) string {
	// Value.Elem() ptr -> val
	if v.Kind() == reflect.Ptr {
		return fmt.Sprintf("%v", v.Elem())
	}
	return "not ptr"
}

func transElem4Type(t reflect.Type) string {
	// Type.Elem() ptrType -> valType  []Type -> Type
	if t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
		return fmt.Sprintf("%v", t.Elem())
	}
	return "not ptr"
}
