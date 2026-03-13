package jsonx

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
)

type JSONValue struct {
	json.RawMessage
	// do not use *gjson.Result directly, use Result() instead
	res *gjson.Result `json:"-"`
}

type Number interface {
	~int64 | ~float64
}

func (jv *JSONValue) Result() gjson.Result {
	if jv.res != nil {
		return *jv.res
	}
	res := gjson.ParseBytes(jv.RawMessage)
	jv.res = &res
	return res
}

func Marshal(v any) (*JSONValue, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	res := gjson.ParseBytes(b)
	return &JSONValue{RawMessage: json.RawMessage(res.Raw), res: &res}, nil
}

func NewEmpty() *JSONValue {
	b := []byte(`""`)
	res := gjson.ParseBytes(b)
	return &JSONValue{RawMessage: json.RawMessage(res.Raw), res: &res}
}

func NewNumber[T Number](n T) *JSONValue {
	b := fmt.Append(nil, n)
	res := gjson.ParseBytes(b)

	return &JSONValue{RawMessage: json.RawMessage(res.Raw), res: &res}
}

func NewString(s string) (*JSONValue, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	res := gjson.ParseBytes(b)

	return &JSONValue{RawMessage: json.RawMessage(res.Raw), res: &res}, nil
}

func NewBool(bo bool) *JSONValue {
	b := []byte("false")
	if bo {
		b = []byte("true")
	}
	res := gjson.ParseBytes(b)

	return &JSONValue{RawMessage: json.RawMessage(res.Raw), res: &res}
}

/*
export wrapped methods to access gjson.Result
*/

func (jv *JSONValue) IsNumber() bool {
	return jv.Result().Type == gjson.Number
}

func (jv *JSONValue) IsString() bool {
	return jv.Result().Type == gjson.String
}

func (jv *JSONValue) IsNull() bool {
	return jv.Result().Type == gjson.Null
}

func (jv *JSONValue) IsBool() bool {
	return jv.Result().IsBool()
}

func (jv *JSONValue) IsObject() bool {
	return jv.Result().IsObject()
}

func (jv *JSONValue) IsArray() bool {
	return jv.Result().IsArray()
}

func (jv *JSONValue) Bool() bool {
	return jv.Result().Bool()
}

func (jv JSONValue) String() string {
	return jv.Result().String()
}

func (jv *JSONValue) Int() int64 {
	return jv.Result().Int()
}

func (jv *JSONValue) Num() float64 {
	return jv.Result().Num
}

func (jv *JSONValue) Map() map[string]JSONValue {
	m := jv.Result().Map()
	res := make(map[string]JSONValue, len(m))
	for k, v := range m {
		res[k] = JSONValue{RawMessage: json.RawMessage(v.Raw), res: &v}
	}
	return res
}

func (jv *JSONValue) Get(key string) JSONValue {
	v := jv.Result().Get(key)
	return JSONValue{RawMessage: json.RawMessage(v.Raw), res: &v}
}
