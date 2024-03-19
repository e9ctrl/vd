package parameter

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type paramType interface {
	uint16 | int | int32 | uint32 | int64 | float32 | float64 | bool | string
}

var (
	ErrValNotFound      = errors.New("value is not found")
	ErrValNotAllowed    = errors.New("value outside opts - ignoring set")
	ErrWrongStringVal   = errors.New("could not convert paramTypeString to string")
	ErrUnknownParamType = errors.New("unknown parameter type")
	ErrWrongIntVal      = errors.New("received param type that cannot be converted to int")
	ErrWrongUintVal     = errors.New("received param type that cannot be converted to uint")
	ErrWrongUint16Val   = errors.New("received param type that cannot be converted to uint16")
	ErrWrongFloatVal    = errors.New("received param type that cannot be converted to float")
	ErrWrongBoolVal     = errors.New("received param type that cannot be converted to bool")
	ErrWrongTypeVal     = errors.New("received value with invalid type")
	ErrUnhandledTypeVal = errors.New("unhandled type")
)

// Parameter interface is responsible for wrapping ConcreteParameter struct, exposing all methods required.
type Parameter interface {
	SetValue(any) error
	Value() any
	String() string
	Opts() []string
	Type() reflect.Kind
}

// ConcreteParameter[T paramType] hold the actual concrete value for each parameter created with New constructor.
type ConcreteParameter[T paramType] struct {
	Parameter
	typ  reflect.Kind
	val  T
	opts []T
	m    sync.RWMutex
}

// as we are using getter and setter it make more sense to have constuctor for Parameter so this can be used outside the module more easily
// e.g.

// Parameter constructor, the constructor will automatically create the ConcreteParameter instance base on the value passed on in the params
func New(val any, opt, typ string) (Parameter, error) {
	switch typ {
	case "byte":
		return newParameter[int32](reflect.Int32, val, opt)
	case "uint16":
		return newParameter[uint16](reflect.Uint16, val, opt)
	case "int16":
		return newParameter[int](reflect.Int, val, opt)
	case "int32":
		return newParameter[int32](reflect.Int32, val, opt)
	case "uint32":
		return newParameter[uint32](reflect.Uint32, val, opt)
	case "int":
		fallthrough
	case "int64":
		return newParameter[int64](reflect.Int64, val, opt)
	case "float32":
		return newParameter[float32](reflect.Float32, val, opt)
	case "float":
		fallthrough
	case "float64":
		return newParameter[float64](reflect.Float64, val, opt)
	case "string":
		return newParameter[string](reflect.String, val, opt)
	case "bool":
		return newParameter[bool](reflect.Bool, val, opt)
	}

	return nil, ErrUnknownParamType
}

// Contructor that fix allowed values and set initial value of the parameter
func newParameter[T paramType](typ reflect.Kind, val any, opt string) (*ConcreteParameter[T], error) {
	opts, err := buildOptions[T](typ, opt)
	if err != nil {
		return nil, err
	}

	param := &ConcreteParameter[T]{
		typ:  typ,
		opts: opts,
	}

	err = param.SetValue(val)
	if err != nil {
		return nil, err
	}

	return param, nil
}

// Value setter
func (p *ConcreteParameter[T]) SetValue(val any) error {
	// Note: the default toml parser will parse the float into float64 and int to int64,
	// direct casting from interface to generice type T will fail here.
	valT, ok := val.(T)
	if !ok {
		valStr, ok := val.(string)
		if !ok {
			return ErrWrongTypeVal
		}

		pVal, err := convertStringToVal[T](p.typ, valStr)
		if err != nil {
			return err
		}

		valT = *pVal
	}

	if len(p.opts) > 0 {
		var isFound bool
		for _, t := range p.opts {
			if t == val {
				isFound = true
				break
			}
		}

		if !isFound {
			return ErrValNotAllowed
		}
	}

	p.m.Lock()
	p.val = valT
	p.m.Unlock()
	return nil
}

// Type getter
func (p *ConcreteParameter[T]) Type() reflect.Kind {
	return p.typ
}

// Value getter
func (p *ConcreteParameter[T]) Value() any {
	p.m.RLock()
	defer p.m.RUnlock()
	return p.val
}

// To String representation
func (p *ConcreteParameter[T]) String() string {
	p.m.RLock()
	defer p.m.RUnlock()
	return fmt.Sprintf("%v", p.val)
}

// Return allowed values if available
func (p *ConcreteParameter[T]) Opts() []string {
	var opts []string
	for _, opt := range p.opts {
		opts = append(opts, fmt.Sprintf("%v", opt))
	}
	return opts
}

// Used mainly while parsing commands received from TCP client
// It converts received string to the corresponding value under parameter.
func convertStringToVal[T paramType](typ reflect.Kind, val string) (*T, error) {
	switch typ {
	case reflect.Uint:
		if uintVal, err := strconv.ParseUint(val, 10, 8); err == nil {
			uintVal8 := uint(uintVal)
			return interface{}(&uintVal8).(*T), nil
		} else {
			return nil, ErrWrongIntVal
		}
	case reflect.Uint16:
		if uintVal, err := strconv.ParseUint(val, 10, 16); err == nil {
			uintVal16 := uint16(uintVal)
			return interface{}(&uintVal16).(*T), nil
		} else {
			return nil, ErrWrongIntVal
		}
	case reflect.Uint32:
		if uintVal, err := strconv.ParseUint(val, 10, 32); err == nil {
			uintVal32 := uint32(uintVal)
			return interface{}(&uintVal32).(*T), nil
		} else {
			return nil, ErrWrongIntVal
		}
	case reflect.Int:
		if intVal, err := strconv.Atoi(val); err == nil {
			return interface{}(&intVal).(*T), nil
		} else if hexVal, err := strconv.ParseInt(val, 16, 64); err == nil {
			// for hex that starts with 0x
			hexValInt := int(hexVal)
			return interface{}(&hexValInt).(*T), nil
		} else if hexVal, err := strconv.ParseInt(val, 0, 64); err == nil {
			// for hex without 0x
			hexValInt := int(hexVal)
			return interface{}(&hexValInt).(*T), nil
		} else {
			return nil, ErrWrongIntVal
		}
	case reflect.Int32:
		if intVal, err := strconv.ParseInt(val, 10, 32); err == nil {
			int32Val := int32(intVal)
			return interface{}(&int32Val).(*T), nil
		} else if hexVal, err := strconv.ParseInt(val, 16, 32); err == nil {
			// for hex that starts with 0x
			hexVal32 := int32(hexVal)
			return interface{}(&hexVal32).(*T), nil
		} else if hexVal, err := strconv.ParseInt(val, 0, 32); err == nil {
			// for hex without 0x
			hexVal32 := int32(hexVal)
			return interface{}(&hexVal32).(*T), nil
		} else {
			return nil, ErrWrongIntVal
		}
	case reflect.Int64:
		if intVal, err := strconv.ParseInt(val, 10, 64); err == nil {
			return interface{}(&intVal).(*T), nil
		} else if hexVal, err := strconv.ParseInt(val, 0, 64); err == nil {
			// for hex that starts with 0x
			return interface{}(&hexVal).(*T), nil
		} else if hexVal, err := strconv.ParseInt(val, 16, 64); err == nil {
			// for hex without 0x
			return interface{}(&hexVal).(*T), nil
		} else {
			return nil, ErrWrongIntVal
		}
	case reflect.Float32:
		if floatVal, err := strconv.ParseFloat(val, 32); err == nil {
			float32Val := float32(floatVal)
			return interface{}(&float32Val).(*T), nil
		} else {
			return nil, ErrWrongFloatVal
		}
	case reflect.Float64:
		if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
			return interface{}(&floatVal).(*T), nil
		} else {
			return nil, ErrWrongFloatVal
		}
	case reflect.Bool:
		var boolVal bool
		switch val {
		case "true":
			boolVal = true
		case "false":
			boolVal = false
		default:
			return nil, ErrWrongBoolVal
		}

		return interface{}(&boolVal).(*T), nil
	default:
		return nil, ErrUnhandledTypeVal
	}
}

// Splits string into parts and tries to convert them according to value type under parameter.
func buildOptions[T paramType](typ reflect.Kind, opt string) ([]T, error) {
	opts := []T{}
	if opt != "" {
		splits := strings.Split(opt, "|")
		switch typ {
		case reflect.Uint:
			for _, val := range splits {
				if uintVal, err := strconv.ParseUint(val, 10, 8); err == nil {
					opts = append(opts, interface{}(uint(uintVal)).(T))
				} else {
					return nil, ErrWrongUintVal
				}
			}
		case reflect.Uint16:
			for _, val := range splits {
				if uintVal, err := strconv.ParseUint(val, 16, 16); err == nil {
					opts = append(opts, interface{}(uint16(uintVal)).(T))
				} else {
					return nil, ErrWrongUint16Val
				}
			}
		case reflect.Int:
			for _, val := range splits {
				if intVal, err := strconv.Atoi(val); err == nil {
					opts = append(opts, interface{}(intVal).(T))
				} else {
					return nil, ErrWrongIntVal
				}
			}
		case reflect.Int32:
			for _, val := range splits {
				if intVal, err := strconv.ParseInt(val, 10, 32); err == nil {
					opts = append(opts, interface{}(int32(intVal)).(T))
				} else {
					return nil, ErrWrongIntVal
				}
			}
		case reflect.Int64:
			for _, val := range splits {
				if intVal, err := strconv.ParseInt(val, 10, 64); err == nil {
					opts = append(opts, interface{}(intVal).(T))
				} else {
					return nil, ErrWrongIntVal
				}
			}
		case reflect.Float32:
			for _, val := range splits {
				if floatVal, err := strconv.ParseFloat(val, 32); err == nil {
					opts = append(opts, interface{}(float32(floatVal)).(T))
				} else {
					return nil, ErrWrongFloatVal
				}
			}
		case reflect.Float64:
			for _, val := range splits {
				if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
					opts = append(opts, interface{}(floatVal).(T))
				} else {
					return nil, ErrWrongFloatVal
				}
			}
		case reflect.String:
			for _, val := range splits {
				opts = append(opts, interface{}(val).(T))
			}
		case reflect.Bool:
			for _, valStr := range splits {
				var val bool
				switch valStr {
				case "true":
					val = true
				case "false":
					val = false
				}
				opts = append(opts, interface{}(val).(T))
			}
		default:
			return opts, ErrUnhandledTypeVal
		}
	}

	return opts, nil
}
