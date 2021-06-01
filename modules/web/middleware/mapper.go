package middleware

import "reflect"

// TypeMapper represents an interface for mapping interface{} values based on type.
type TypeMapper interface {
	// Map Maps the interface{} value based on its immediate type from reflect.TypeOf.
	Map(interface{}) TypeMapper
	// MapTo Maps the interface{} value based on the pointer of an Interface provided.
	// This is really only useful for mapping a value as an interface, as interfaces
	// cannot at this time be referenced directly without a pointer.
	MapTo(interface{}, interface{}) TypeMapper
	// Set Provides a possibility to directly insert a mapping based on type and value.
	// This makes it possible to directly map type arguments not possible to instantiate
	// with reflect like unidirectional channels.
	Set(reflect.Type, reflect.Value) TypeMapper
	// GetVal Returns the Value that is mapped to the current type. Returns a zeroed Value if
	// the Type has not been mapped.
	GetVal(reflect.Type) reflect.Value
}

