package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"go/parser"

	log "github.com/sirupsen/logrus"
)

type inputDescr struct {
	needName bool
	path     string
}

func generate(fname string) error {
	_, err := os.Stat(fname)
	if err != nil {
		return err
	}

	usedGlobalInputs := make(map[string]struct{})
	mGen := false
	iGen := false
	mmText := ""
	intText := ""
	intInitText := ""

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, fname, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	packageName := node.Name.Name

	inputs := make(map[string]inputDescr, 0)
	for _, imp := range node.Imports {
		name := ""
		if imp.Name != nil {
			name = imp.Name.Name
		} else {
			if strings.Contains(imp.Path.Value, "/") {
				name = imp.Path.Value[strings.LastIndex(imp.Path.Value, "/")+1 : len(imp.Path.Value)-1]
			} else {
				name = imp.Path.Value[1 : len(imp.Path.Value)-1]
			}
		}
		inputs[name] = inputDescr{
			needName: imp.Name != nil,
			path:     imp.Path.Value,
		}
	}

	for _, f := range node.Decls {
		genD, ok := f.(*ast.GenDecl)
		if !ok {
			log.Tracef("SKIP Decl %v is not *ast.GenDecl, `%v`\n", f, reflect.TypeOf(f))
			continue
		}
		var thisIsStruct bool
		var structName string
		var thisIsArray bool
		var thisIsMap bool
		var arrType *ast.ArrayType
		var mapType *ast.MapType
		for _, spec := range genD.Specs {
			currType, ok := spec.(*ast.TypeSpec)
			if !ok {
				log.Tracef("SKIP Spec %v is not ast.TypeSpec, `%v`\n", spec, reflect.TypeOf(spec))
				continue
			}

			switch currType.Type.(type) {
			case *ast.StructType:
				structName = currType.Name.Name
				thisIsStruct = true
			case *ast.ArrayType:
				structName = currType.Name.Name
				thisIsArray = true
				arrType = currType.Type.(*ast.ArrayType)
			case *ast.MapType:
				structName = currType.Name.Name
				thisIsMap = true
				mapType = currType.Type.(*ast.MapType)
			default:
				log.Tracef("SKIP currType %v is not ast.StructType, `%v`\n", currType.Type, reflect.TypeOf(currType.Type))
				continue
			}

		}
		needMarshalMethods := false
		needCodegenInterface := false
		var interfaceName string
		commList := make([]string, 0)
		if thisIsStruct || thisIsArray || thisIsMap {
			if genD.Doc == nil || genD.Doc.List == nil {
				continue
			}
			for _, comment := range genD.Doc.List {
				needMarshalMethods = needMarshalMethods || strings.HasPrefix(comment.Text, "//mfjson:marshal")
				if strings.HasPrefix(comment.Text, "//mfjson:interface") {
					needCodegenInterface = true
					interfaceName = strings.Trim(strings.Replace(comment.Text, "//mfjson:interface", "", 1), "\t ")
				}
				if strings.HasPrefix(comment.Text, "//mfjson:add") {
					comm := strings.Trim(strings.Replace(comment.Text, "//mfjson:add", "", 1), "\t ")
					if comm != "" {
						commList = append(commList, comm)
					}
				}
			}
		}
		if thisIsStruct {
			if needMarshalMethods {
				txt, usedInputs := generateMarshalMethods(genD, structName, commList)

				for k := range usedInputs {
					usedGlobalInputs[k] = struct{}{}
				}

				mmText += "\n" + txt
				mGen = true
			}

			if needCodegenInterface {
				if interfaceName == "" {
					interfaceName = structName
				}
				//generateMethods
				intText += "\n" + fmt.Sprintf(`func (obj *%v) UnmarshalJSONTypeName() string {
	return "%v"
}`, structName, interfaceName)

				intInitText += "\n\t" + fmt.Sprintf(
					`mfj.GlobalStructFactory.Add("%v", func() mfj.JsonInterfaceMarshaller { return &%v{} })
	mfj.GlobalStructFactory.AddNil("%v", func() mfj.JsonInterfaceMarshaller {
		var out *%v
		return out
	})`,
					interfaceName, structName,
					interfaceName, structName,
				)
				iGen = true
			}
		}

		if thisIsArray {
			if needMarshalMethods {
				txt, usedInputs := generateMarshalArrayMethods(arrType, structName, commList)

				for k := range usedInputs {
					usedGlobalInputs[k] = struct{}{}
				}

				mmText += "\n" + txt
				mGen = true
			}
		}
		if thisIsMap {
			if needMarshalMethods {
				txt, usedInputs := generateMarshalMapMethods(mapType, structName, commList)

				for k := range usedInputs {
					usedGlobalInputs[k] = struct{}{}
				}

				mmText += "\n" + txt
				mGen = true
			}
		}
	}

	if !mGen && !iGen {
		return nil
	}
	outFile := `// Code generated by mfjson for marshaling/unmarshaling. DO NOT EDIT.
// https://github.com/myfantasy/json

package ` + packageName + `
`
	if mGen {
		outFile += `
import (
	"encoding/json"

	"github.com/myfantasy/mft"

	mfj "github.com/myfantasy/json"

`
		for k := range usedGlobalInputs {
			imp, ok := inputs[k]
			if !ok {
				return fmt.Errorf("Unknown input name `%v`", k)
			}
			if imp.needName {
				if k != "mfj" {
					outFile += "\t" + k + " " + imp.path + "\n"
				}
			} else {
				if imp.path != `"github.com/myfantasy/mft"` && imp.path != `"encoding/json"` {
					outFile += "\t" + imp.path + "\n"
				}
			}
		}
		outFile += `)
`
	} else {
		outFile += `
import (
	mfj "github.com/myfantasy/json"
)`
	}

	outFile += mmText
	outFile += intText

	if iGen {
		outFile += fmt.Sprintf(`

func init() {%v
}
`, intInitText)

	}

	outFileName := strings.TrimSuffix(fname, ".go") + ".mfjson.go"
	err = ioutil.WriteFile(outFileName, []byte(outFile), 0666)
	if err != nil {
		return err
	}
	return nil
}

func getType(at interface{}) (fieldType string, isArray bool, arrLen string, isMap bool, mapKeyType string, usedInputs map[string]struct{}) {
	usedInputs = make(map[string]struct{})
	switch at.(type) {
	case *ast.Ident:
		fieldType = at.(*ast.Ident).Name
	case *ast.SelectorExpr:
		fieldType = at.(*ast.SelectorExpr).Sel.Name
		if expX, ok := at.(*ast.SelectorExpr).X.(*ast.Ident); ok {
			fieldType = expX.Name + "." + fieldType
			usedInputs[expX.Name] = struct{}{}
		}
	case *ast.StarExpr:
		subFieldType, subIsArray, subArrLen, subIsMap, subMapKeyType, subUsedInputs := getType(at.(*ast.StarExpr).X)
		for k := range subUsedInputs {
			usedInputs[k] = struct{}{}
		}
		switch {
		case !subIsArray && !subIsMap:
			fieldType = "*" + subFieldType
		case subIsArray:
			fieldType = "*[" + subArrLen + "]" + subFieldType
		case subIsMap:
			fieldType = "*map[" + subMapKeyType + "]" + subFieldType
		}
	case *ast.MapType:
		isMap = true
		{
			subFieldType, subIsArray, subArrLen, subIsMap, subMapKeyType, subUsedInputs := getType(at.(*ast.MapType).Key)
			for k := range subUsedInputs {
				usedInputs[k] = struct{}{}
			}
			switch {
			case !subIsArray && !subIsMap:
				mapKeyType = subFieldType
			case subIsArray:
				mapKeyType = "[" + subArrLen + "]" + subFieldType
			case subIsMap:
				mapKeyType = "map[" + subMapKeyType + "]" + subFieldType
			}
		}
		{
			subFieldType, subIsArray, subArrLen, subIsMap, subMapKeyType, subUsedInputs := getType(at.(*ast.MapType).Value)
			for k := range subUsedInputs {
				usedInputs[k] = struct{}{}
			}
			switch {
			case !subIsArray && !subIsMap:
				fieldType = subFieldType
			case subIsArray:
				fieldType = "[" + subArrLen + "]" + subFieldType
			case subIsMap:
				fieldType = "map[" + subMapKeyType + "]" + subFieldType
			}
		}
	case *ast.ArrayType:
		isArray = true
		if at.(*ast.ArrayType).Len != nil {
			arrLen = at.(*ast.ArrayType).Len.(*ast.BasicLit).Value
		}
		subFieldType, subIsArray, subArrLen, subIsMap, subMapKeyType, subUsedInputs := getType(at.(*ast.ArrayType).Elt)
		for k := range subUsedInputs {
			usedInputs[k] = struct{}{}
		}
		switch {
		case !subIsArray && !subIsMap:
			fieldType = subFieldType
		case subIsArray:
			fieldType = "[" + subArrLen + "]" + subFieldType
		case subIsMap:
			fieldType = "map[" + subMapKeyType + "]" + subFieldType
		}
	}
	return fieldType, isArray, arrLen, isMap, mapKeyType, usedInputs
}

func generateMarshalArrayMethods(arrType *ast.ArrayType, structName string, commList []string) (text string, usedInputs map[string]struct{}) {
	usedInputs = make(map[string]struct{})

	log.Tracef("Array `%v` gen", structName)

	newStructName := structName + "_mfjson_wrap"

	var fieldType string
	var fieldRawType string
	var arrLen string

	{
		var ui map[string]struct{}
		fieldType, _, arrLen, _, _, ui = getType(arrType)
		fieldRawType = fieldType
		for k := range ui {
			usedInputs[k] = struct{}{}
		}
		fieldRawType = "[" + arrLen + "]" + fieldRawType

	}

	var swlStr string
	if arrLen == "" {
		swlStr = fmt.Sprintf("swl := make([]mfj.IStructView, len(obj))")
	} else {
		swlStr = fmt.Sprintf("var swl [%v]mfj.IStructView", arrLen)
	}

	prefixStruct := ""
	for _, v := range commList {
		prefixStruct += v + "\n"
	}

	structText := prefixStruct + fmt.Sprintf("type %v []mfj.IStructView\n\n", newStructName)

	marshalText := fmt.Sprintf(`func (obj %v) MarshalJSON() (res []byte, err error) {
	if obj == nil {
		var out %v
		return json.Marshal(out)
	}
	out := make(%v, len(obj))
	`+swlStr+`
	for i := 0; i < len(obj); i++ {
		if ujo, ok := obj[i].(mfj.JsonInterfaceMarshaller); ok {
			sw := mfj.IStructView{}
			sw.Type = ujo.UnmarshalJSONTypeName()
			sw.Data, err = json.Marshal(obj[i])
			swl[i] = sw
		} else {
			swl[i] = mfj.IStructView{}
		}
	}

	return json.Marshal(out)
}
`, structName, newStructName, newStructName)

	unmarshalText := fmt.Sprintf(`func (obj *%v) UnmarshalJSON(data []byte) (err error) {
	if data == nil {
		return nil
	}
	var tmp %v
	err = json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	if tmp == nil {
		var d %v
		*obj = d
		return nil
	}
	objRaw := make(%v, len(tmp))
	*obj = objRaw
	for i := 0; i < len(tmp); i++ {
		if tmp[i].Type == "" {
			objRaw[i] = nil
		} else if tmp[i].Data == nil {
			to, er0 := mfj.GlobalStructFactory.GetNil(tmp[i].Type)
			if er0 != nil {
				return er0
			}
			toTrans, ok := to.(%v)
			if !ok {
				return mft.ErrorS("Type '%v' not valid in generations '%v' (ARR, NIL)")
			}
			objRaw[i] = toTrans
		} else {
			to, er0 := mfj.GlobalStructFactory.Get(tmp[i].Type)
			if er0 != nil {
				return er0
			}
			toTrans, ok := to.(%v)
			if !ok {
				return mft.ErrorS("Type '%v' not valid in generations '%v' (ARR)")
			}
			err = json.Unmarshal(tmp[i].Data, &toTrans)
			if err != nil {
				return err
			}
			objRaw[i] = toTrans
		}
	}
	return nil
}
`,
		structName, newStructName, structName, structName,
		fieldType,
		fieldType, structName,
		fieldType,
		fieldType, structName,
	)

	text = structText + marshalText + unmarshalText

	return text, usedInputs
}

func generateMarshalMapMethods(mapType *ast.MapType, structName string, commList []string) (text string, usedInputs map[string]struct{}) {
	usedInputs = make(map[string]struct{})

	log.Tracef("Map `%v` gen", structName)

	newStructName := structName + "_mfjson_wrap"

	var fieldType string
	var fieldRawType string
	var mapKeyType string

	{
		var ui map[string]struct{}
		fieldType, _, _, _, mapKeyType, ui = getType(mapType)
		fieldRawType = fieldType
		for k := range ui {
			usedInputs[k] = struct{}{}
		}
		fieldRawType = "map[" + mapKeyType + "]" + fieldRawType
	}

	var swlStr = fmt.Sprintf("swl := make(map[%v]mfj.IStructView, len(obj))", mapKeyType)

	prefixStruct := ""
	for _, v := range commList {
		prefixStruct += v + "\n"
	}

	structText := prefixStruct + fmt.Sprintf("type %v map[%v]mfj.IStructView\n\n", newStructName, mapKeyType)

	marshalText := fmt.Sprintf(`func (obj %v) MarshalJSON() (res []byte, err error) {
	if obj == nil {
		var out %v
		return json.Marshal(out)
	}
	`+swlStr+`
	for k, v := range obj {
		if ujo, ok := v.(mfj.JsonInterfaceMarshaller); ok {
			sw := mfj.IStructView{}
			sw.Type = ujo.UnmarshalJSONTypeName()
			sw.Data, err = json.Marshal(v)
			swl[k] = sw
		} else {
			swl[k] = mfj.IStructView{}
		}
	}

	return json.Marshal(swl)
}
`, structName, newStructName)

	unmarshalText := fmt.Sprintf(`func (obj *%v) UnmarshalJSON(data []byte) (err error) {
	if data == nil {
		return nil
	}
	var tmp %v
	err = json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	if tmp == nil {
		var d %v
		*obj = d
		return nil
	}
	objRaw := make(%v, len(tmp))
	*obj = objRaw
	for k, v := range tmp {
		if v.Type == "" {
			objRaw[k] = nil
		} else if v.Data == nil {
			to, er0 := mfj.GlobalStructFactory.GetNil(v.Type)
			if er0 != nil {
				return er0
			}
			toTrans, ok := to.(%v)
			if !ok {
				return mft.ErrorS("Type '%v' not valid in generations '%v' (ARR, NIL)")
			}
			objRaw[k] = toTrans
		} else {
			to, er0 := mfj.GlobalStructFactory.Get(v.Type)
			if er0 != nil {
				return er0
			}
			toTrans, ok := to.(%v)
			if !ok {
				return mft.ErrorS("Type '%v' not valid in generations '%v' (ARR)")
			}
			err = json.Unmarshal(v.Data, &toTrans)
			if err != nil {
				return err
			}
			objRaw[k] = toTrans
		}
	}
	return nil
}
`,
		structName, newStructName, structName, structName,
		fieldType,
		fieldType, structName,
		fieldType,
		fieldType, structName,
	)

	text = structText + marshalText + unmarshalText

	return text, usedInputs
}

func generateMarshalMethods(genD *ast.GenDecl, structName string, commList []string) (text string, usedInputs map[string]struct{}) {
	usedInputs = make(map[string]struct{})

	newStructName := structName + "_mfjson_wrap"

	prefixStruct := ""
	for _, v := range commList {
		prefixStruct += v + "\n"
	}

	structText := prefixStruct + fmt.Sprintf("type %v struct {\n", newStructName)

	marshalText := fmt.Sprintf("func (obj %v) MarshalJSON() (res []byte, err error) {\n\tout := %v{}\n", structName, newStructName)

	unmarshalText := fmt.Sprintf("func (obj *%v) UnmarshalJSON(data []byte) (err error) {\n\ttmp := %v{}\n\tif data == nil {\n\t\treturn nil\n\t}\n", structName, newStructName)
	unmarshalText +=
		`	err = json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
`

	for _, spec := range genD.Specs {
		currType, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		currStruct, ok := currType.Type.(*ast.StructType)
		if !ok {
			continue
		}

		for idxField, field := range currStruct.Fields.List {
			if len(field.Names) == 0 {
				continue
			}
			if strings.ToLower(field.Names[0].Name[0:1]) == field.Names[0].Name[0:1] {
				continue
			}
			var fieldType string
			var fieldRawType string
			var arrLen string
			var mapKeyType string
			isArray := false
			isMap := false

			{
				var ui map[string]struct{}
				fieldType, isArray, arrLen, isMap, mapKeyType, ui = getType(field.Type)
				fieldRawType = fieldType
				for k := range ui {
					usedInputs[k] = struct{}{}
				}
				if isArray {
					fieldRawType = "[" + arrLen + "]" + fieldRawType
				}
				if isMap {
					fieldRawType = "map[" + mapKeyType + "]" + fieldRawType
				}
			}

			if idxField != 0 {
				structText += "\n"
			}

			if field.Tag != nil {
				tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
				tagVal := tag.Get("mfjson")
				if strings.HasPrefix(tagVal, "true") {
					//*****************
					structText += fmt.Sprintf("\t// %v %v %v\n", field.Names[0].Name, fieldRawType, field.Tag.Value)
					if isArray {
						structText += fmt.Sprintf("\t%v [%v]mfj.IStructView %v\n", field.Names[0].Name, arrLen, field.Tag.Value)
						swlStr := ""
						ulStr := ""
						if arrLen == "" {
							ulStr = fmt.Sprintf("obj.%v = make(%v, len(tmp.%v))", field.Names[0].Name, fieldRawType, field.Names[0].Name)
							swlStr = fmt.Sprintf("swl := make([]mfj.IStructView, len(obj.%v))", field.Names[0].Name)
						} else {
							swlStr = fmt.Sprintf("var swl [%v]mfj.IStructView", arrLen)
						}
						marshalText += fmt.Sprintf(`	{
		if obj.%v == nil {
			out.%v = nil
		} else {
			`+swlStr+`
			for i := 0; i < len(obj.%v); i++ {
				if ujo, ok := obj.%v[i].(mfj.JsonInterfaceMarshaller); ok {
					sw := mfj.IStructView{}
					sw.Type = ujo.UnmarshalJSONTypeName()
					sw.Data, err = json.Marshal(obj.%v[i])
					swl[i] = sw
				} else {
					swl[i] = mfj.IStructView{}
				}
			}
			out.%v = swl
		}
	}
`,
							field.Names[0].Name, field.Names[0].Name, field.Names[0].Name,
							field.Names[0].Name, field.Names[0].Name,
							field.Names[0].Name)

						unmarshalText += fmt.Sprintf(`	{
		if tmp.%v == nil {
			obj.%v = nil
		} else {
			`+ulStr+`
			for i := 0; i < len(tmp.%v); i++ {
				if tmp.%v[i].Type == "" {
					obj.%v[i] = nil
				} else if tmp.%v[i].Data == nil {
					to, er0 := mfj.GlobalStructFactory.GetNil(tmp.%v[i].Type)
					if er0 != nil {
						return er0
					}
					toTrans, ok := to.(%v)
					if !ok {
						return mft.ErrorS("Type '%v' not valid in generations '%v..%v' (NIL)")
					}
					obj.%v[i] = toTrans
				} else {
					to, er0 := mfj.GlobalStructFactory.Get(tmp.%v[i].Type)
					if er0 != nil {
						return er0
					}
					toTrans, ok := to.(%v)
					if !ok {
						return mft.ErrorS("Type '%v' not valid in generations '%v..%v'")
					}
					err = json.Unmarshal(tmp.%v[i].Data, &toTrans)
					if err != nil {
						return err
					}
					obj.%v[i] = toTrans
				}
			}
		}
	}
`,
							field.Names[0].Name, field.Names[0].Name,
							field.Names[0].Name,
							field.Names[0].Name, field.Names[0].Name,
							field.Names[0].Name, field.Names[0].Name, fieldType,
							field.Names[0].Name,
							fieldType, structName, field.Names[0].Name,
							field.Names[0].Name, fieldType,
							fieldType, structName, field.Names[0].Name,
							field.Names[0].Name, field.Names[0].Name)
					} else if isMap {
						structText += fmt.Sprintf("\t%v map[%v]mfj.IStructView %v\n", field.Names[0].Name, mapKeyType, field.Tag.Value)

						swlStr := fmt.Sprintf("swl := make(map[%v]mfj.IStructView, len(obj.%v))", mapKeyType, field.Names[0].Name)
						ulStr := fmt.Sprintf("obj.%v = make(%v, len(tmp.%v))", field.Names[0].Name, fieldRawType, field.Names[0].Name)
						marshalText += fmt.Sprintf(`	{
		if obj.%v == nil {
			out.%v = nil
		} else {
			`+swlStr+`
			for k, v := range obj.%v {
				if ujo, ok := v.(mfj.JsonInterfaceMarshaller); ok {
					sw := mfj.IStructView{}
					sw.Type = ujo.UnmarshalJSONTypeName()
					sw.Data, err = json.Marshal(v)
					swl[k] = sw
				} else {
					swl[k] = mfj.IStructView{}
				}
			}
			out.%v = swl
		}
	}
`,
							field.Names[0].Name, field.Names[0].Name, field.Names[0].Name,
							field.Names[0].Name)

						unmarshalText += fmt.Sprintf(`	{
		if tmp.%v == nil {
			obj.%v = nil
		} else {
			`+ulStr+`
			for k, v := range tmp.%v {
				if v.Type == "" {
					obj.%v[k] = nil
				} else if v.Data == nil {
					to, er0 := mfj.GlobalStructFactory.GetNil(v.Type)
					if er0 != nil {
						return er0
					}
					toTrans, ok := to.(%v)
					if !ok {
						return mft.ErrorS("Type '%v' not valid in generations '%v[]%v' (NIL)")
					}
					obj.%v[k] = toTrans
				} else {
					to, er0 := mfj.GlobalStructFactory.Get(v.Type)
					if er0 != nil {
						return er0
					}
					toTrans, ok := to.(%v)
					if !ok {
						return mft.ErrorS("Type '%v' not valid in generations '%v[]%v'")
					}
					err = json.Unmarshal(v.Data, &toTrans)
					if err != nil {
						return err
					}
					obj.%v[k] = toTrans
				}
			}
		}
	}
`,
							field.Names[0].Name, field.Names[0].Name,
							field.Names[0].Name,
							field.Names[0].Name,
							fieldType,
							field.Names[0].Name,
							fieldType, structName, field.Names[0].Name,
							fieldType,
							fieldType, structName, field.Names[0].Name,
							field.Names[0].Name)
					} else {
						structText += fmt.Sprintf("\t%v mfj.IStructView %v\n", field.Names[0].Name, field.Tag.Value)
						marshalText += fmt.Sprintf(`	{
		if ujo, ok := obj.%v.(mfj.JsonInterfaceMarshaller); ok {
			sw := mfj.IStructView{}
			sw.Type = ujo.UnmarshalJSONTypeName()
			sw.Data, err = json.Marshal(obj.%v)
			out.%v = sw
		} else {
			out.%v = mfj.IStructView{}
		}
	}
`, field.Names[0].Name, field.Names[0].Name, field.Names[0].Name, field.Names[0].Name)
						unmarshalText += fmt.Sprintf(`	{
		if tmp.%v.Type == "" {
			obj.%v = nil
		} else if tmp.%v.Data == nil {
			to, er0 := mfj.GlobalStructFactory.GetNil(tmp.%v.Type)
			if er0 != nil {
				return er0
			}
			toTrans, ok := to.(%v)
			if !ok {
				return mft.ErrorS("Type '%v' not valid in generations '%v.%v' (NIL)")
			}
			obj.%v = toTrans
		} else {
			to, er0 := mfj.GlobalStructFactory.Get(tmp.%v.Type)
			if er0 != nil {
				return er0
			}
			toTrans, ok := to.(%v)
			if !ok {
				return mft.ErrorS("Type '%v' not valid in generations '%v.%v'")
			}
			err = json.Unmarshal(tmp.%v.Data, &toTrans)
			if err != nil {
				return err
			}
			obj.%v = toTrans
		}
	}
`,
							field.Names[0].Name, field.Names[0].Name,
							field.Names[0].Name,
							field.Names[0].Name, fieldType,
							fieldType, structName, field.Names[0].Name,
							field.Names[0].Name,
							field.Names[0].Name, fieldType,
							fieldType, structName, field.Names[0].Name,
							field.Names[0].Name,
							field.Names[0].Name)
					}
					//*****************
				} else {
					structText += fmt.Sprintf("\t%v %v %v\n", field.Names[0].Name, fieldRawType, field.Tag.Value)
					marshalText += fmt.Sprintf("\tout.%v = obj.%v\n", field.Names[0].Name, field.Names[0].Name)
					unmarshalText += fmt.Sprintf("\tobj.%v = tmp.%v\n", field.Names[0].Name, field.Names[0].Name)
				}
			} else {
				structText += fmt.Sprintf("\t%v %v\n", field.Names[0].Name, fieldRawType)
				marshalText += fmt.Sprintf("\tout.%v = obj.%v\n", field.Names[0].Name, field.Names[0].Name)
				unmarshalText += fmt.Sprintf("\tobj.%v = tmp.%v\n", field.Names[0].Name, field.Names[0].Name)
			}
		}
	}
	structText += "}\n\n"
	marshalText +=
		`	return json.Marshal(out)
}
`
	unmarshalText +=
		`	return nil
}
`
	text = structText + marshalText + unmarshalText
	return text, usedInputs
}

func main() {
	flag.Parse()

	files := flag.Args()

	for _, fname := range files {
		if err := generate(fname); err != nil {
			log.Fatal(err)
		}
	}
}
