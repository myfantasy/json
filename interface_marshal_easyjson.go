// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package mfj

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson786cc928DecodeGithubComMyfantasyJson(in *jlexer.Lexer, out *IStructView) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "_type":
			out.Type = string(in.String())
		case "data":
			if data := in.Raw(); in.Ok() {
				in.AddError((out.Data).UnmarshalJSON(data))
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson786cc928EncodeGithubComMyfantasyJson(out *jwriter.Writer, in IStructView) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"_type\":"
		out.RawString(prefix[1:])
		out.String(string(in.Type))
	}
	{
		const prefix string = ",\"data\":"
		out.RawString(prefix)
		out.Raw((in.Data).MarshalJSON())
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v IStructView) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson786cc928EncodeGithubComMyfantasyJson(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v IStructView) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson786cc928EncodeGithubComMyfantasyJson(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *IStructView) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson786cc928DecodeGithubComMyfantasyJson(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *IStructView) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson786cc928DecodeGithubComMyfantasyJson(l, v)
}
