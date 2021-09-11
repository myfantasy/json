package testex

import (
	mfj "github.com/myfantasy/json"
	log "github.com/sirupsen/logrus"
)

type A struct {
	A int
	B float64
}

//mfjson:interface
type B struct {
	A string
	B float64
}

//mfjson:interface c
type C struct {
	A string
	B string
}

type TestInt1 interface {
	mfj.JsonInterfaceMarshaller
	BlaName() string
}

type TestInt2 interface {
	UnmarshalJSONTypeName() string
	BlaName() string
}

func (a *A) UnmarshalJSONTypeName() string {
	return "A"
}
func init() {
	mfj.GlobalStructFactory.Add("A", func() mfj.JsonInterfaceMarshaller { return &A{} })
	mfj.GlobalStructFactory.AddNil("A", func() mfj.JsonInterfaceMarshaller {
		var out *A
		return out
	})
}

func (a *A) BlaName() string {
	return "A_BlaName"
}
func (b *B) BlaName() string {
	return "B_BlaName"
}
func (b *C) BlaName() string {
	return "C_BlaName"
}

//mfjson:marshal
type Example struct {
	A int `json:"id"`
	B int64
	C TestInt1                    `json:"c" mfjson:"true"`
	D mfj.JsonInterfaceMarshaller `json:"d" mfjson:"true"`
	E TestInt2                    `json:"e" mfjson:"true"`
	F log.Level
	G []TestInt2          `json:"g" mfjson:"true"`
	H map[string]TestInt2 `json:"h" mfjson:"true"`
	I []TestInt2
	J map[int]TestInt2
	K []log.Level
	L map[int]log.Level
	M [][]*log.Level
	N *log.Level
	O [4]*log.Level
	P TestInt2 `json:"p" mfjson:"true"`
	Q TestInt2 `json:"q" mfjson:"true"`
}
