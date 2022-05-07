package golang

import (
	"fmt"
	"github.com/libyarp/idl"
	"strings"
)

func typeForGo(p idl.PrimitiveType) string {
	switch p {
	case idl.Uint8:
		return "uint8"
	case idl.Uint16:
		return "uint16"
	case idl.Uint32:
		return "uint32"
	case idl.Uint64:
		return "uint64"
	case idl.Int8:
		return "int8"
	case idl.Int16:
		return "int16"
	case idl.Int32:
		return "int32"
	case idl.Int64:
		return "int64"
	case idl.Float32:
		return "float32"
	case idl.Float64:
		return "float64"
	case idl.Struct:
		return "struct"
	case idl.Bool:
		return "bool"
	case idl.String:
		return "string"
	}
	return "INVALID?"
}

func composeTypeString(t idl.Type, resolve func(name string) string) string {
	switch t := t.(type) {
	case idl.Primitive:
		return typeForGo(t.Kind)
	case idl.Map:
		k := typeForGo(t.Key)
		v := composeTypeString(t.Value, resolve)
		return "map[" + k + "]" + v
	case idl.Array:
		return "[]" + composeTypeString(t.Of, resolve)
	case idl.Unresolved:
		return resolve(t.Name)
	default:
		panic(fmt.Sprintf("BUG: Unexpected idl type %T", t))
	}
}

func fieldType(f idl.Field, resolve func(string) string) string {
	typeString := composeTypeString(f.Type, resolve)
	if _, ok := f.Annotations.FindByName(idl.RepeatedAnnotation); ok {
		typeString = "[]" + typeString
	} else if _, ok := f.Annotations.FindByName(idl.OptionalAnnotation); ok {
		typeString = "*" + typeString
	}
	return typeString
}

func processType(s *idl.FileSet, p map[string]string, t string) (provider string, typeName string) {
	if s.FromSamePackage(t) {
		return "", t
	}
	pkg, name := idl.SplitComponents(t)
	if pr, ok := p[pkg]; ok {
		url := strings.Split(pr, "/")
		return pkg, url[len(url)-1] + "." + name
	}
	if pkg != "" {
		pkc := strings.Split(pkg, ".")
		return "", pkc[len(pkc)-1] + "." + name
	}
	return "", name
}
