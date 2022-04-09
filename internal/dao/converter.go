/*
	Local version of .../pkg/mod/k8s.io/apimachinery@v0.22.3/pkg/runtime/converter.go
*/

package dao

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"

	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/structured-merge-diff/v4/value"
)

type structField struct {
	structType reflect.Type
	field      int
}

type fieldInfo struct {
	name      string
	nameValue reflect.Value
	omitempty bool
}

type fieldsCacheMap map[structField]*fieldInfo

type fieldsCache struct {
	sync.Mutex
	value atomic.Value
}

func newFieldsCache() *fieldsCache {
	cache := &fieldsCache{}
	cache.value.Store(make(fieldsCacheMap))
	return cache
}

var (
	fieldCache = newFieldsCache()
)

func FromUnstructured(sv, dv reflect.Value) error {
	sv = unwrapInterface(sv)
	if !sv.IsValid() {
		dv.Set(reflect.Zero(dv.Type()))
		return nil
	}
	st, dt := sv.Type(), dv.Type()

	switch dt.Kind() {
	case reflect.Map, reflect.Slice, reflect.Ptr, reflect.Struct, reflect.Interface:
		// Those require non-trivial conversion.
	default:
		// This should handle all simple types.
		if st.AssignableTo(dt) {
			dv.Set(sv)
			return nil
		}
		// We cannot simply use "ConvertibleTo", as JSON doesn't support conversions
		// between those four groups: bools, integers, floats and string. We need to
		// do the same.
		if st.ConvertibleTo(dt) {
			switch st.Kind() {
			case reflect.String:
				switch dt.Kind() {
				case reflect.String:
					dv.Set(sv.Convert(dt))
					return nil
				}
			case reflect.Bool:
				switch dt.Kind() {
				case reflect.Bool:
					dv.Set(sv.Convert(dt))
					return nil
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				switch dt.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
					reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					dv.Set(sv.Convert(dt))
					return nil
				case reflect.Float32, reflect.Float64:
					dv.Set(sv.Convert(dt))
					return nil
				}
			case reflect.Float32, reflect.Float64:
				switch dt.Kind() {
				case reflect.Float32, reflect.Float64:
					dv.Set(sv.Convert(dt))
					return nil
				}
				if sv.Float() == math.Trunc(sv.Float()) {
					dv.Set(sv.Convert(dt))
					return nil
				}
			}
			return fmt.Errorf("cannot convert %s to %s", st.String(), dt.String())
		}
	}

	// Check if the object has a custom JSON marshaller/unmarshaller.
	entry := value.TypeReflectEntryOf(dv.Type())
	if entry.CanConvertFromUnstructured() {
		return entry.FromUnstructured(sv, dv)
	}

	switch dt.Kind() {
	case reflect.Map:
		return mapFromUnstructured(sv, dv)
	case reflect.Slice:
		return sliceFromUnstructured(sv, dv)
	case reflect.Ptr:
		return pointerFromUnstructured(sv, dv)
	case reflect.Struct:
		return structFromUnstructured(sv, dv)
	case reflect.Interface:
		return interfaceFromUnstructured(sv, dv)
	default:
		return fmt.Errorf("unrecognized type: %v", dt.Kind())
	}
}

func fieldInfoFromField(structType reflect.Type, field int) *fieldInfo {
	fieldCacheMap := fieldCache.value.Load().(fieldsCacheMap)
	if info, ok := fieldCacheMap[structField{structType, field}]; ok {
		return info
	}

	// Cache miss - we need to compute the field name.
	info := &fieldInfo{}
	typeField := structType.Field(field)
	jsonTag := typeField.Tag.Get("json")
	if len(jsonTag) == 0 {
		// Make the first character lowercase.
		if typeField.Name == "" {
			info.name = typeField.Name
		} else {
			info.name = strings.ToLower(typeField.Name[:1]) + typeField.Name[1:]
		}
	} else {
		items := strings.Split(jsonTag, ",")
		info.name = items[0]
		for i := range items {
			if items[i] == "omitempty" {
				info.omitempty = true
				break
			}
		}
	}
	info.nameValue = reflect.ValueOf(info.name)

	fieldCache.Lock()
	defer fieldCache.Unlock()
	fieldCacheMap = fieldCache.value.Load().(fieldsCacheMap)
	newFieldCacheMap := make(fieldsCacheMap)
	for k, v := range fieldCacheMap {
		newFieldCacheMap[k] = v
	}
	newFieldCacheMap[structField{structType, field}] = info
	fieldCache.value.Store(newFieldCacheMap)
	return info
}

func unwrapInterface(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v
}

func mapFromUnstructured(sv, dv reflect.Value) error {
	st, dt := sv.Type(), dv.Type()
	if st.Kind() != reflect.Map {
		return fmt.Errorf("cannot restore map from %v", st.Kind())
	}

	if !st.Key().AssignableTo(dt.Key()) && !st.Key().ConvertibleTo(dt.Key()) {
		return fmt.Errorf("cannot copy map with non-assignable keys: %v %v", st.Key(), dt.Key())
	}

	if sv.IsNil() {
		dv.Set(reflect.Zero(dt))
		return nil
	}
	dv.Set(reflect.MakeMap(dt))
	for _, key := range sv.MapKeys() {
		value := reflect.New(dt.Elem()).Elem()
		if val := unwrapInterface(sv.MapIndex(key)); val.IsValid() {
			if err := FromUnstructured(val, value); err != nil {
				return err
			}
		} else {
			value.Set(reflect.Zero(dt.Elem()))
		}
		if st.Key().AssignableTo(dt.Key()) {
			dv.SetMapIndex(key, value)
		} else {
			dv.SetMapIndex(key.Convert(dt.Key()), value)
		}
	}
	return nil
}

func sliceFromUnstructured(sv, dv reflect.Value) error {
	st, dt := sv.Type(), dv.Type()
	if st.Kind() == reflect.String && dt.Elem().Kind() == reflect.Uint8 {
		// We store original []byte representation as string.
		// This conversion is allowed, but we need to be careful about
		// marshaling data appropriately.
		if len(sv.Interface().(string)) > 0 {
			marshalled, err := json.Marshal(sv.Interface())
			if err != nil {
				return fmt.Errorf("error encoding %s to json: %v", st, err)
			}
			// TODO: Is this Unmarshal needed?
			var data []byte
			err = json.Unmarshal(marshalled, &data)
			if err != nil {
				return fmt.Errorf("error decoding from json: %v", err)
			}
			dv.SetBytes(data)
		} else {
			dv.Set(reflect.Zero(dt))
		}
		return nil
	}
	if st.Kind() != reflect.Slice {
		return fmt.Errorf("cannot restore slice from %v", st.Kind())
	}

	if sv.IsNil() {
		dv.Set(reflect.Zero(dt))
		return nil
	}
	dv.Set(reflect.MakeSlice(dt, sv.Len(), sv.Cap()))
	for i := 0; i < sv.Len(); i++ {
		if err := FromUnstructured(sv.Index(i), dv.Index(i)); err != nil {
			return err
		}
	}
	return nil
}

func pointerFromUnstructured(sv, dv reflect.Value) error {
	st, dt := sv.Type(), dv.Type()

	if st.Kind() == reflect.Ptr && sv.IsNil() {
		dv.Set(reflect.Zero(dt))
		return nil
	}
	dv.Set(reflect.New(dt.Elem()))
	switch st.Kind() {
	case reflect.Ptr, reflect.Interface:
		return FromUnstructured(sv.Elem(), dv.Elem())
	default:
		return FromUnstructured(sv, dv.Elem())
	}
}

func structFromUnstructured(sv, dv reflect.Value) error {
	st, dt := sv.Type(), dv.Type()
	if st.Kind() != reflect.Map {
		return fmt.Errorf("cannot restore struct from: %v", st.Kind())
	}

	for i := 0; i < dt.NumField(); i++ {
		fieldInfo := fieldInfoFromField(dt, i)
		fv := dv.Field(i)

		if len(fieldInfo.name) == 0 {
			// This field is inlined.
			if err := FromUnstructured(sv, fv); err != nil {
				return err
			}
		} else {
			value := unwrapInterface(sv.MapIndex(fieldInfo.nameValue))
			if value.IsValid() {
				if err := FromUnstructured(value, fv); err != nil {
					return err
				}
			} else {
				fv.Set(reflect.Zero(fv.Type()))
			}
		}
	}
	return nil
}

func interfaceFromUnstructured(sv, dv reflect.Value) error {
	// TODO: Is this conversion safe?
	dv.Set(sv)
	return nil
}
