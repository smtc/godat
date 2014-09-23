package godat

import (
	"fmt"
	"testing"
)

func TestString(t *testing.T) {
	var (
		a []int
		m map[string]string
	)
	s := "权利越大，责任越大！饕餮盛宴。 more power, more duty!"

	for _, r := range s {
		fmt.Println(r, string(r))
	}
	for k, v := range m {
		fmt.Println(k, v)
	}
	a = append(a, 1)
	fmt.Println(m, a, m == nil)
}
