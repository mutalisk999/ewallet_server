package main

import (
	"testing"
	"fmt"
	"strings"
)

func TestClientLoop(t *testing.T) {
	data := "abcdefg:eabcdefg"
	fmt.Println(data)
	data_trim := strings.TrimLeft(data,"abcdefg")
	fmt.Println("data1:",data_trim)
	data = "abcdefg:eabcdefg"
	data_trim = strings.TrimPrefix(data,"abcdefg:")
	fmt.Println("data2:",data_trim)
}

