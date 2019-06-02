package main

import (
	"os"
	"reflect"
	"testing"

	"github.com/mresvanis/argo/pkg/registry"
)

var (
	testcfg *Config
	testreg registry.Registrar
)

func TestMain(m *testing.M) {
	f, err := os.Open("config.test.json")
	if err != nil {
		panic(err)
	}

	testcfg, err = ParseConfig(f)
	if err != nil {
		panic(err)
	}

	testreg = registry.NewRegistry("./testdata/argo_test.db")
	if err := testreg.Open(); err != nil {
		panic(err)
	}
	defer testreg.Close()

	result := m.Run()
	os.Exit(result)
}

func assertEq(a, b interface{}, t *testing.T) {
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("Expected %#v and %#v to be equal", a, b)
	}
}
