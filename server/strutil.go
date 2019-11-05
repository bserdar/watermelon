package server

import (
	"bufio"
	"fmt"
	"strings"
)

// ArrayContains returns if the array contains the string
func ArrayContains(arr []string, str string) bool {
	for _, x := range arr {
		if x == str {
			return true
		}
	}
	return false
}

// Words breaks the line into words separated by whitespace
func Words(in string) []string {
	scn := bufio.NewScanner(strings.NewReader(in))
	scn.Split(bufio.ScanWords)
	out := make([]string, 0)
	for scn.Scan() {
		out = append(out, scn.Text())
	}
	return out
}

// MapYaml converts map[interface{}]interface{} to map[string]interface{}
func MapYaml(in interface{}) interface{} {
	if arr, ok := in.([]interface{}); ok {
		out := make([]interface{}, 0, len(arr))
		for _, x := range arr {
			out = append(out, MapYaml(x))
		}
		return out
	}
	if m, ok := in.(map[interface{}]interface{}); ok {
		out := map[string]interface{}{}
		for k, v := range m {
			out[fmt.Sprint(k)] = MapYaml(v)
		}
		return out
	}
	return in
}
