package pretty

import (
	"encoding/json"
)

func pretty(v interface{}) []byte {
	pretty, _ := json.MarshalIndent(v, "", "	")

	return pretty
}

func String(v interface{}) string { return string(pretty(v)) }

func Bytes(v interface{}) []byte { return pretty(v) }

type stringer func() string

func (s stringer) String() string { return s() }

func Stringer(v interface{}) stringer { return func() string { return String(v) } }
