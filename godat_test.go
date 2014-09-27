package godat

import (
	"fmt"
	"sort"
	"testing"
)

func TestString(t *testing.T) {
	var (
		a []int
		m map[string]string
	)
	s := "权利越大，责任越大！饕餮盛宴。 more power, more duty!"

	for _, r := range s {
		_ = r
		//fmt.Println(r, string(r))
	}
	fmt.Println(s[0])
	fmt.Println(s[1])
	fmt.Println(len(s))
	fmt.Println(len("权利越大"))
	for k, v := range m {
		fmt.Println(k, v)
	}
	a = append(a, 1)
	fmt.Println(m, a, m == nil)
}

func TestBinSearch(t *testing.T) {
	gd := GoDat{
		pats: []string{"aaa", "zzbc", "fals", "hi!", "ab", "cc", "ca", "sets",
			"abcd", "wow", "baa", "ma", "mm",
			"how", "bcefd", "apple", "google", "ms", "tencent", "baidu", "axon"},
	}
	sort.Sort(sort.StringSlice(gd.pats))
	fmt.Println(gd.pats)
	for i := 0; i < len(gd.pats); i++ {
		pat := gd.pats[i]
		s1 := gd.binSearch(pat)
		s2 := gd.binSearch(pat[0 : len(pat)-1])
		if i != s1 {
			t.Fatal("bin search failed at", i, s1, s2, "\n")
		}
		if i != s2 {
			fmt.Printf("Check it: i=%d s2=%d, %s %s\n", i, s2, pat, pat[0:len(pat)-1])
		}
	}
}
