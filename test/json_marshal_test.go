package testex

import (
	"encoding/json"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestMarshalAndUnmarshal(t *testing.T) {
	var p *A
	exIn := Example{
		A: 5,
		B: 6,
		C: &A{
			A: 1,
			B: 2,
		},
		D: &B{
			A: "234",
			B: 4.5,
		},
		E: &C{
			A: "667",
			B: "668",
		},
		F: log.WarnLevel,
		G: []TestInt2{
			&A{5, 7},
			p,
			&C{"x", "y"},
		},
		H: map[string]TestInt2{
			"o": &A{9, 7},
			"p": p,
			"q": &C{"x12", "y32"},
		},
		P: p,
	}

	ll := log.ErrorLevel
	exIn.O[3] = &ll

	body, err := json.Marshal(exIn)
	if err != nil {
		t.Fatal(err)
	}

	exOut := &Example{}
	err = json.Unmarshal(body, &exOut)
	if err != nil {
		t.Fatal(err)
	}

	body2, err := json.Marshal(exOut)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != string(body2) {
		t.Fatalf("C: `%v`\nP: `%v`\n\n%v\n\n%v", exOut.C, exOut.P,
			string(body), string(body2))
	}
}

func TestMarshalAndUnmarshalArray(t *testing.T) {
	var p *A
	exIn := make(AasList, 3)
	exIn[0] = &A{2, 4}
	exIn[1] = &B{"32", 34}
	exIn[2] = p

	body, err := json.Marshal(exIn)
	if err != nil {
		t.Fatal(err)
	}

	exOut := AasList{}
	err = json.Unmarshal(body, &exOut)
	if err != nil {
		t.Fatal(err)
	}

	body2, err := json.Marshal(exOut)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != string(body2) {
		t.Fatalf("%v\n\n%v",
			string(body), string(body2))
	}
}

func TestMarshalAndUnmarshalArrayNil(t *testing.T) {
	var exIn AasList

	body, err := json.Marshal(exIn)
	if err != nil {
		t.Fatal(err)
	}

	exOut := AasList{}
	err = json.Unmarshal(body, &exOut)
	if err != nil {
		t.Fatal(err)
	}

	body2, err := json.Marshal(exOut)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != string(body2) {
		t.Fatalf("%v\n\n%v",
			string(body), string(body2))
	}
}

func TestMarshalAndUnmarshalMap(t *testing.T) {
	var p *A
	exIn := make(AasMap, 3)
	exIn["a"] = &A{2, 4}
	exIn["b"] = &B{"32", 34}
	exIn["c"] = p

	body, err := json.Marshal(exIn)
	if err != nil {
		t.Fatal(err)
	}

	exOut := AasMap{}
	err = json.Unmarshal(body, &exOut)
	if err != nil {
		t.Fatal(err)
	}

	body2, err := json.Marshal(exOut)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != string(body2) {
		t.Fatalf("%v\n\n%v",
			string(body), string(body2))
	}
}

func TestMarshalAndUnmarshalMapNil(t *testing.T) {
	var exIn AasMap

	body, err := json.Marshal(exIn)
	if err != nil {
		t.Fatal(err)
	}

	exOut := AasMap{}
	err = json.Unmarshal(body, &exOut)
	if err != nil {
		t.Fatal(err)
	}

	body2, err := json.Marshal(exOut)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != string(body2) {
		t.Fatalf("%v\n\n%v",
			string(body), string(body2))
	}
}
