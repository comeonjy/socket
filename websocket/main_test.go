package main_test

import (
	"fmt"
	"testing"
)

func TestDemo(t *testing.T) {
	m := make(map[string]int)
	v, ok := m["demo"]
	fmt.Println(ok, v,m["demo"])

}
