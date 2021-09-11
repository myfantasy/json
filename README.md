# json (mfjson)
The mfjson package provides marshal/unmarshal for objects containing interface fields  
`mfjson` generates `<file>.mfjson.go` file. The file contains proxy structs for marshaling struct.

## Attributes
`//mfjson:interface some_name` - Marks struct for add to generator with name `some_name`

`//mfjson:add //easyjson:json` - Add additional comment to proxy class

`mfjson:"true"` - Marks fiels as interfacable. The field will marshal with generator


## Usage
### Run code gen
```sh
# install
go get -u github.com/myfantasy/json/...

# run
mfjson <file>.go
```

### Code
Example for generate:
```go

type TestInt1 interface {
	BlaName() string
}

// Generate registaration struct in generator
//mfjson:interface b_test_struct
type B struct {
	A string
	B float64
}

func (b *B) BlaName() string {
	return "B_BlaName"
}

//mfjson:add //easyjson:json
//mfjson:marshal
type Example struct {
	A int `json:"id"`
	B int64
	C TestInt1                    `json:"c" mfjson:"true"`
}

//mfjson:marshal
type AasList []TestInt1

//mfjson:add //easyjson:json
//mfjson:marshal
type AasMap map[string]TestInt1
```
