package client

import (
	"encoding/json"
	"reflect"
)

// Wrap wraps a function and translates json input/output for the wrapped function.
// The input function must be of the form:
//
//     in func(*Session,InStruct) (OutStruct,error)
//     in func(*Session,InStruct) error
//     in func(*Session,InStruct)
//     in func(*Session)
//
// InStruct and OutStruct are public structs. Wrap translates the
// input to InStruct, calls the function, and then translates the
// OutStruct to byte array. If either in or outstruct are []byte, they
// are untranslated. If the function returns only error, or nothing,
// those are passed as nil
func Wrap(in interface{}) func(*Session, []byte) ([]byte, error) {
	f := wrap(in, false)
	return func(s *Session, d []byte) ([]byte, error) {
		out, err := f(s, d)
		if err != nil {
			return nil, err
		}
		if out == nil {
			return nil, nil
		}
		return out.([]byte), nil
	}
}

func wrap(in interface{}, preserveOut bool) func(*Session, []byte) (interface{}, error) {
	f := reflect.ValueOf(in)
	fType := f.Type()
	if fType.Kind() != reflect.Func {
		panic("Input to Wrap is not a function")
	}
	numInArgs := fType.NumIn()
	if numInArgs != 1 && numInArgs != 2 {
		panic("Input function must get one or two args")
	}
	firstArgType := fType.In(0)
	// Must be pointer
	if firstArgType != reflect.TypeOf(&Session{}) {
		panic("First arg must be session pointer")
	}
	noInTranslation := false
	var inStructType reflect.Type
	var inStructPtr bool
	if numInArgs == 2 {
		secondArgType := fType.In(1)
		if secondArgType == reflect.TypeOf([]byte{}) {
			noInTranslation = true
		} else {
			// second arg is a struct, or ptr to struct
			if secondArgType.Kind() == reflect.Ptr {
				if secondArgType.Elem().Kind() != reflect.Struct {
					panic("Second arg must be a struct or *struct")
				}
				inStructType = secondArgType.Elem()
				inStructPtr = true
			} else {
				if secondArgType.Kind() != reflect.Struct {
					panic("Second arg must be a struct or *struct")
				}
				inStructType = secondArgType
			}
		}
	}
	return func(s *Session, bin []byte) (interface{}, error) {
		args := make([]reflect.Value, numInArgs)
		args[0] = reflect.ValueOf(s)
		if numInArgs == 2 {
			if noInTranslation {
				args[1] = reflect.ValueOf(bin)
			} else {
				// Instantiate
				arg := reflect.New(inStructType)
				if len(bin) > 0 {
					err := json.Unmarshal(bin, arg.Interface())
					if err != nil {
						return nil, err
					}
				}
				if inStructPtr {
					args[1] = arg
				} else {
					args[1] = arg.Elem()
				}
			}
		}
		out := f.Call(args)
		if len(out) == 0 {
			return nil, nil
		}
		if len(out) == 1 {
			var e error
			if reflect.TypeOf(out[0]) == reflect.TypeOf(e) {
				return nil, out[0].Interface().(error)
			}
			if out[0].Interface() == nil {
				return nil, nil
			}
			if preserveOut {
				return out[0].Interface(), nil
			}
			b, _ := json.Marshal(out[0].Interface())
			return b, nil
		}

		if out[0].Interface() == nil {
			return nil, out[1].Interface().(error)
		}
		if preserveOut {
			return out[0].Interface(), out[1].Interface().(error)
		}
		b, _ := json.Marshal(out[0].Interface())
		return b, out[1].Interface().(error)
	}
}
