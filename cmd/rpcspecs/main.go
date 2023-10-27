//go:build ignore

// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	publicrpc "ctk/internal/pkg/public-rpc"
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

var (
	output = flag.String("o", "", "output file location. If not provided, the spec will only be printed in console.")
	help   = flag.Bool("h", false, "help text")
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("rpcspecs2: ")
	flag.Parse()

	if *help == true {
		flag.Usage()
		os.Exit(0)
	}

	log.Print("generating CTK RPC specs")
	methodMap := publicrpc.MethodMap
	out := Output{Methods: map[string]Method{}}

	for method, param := range methodMap {
		m := newMethod(param)
		out.Methods[method] = m
	}
	seqByte, _ := json.MarshalIndent(out, "", "    ")
	if *output == "" {
		log.Printf("json: %v", string(seqByte))
		os.Exit(0)
	} else {
		dir := filepath.Dir(".")
		outputName := filepath.Join(dir, strings.ToLower(*output))
		err := os.WriteFile(outputName, seqByte, 0644)
		if err != nil {
			log.Fatalf("writing output: %s", err)
		}
		log.Printf("written to: %v", outputName)
	}
}

func newMethod(param publicrpc.Method) Method {
	m := Method{}
	p := newProp(reflect.ValueOf(param))
	if p != nil {
		m.Parameters = *p
	}
	r := newProp(param.ReflectOutput())
	if r != nil {
		m.Returns = *r
	}
	return m
}

func newProp(rVal reflect.Value) *Property {
	dVal := deref(rVal)

	// construct prop
	switch dVal.Kind() {
	case reflect.String:
		return &Property{Type: "string"}
	case reflect.Bool:
		return &Property{Type: "bool"}
	case reflect.Float32, reflect.Float64:
		return &Property{Type: "number"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &Property{Type: "integer"}
	case reflect.Map:
		// NOTE: safe to assume Map is always map[string]string. We should use struct if it is not an abitrary map
		return &Property{Type: "object"}
	case reflect.Array, reflect.Slice:
		if items := extractArrayItems(dVal); items != nil {
			return &Property{
				Type:  "array",
				Items: items,
			}
		}
		return &Property{Type: "array"}
	case reflect.Struct:
		return &Property{
			Type:       "object",
			Properties: extractStructFields(dVal),
		}
	default:
		return nil
	}
}

func deref(rVal reflect.Value) reflect.Value {
	// dereference pointer
	var dVal reflect.Value
	if rVal.Kind() == reflect.Pointer {
		dVal = reflect.New(rVal.Type().Elem()).Elem()
	} else {
		dVal = rVal
	}
	return dVal
}

func extractStructFields(dVal reflect.Value) map[string]*Property {
	// assume dVal is already deref-ed and dVal.Kind is reflect.Struct
	dType := dVal.Type()
	m := map[string]*Property{}
	for i := 0; i < dVal.NumField(); i++ {
		field := dType.Field(i)
		if !field.IsExported() {
			continue
		}
		fName := getFieldNameFromTag(field.Tag)
		if fName == "" {
			fName = field.Name
		}
		fVal := deref(dVal.Field(i))
		m[fName] = newProp(fVal)
	}
	return m
}

func extractArrayItems(dVal reflect.Value) *Property {
	// assume dVal is already deref-ed and dVal.Kind is reflect.Array or reflect.Slice
	dType := dVal.Type()
	nVal := reflect.New(dType.Elem())
	return newProp(nVal)
}

func getFieldNameFromTag(tag reflect.StructTag) string {
	s := tag.Get("json")
	if s != "" {
		return strings.Split(s, ",")[0]
	}
	return ""
}

type Output struct {
	Methods map[string]Method `json:"methods"`
}

type Method struct {
	Parameters Property `json:"parameters"`
	Returns    Property `json:"returns"`
}

type Property struct {
	Type       string               `json:"type"`
	Properties map[string]*Property `json:"properties,omitempty"`
	Required   []string             `json:"required,omitempty"`
	Items      *Property            `json:"items,omitempty"`
}
