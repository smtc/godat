package godat

import (
	"fmt"
	"sort"
	"testing"
)

func TestBuildDat(t *testing.T) {
	gd, err := CreateGoDat([]string{"abcd", "c", "aa", "ad", "djkafiew", "aceiw", "bbbbb", "bbbbbbae", "asd", "aglmnqioew",
		"http://www.sina.cn", "alpha", "aaa", "zzbc", "fals", "hi!", "ab", "cc", "ca", "sets",
		"abcd", "wow", "baa", "ma", "mm",
		"how", "bcefd", "apple", "google", "ms", "tencent", "baidu", "axon"}, true)
	if err != nil {
		t.Fatal("create dat failed:", err)
	}
	gd.dump()
	/*
		runes := gd.prefixCount("", 0)
		for _, r := range runes {
			fmt.Println(string(r))
		}
	*/
	for _, pat := range gd.pats {
		if gd.Match(pat) == true {
			fmt.Printf("Found pattern %s\n", pat)
		}
		if gd.Match(pat+"!") == true {
			fmt.Printf("Found pattern %s\n", pat)
		}
	}
}

func testExtend(t *testing.T) {
	gd := GoDat{}
	gd.base = make([]int, initArrayLen)
	gd.check = make([]int, initArrayLen)
	gd.maxLen = 1024
	gd.initLink()
	gd.dump()
	if err := gd.extend(); err != nil {
		t.Fatal(err)
	}
	gd.dump()
}

func testBinSearch(t *testing.T) {
	gd := GoDat{
		pats: []string{"aaa", "zzbc", "fals", "hi!", "ab", "cc", "ca", "sets",
			"abcd", "wow", "baa", "ma", "mm",
			"how", "bcefd", "apple", "google", "ms", "tencent", "baidu", "axon"},
	}
	sort.Sort(sort.StringSlice(gd.pats))

	for i := 0; i < len(gd.pats); i++ {
		pat := gd.pats[i]
		s1 := gd.binSearch(pat)
		if i != s1 {
			t.Fatal("bin search failed at", i, s1, "\n")
		}
	}
}
