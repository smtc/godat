package godat

import (
	"bufio"
	//"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"testing"
)

func TestBuildDat(t *testing.T) {
	pats := []string{
		"abcd",
		"aa", "aaa",
		"ad",
		"aceiw",
		"djkafiew",
		"asd", "aglmnqioew",
		"bbbbb",
		"c",
		"bbbbbbae",
		"http://www.sina.cn", "alpha", "aaa", "zzbc", "fals", "hi!", "ab", "cc", "ca", "sets",
		"wow", "baa", "ma", "mm",
		"how", "bcefd", "apple", "google", "ms", "tencent", "baidu", "axon",
	}
	gd, err := CreateGoDat(pats, true)
	if err != nil {
		t.Fatal("create dat failed:", err)
	}
	gd.dump()

	for _, pat := range gd.pats {
		if gd.Match(pat) == false {
			t.Fatal("match should be true:" + pat)
		}
		//if gd.Match(pat+"!") == true {
		//	t.Fatal("match should be false")
		//}
	}
}

func testBuildDict(t *testing.T) {
	gd := GoDat{}
	f, err := os.Open("./dictionary.txt")
	if err != nil {
		t.Fatal("open dictionary file failed!\n")
	}
	rd := bufio.NewReader(f)
	for line, err := rd.ReadString(byte('\n')); err == nil || err != io.EOF; line, err = rd.ReadString(byte('\n')) {
		segs := strings.Split(line, " ")
		if len(segs) != 3 {
			continue
		}
		gd.add(segs[0])
	}
	gd.build()
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
