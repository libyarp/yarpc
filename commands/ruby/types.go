package ruby

import (
	"fmt"
	"github.com/libyarp/idl"
	"github.com/libyarp/yarpc/common"
)

func typeForRuby(p idl.PrimitiveType) string {
	switch p {
	case idl.Uint8:
		return ":uint8"
	case idl.Uint16:
		return ":uint16"
	case idl.Uint32:
		return ":uint32"
	case idl.Uint64:
		return ":uint64"
	case idl.Int8:
		return ":int8"
	case idl.Int16:
		return ":int16"
	case idl.Int32:
		return ":int32"
	case idl.Int64:
		return ":int64"
	case idl.Float32:
		return ":float32"
	case idl.Float64:
		return ":float64"
	case idl.Struct:
		return ":struct"
	case idl.Bool:
		return ":bool"
	case idl.String:
		return ":string"
	}
	return "INVALID?"
}

func composeTypeString(t idl.Type, resolve func(name string) string) string {
	switch t := t.(type) {
	case idl.Primitive:
		return typeForRuby(t.Kind)
	case idl.Map:
		k := typeForRuby(t.Key)
		v := composeTypeString(t.Value, resolve)
		return "Yarp::Proto::Map[" + k + ", " + v + "]"
	case idl.Array:
		return "Yarp::Proto::Array[" + composeTypeString(t.Of, resolve) + "]"
	case idl.Unresolved:
		return resolve(t.Name)
	default:
		panic(fmt.Sprintf("BUG: Unexpected idl type %T", t))
	}
}

func composeClassField(f idl.Field, resolve func(name string) string) string {
	switch t := f.Type.(type) {
	case idl.Primitive:
		str := ""
		if _, ok := f.Annotations.FindByName(idl.RepeatedAnnotation); ok {
			str = fmt.Sprintf("array :%s, %d, of: %s", f.Name, f.Index, typeForRuby(t.Kind))
		} else {
			str = fmt.Sprintf("primitive :%s, %s, %d", f.Name, typeForRuby(t.Kind), f.Index)
		}
		if _, ok := f.Annotations.FindByName(idl.OptionalAnnotation); ok {
			str += ", optional: true"
		}
		return str
	case idl.Map:
		k := typeForRuby(t.Key)
		v := composeTypeString(t.Value, resolve)
		return fmt.Sprintf("map :%s, %d, key: %s, value: %s", f.Name, f.Index, k, v)
	case idl.Array:
		return fmt.Sprintf("array :%s, %d, of: %s", f.Name, f.Index, composeTypeString(t.Of, resolve))
	case idl.Unresolved:
		str := ""
		if _, ok := f.Annotations.FindByName(idl.RepeatedAnnotation); ok {
			str = fmt.Sprintf("array :%s, %d, of: %s", f.Name, f.Index, resolve(t.Name))
		} else {
			str = fmt.Sprintf("struct :%s, %s, %d", f.Name, resolve(t.Name), f.Index)
		}
		if _, ok := f.Annotations.FindByName(idl.OptionalAnnotation); ok {
			str += ", optional: true"
		}
		return str
	default:
		panic(fmt.Sprintf("BUG: Unexpected idl type %T", t))
	}
}

func composeOneofField(b *common.Builder, f idl.OneOfField, resolver func(n string) string) {
	b.Printfln("oneof(%d) do", f.Index)
	b.Indent(func(b *common.Builder) {
		for _, v := range f.Items {
			switch t := v.(type) {
			case idl.Field:
				b.Printfln(composeClassField(t, resolver))
			case idl.OneOfField:
				composeOneofField(b, f, resolver)
			}
		}
	})
	b.Printfln("end")
}
