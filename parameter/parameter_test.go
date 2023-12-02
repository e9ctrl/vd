package parameter

import (
	"errors"
	"reflect"
	"testing"
)

func TestSetValue(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		typ     string
		initVal any
		val     any
		opt     string
		exp     any
		expErr  error
	}{
		{"string param", "string", "init", "test", "", "test", nil},
		{"string param with opts", "string", "one", "two", "one|two", "two", nil},
		{"string param with wrong opts", "string", "one", "two", "one|three", "one", ErrValNotAllowed},
		{"string param wrong value", "string", "init", 50.0, "", "init", ErrWrongTypeVal},

		{"int param", "int", 20, 50, "", 50, nil},
		{"int param with opts", "int", 60, 50, "50|60", 50, nil},
		{"int param int64", "int64", int64(30), int64(50), "", int64(50), nil},
		{"int param int32", "int", int32(30), int32(50), "", int32(50), nil},
		{"int param string value", "int", 20, "50", "", 50, nil},
		{"int param wrong type", "int", false, 50, "", false, ErrWrongTypeVal},
		{"int param wrong string value", "int", 30, "test", "", 30, ErrWrongIntVal},
		{"int param float value", "int", 30, 50.0, "", 30, ErrWrongTypeVal},

		{"float param", "float64", 20.0, 50.0, "", 50.0, nil},
		{"float param with opts", "float32", 60.0, 50.0, "50|60", 50.0, nil},
		{"float param float32", "float32", float32(20.0), float32(50.0), "", float32(50.0), nil},
		{"float param float64", "float64", 20.0, 50.0, "", 50.0, nil},
		{"float param string value", "float32", 20.0, "50.0", "", 50.0, nil},
		{"float param wrong type", "float32", false, 50.0, "", false, ErrWrongTypeVal},
		{"float param wrong string value", "float32", 20.0, "test", "", 20.0, ErrWrongFloatVal},
		{"float param int value", "float32", 20.0, 50, "", 20.0, ErrWrongTypeVal},

		{"bool param", "bool", true, false, "", false, nil},
		{"bool param with opts", "bool", false, true, "true|false", true, nil},
		{"bool param string value", "bool", false, "true", "", true, nil},
		{"bool param wrong type", "bool", 20.0, true, "", 20.0, ErrWrongTypeVal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			param, err := New(tt.initVal, tt.opt, tt.typ)
			if err != nil {
				t.Error(err)
				return
			}

			err = param.SetValue(tt.val)
			if !errors.Is(err, tt.expErr) {
				t.Errorf("%s: exp error: %v got %v\n", tt.name, tt.expErr, err)
			}

			if param.Value() != tt.exp {
				t.Errorf("%s: exp value: %v got %v\n", tt.name, tt.exp, param.Value())
			}
		})
	}
}

func TestConvertStringToVal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		val  string
		typ  reflect.Kind
		exp  any
	}{
		{"convert int", "50", reflect.Int, 50},
		{"convert int32", "50", reflect.Int32, int32(50)},
		{"convert int64", "50", reflect.Int64, int64(50)},
		{"convert float32", "50.0", reflect.Float32, float32(50)},
		{"convert float64", "50.0", reflect.Float64, float64(50)},
		{"convert bool", "true", reflect.Bool, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got any
			switch tt.typ {
			case reflect.Int:
				res, err := convertStringToVal[int](tt.typ, tt.val)
				if err != nil {
					t.Errorf("unexpected err: %s", err)
				}

				got = *res
			case reflect.Int32:
				res, err := convertStringToVal[int32](tt.typ, tt.val)
				if err != nil {
					t.Errorf("unexpected err: %s", err)
				}

				got = *res
			case reflect.Int64:
				res, err := convertStringToVal[int64](tt.typ, tt.val)
				if err != nil {
					t.Errorf("unexpected err: %s", err)
				}

				got = *res
			case reflect.Float32:
				res, err := convertStringToVal[float32](tt.typ, tt.val)
				if err != nil {
					t.Errorf("unexpected err: %s", err)
				}

				got = *res
			case reflect.Float64:
				res, err := convertStringToVal[float64](tt.typ, tt.val)
				if err != nil {
					t.Errorf("unexpected err: %s", err)
				}

				got = *res
			case reflect.Bool:
				res, err := convertStringToVal[bool](tt.typ, tt.val)
				if err != nil {
					t.Errorf("unexpected err: %s", err)
				}

				got = *res
			}

			if got != tt.exp {
				t.Errorf("%s: exp value: %v got %v\n", tt.name, tt.exp, got)
			}
		})
	}
}

func TestBuildOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		conc string
		typ  reflect.Kind
		exp  []any
	}{
		{"build int options", "1|2|3", reflect.Int, []any{1, 2, 3}},
		{"build string options", "one|two|three", reflect.String, []any{"one", "two", "three"}},
		{"build bool options", "false|true", reflect.Bool, []any{false, true}},
		{"empty options", "", reflect.Float64, []any{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []any
			switch tt.typ {
			case reflect.Int:
				res, err := buildOptions[int](tt.typ, tt.conc)
				if err != nil {
					t.Errorf("unexpected err: %s", err)
				}

				for _, val := range res {
					got = append(got, any(val))
				}
			case reflect.String:
				res, err := buildOptions[string](tt.typ, tt.conc)
				if err != nil {
					t.Errorf("unexpected err: %s", err)
				}

				for _, val := range res {
					got = append(got, any(val))
				}
			case reflect.Float64:
				res, err := buildOptions[float64](tt.typ, tt.conc)
				if err != nil {
					t.Errorf("unexpected err: %s", err)
				}

				for _, val := range res {
					got = append(got, any(val))
				}
			case reflect.Bool:
				res, err := buildOptions[bool](tt.typ, tt.conc)
				if err != nil {
					t.Errorf("unexpected err: %s", err)
				}

				for _, val := range res {
					got = append(got, any(val))
				}

			}

			if len(got) != len(tt.exp) {
				t.Errorf("%s: exp length: %v got %v\n", tt.name, len(tt.exp), len(got))
			}

			equal := true
			for i, val := range got {
				if val != tt.exp[i] {
					equal = false
					break
				}
			}

			if !equal {
				t.Errorf("%s: exp value: %v got %v\n", tt.name, tt.exp, got)
			}
		})
	}
}
