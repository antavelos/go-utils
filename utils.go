package goutils

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

func IsMap(t any) bool {
	return reflect.ValueOf(t).Kind() == reflect.Map
}

func IsSlice(t any) bool {
	return reflect.ValueOf(t).Kind() == reflect.Slice
}

func IsMapOrSlice(t any) bool {
	return IsMap(t) || IsSlice(t)
}

func IsString(t any) bool {
	return reflect.ValueOf(t).Kind() == reflect.String
}

// flattenArray flattens any array of arrays.
// Example: flattenArray([1, 2, 3, [4, 5, [6, 7, [8, 9]]]]) = [1, 2, 3, 4, 5, 6, 7, 8, 9].
func FlattenArray(arr []any) (result []any) {
	for _, item := range arr {
		if !IsSlice(item) {
			result = append(result, item)
			continue
		}

		flattenedItem := FlattenArray(item.([]any))

		for _, fitem := range flattenedItem {
			result = append(result, fitem)
		}
	}

	return
}

// mapGetDeep returns an array of the values of all the nodes withkey `key`.
// m can be either a map of a slice.
func MapGetDeep(m any, key string) (result []any) {
	if IsMap(m) {
		v, ok := m.(map[string]any)[key]
		if ok {
			result = append(result, v)
			return
		}

		for mkey := range IterMapKeys(m, nil) {
			v, _ := m.(map[string]any)[mkey]
			if IsMapOrSlice(v) {
				if dv := MapGetDeep(v, key); dv != nil {
					result = append(result, dv)
				}
			}
		}
	}

	if IsSlice(m) {
		for _, item := range m.([]any) {
			if v := MapGetDeep(item, key); v != nil {
				result = append(result, v)
			}
		}
	}

	return
}

// mapPutDeep updates the values of all the nodes withey `key`.
// m can be either a map of a slice.
func MapPutDeep(m any, key string, value any) error {
	if IsMap(m) {
		if _, ok := m.(map[string]any)[key]; ok {
			m.(map[string]any)[key] = value
			return nil
		}

		for mkey := range IterMapKeys(m, nil) {
			mvalue, _ := m.(map[string]any)[mkey]
			if IsMapOrSlice(mvalue) {
				MapPutDeep(mvalue, key, value)
			}
		}
	}

	if IsSlice(m) {
		for _, item := range m.([]any) {
			MapPutDeep(item, key, value)
		}
	}
	return nil
}

// mapGetDeepFlattened returns the same as `getDeepFlattened` but the result will be a flattened array.
func MapGetDeepFlattened(m any, key string) []any {
	return FlattenArray(MapGetDeep(m, key))
}

// Holds the numbder of occurences of a item in an array.
type counter map[any]int

// counterFromArray returns a counter object of the elements of a given array.
func counterFromArray(arr []any) counter {
	result := make(counter)
	for _, item := range arr {
		itemx, ok := result[item]
		if !ok {
			result[item] = 1
		} else {
			result[item] = itemx + 1
		}
	}

	return result
}

// compareCounters compares two counter objects.
func compareCounters(c1, c2 counter) bool {
	if len(c1) != len(c2) {
		return false
	}

	for k, v := range c1 {
		if c2[k] != v {
			return false
		}
	}
	return true
}

// sendOrQuit coordinated the value generation of a go routine that generates values.
func sendOrQuit[T any](t T, out chan<- T, quit <-chan struct{}) bool {
	select {
	case out <- t:
		return true
	case <-quit:
		return false
	}
}

// iterMapKeys generates the keys of a given map
func IterMapKeys(m any, quit <-chan struct{}) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)

		vm := reflect.ValueOf(m)
		if vm.Kind() == reflect.Map {
			for _, key := range vm.MapKeys() {
				skey := fmt.Sprintf("%v", key)
				sendOrQuit(skey, out, quit)
			}
		}
	}()

	return out
}

func IterAny(t any, quit <-chan struct{}) <-chan any {
	out := make(chan any)

	go func() {
		defer close(out)

		vm := reflect.ValueOf(t)
		if vm.Kind() == reflect.Slice {
			for i := 0; i < vm.Len(); i++ {
				sendOrQuit(vm.Index(i).Interface(), out, quit)
			}
		}
	}()

	return out
}

// mapHasKey determines whether a given key exists in a given map
func MapHasKey(m any, key string) bool {
	for mkey := range IterMapKeys(m, nil) {
		if mkey == key {
			return true
		}
	}
	return false
}

// toFloat converts any number like value or any string number to float64.
func ToFloat64(value any) (float64, error) {
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return float64(v), nil
	case string:
		fv, err := strconv.ParseFloat(v, 1)
		if err == nil {
			return fv, nil
		}
	}
	return 0, errors.New("Can't convert to float64.")
}

func Prettify(x any) any {
	b, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		return x
	}
	return b
}
