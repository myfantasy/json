package mfj

import "encoding/json"

func MarshalIndentMust(v interface{}, prefix, indent string) string {
	b, err := json.MarshalIndent(v, prefix, indent)
	if err != nil {
		panic(err)
	}
	return string(b)
}
