package mfj

import (
	"encoding/json"
	"sync"

	"github.com/myfantasy/mft"
)

var (
	nilValue  = "null"
	trueBool  = "true"
	falseBool = "false"
)

type JsonInterfaceMarshaller interface {
	UnmarshalJSONTypeName() string
}

type JsonUnmarshalObjectGenerate func() JsonInterfaceMarshaller

// StructFactory generate structs for json unmarshal
type StructFactory struct {
	Generators    map[string]JsonUnmarshalObjectGenerate
	GeneratorsNil map[string]JsonUnmarshalObjectGenerate

	mx sync.RWMutex
}

var GlobalStructFactory = &StructFactory{
	Generators:    map[string]JsonUnmarshalObjectGenerate{},
	GeneratorsNil: map[string]JsonUnmarshalObjectGenerate{},
}

func (jsf *StructFactory) Add(name string, generator JsonUnmarshalObjectGenerate) {
	jsf.mx.Lock()
	defer jsf.mx.Unlock()
	jsf.Generators[name] = generator
}
func (jsf *StructFactory) AddNil(name string, generator JsonUnmarshalObjectGenerate) {
	jsf.mx.Lock()
	defer jsf.mx.Unlock()
	jsf.GeneratorsNil[name] = generator
}

func (jsf *StructFactory) Get(name string) (obj interface{}, err *mft.Error) {
	jsf.mx.RLock()
	defer jsf.mx.RUnlock()
	fn, ok := jsf.Generators[name]
	if !ok {
		return nil, mft.ErrorSf("Not found object generator for `%v`", name)
	}
	return fn(), nil
}

func (jsf *StructFactory) GetNil(name string) (obj interface{}, err *mft.Error) {
	jsf.mx.RLock()
	defer jsf.mx.RUnlock()
	fn, ok := jsf.GeneratorsNil[name]
	if !ok {
		return nil, mft.ErrorSf("Not found nil object generator for `%v`", name)
	}
	return fn(), nil
}

//go:generate easyjson interface_marshal.go

//easyjson:json
type IStructView struct {
	Type string          `json:"_type"`
	Data json.RawMessage `json:"data"`
}
