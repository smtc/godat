package godat

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// double array trie algorithm implement in golang
// go实现的双数组

const DAT_END_POS = -2147483648

var (
	initArrayLen int = 32        // 数组初始长度
	maxArrayLen  int = (1 << 24) // 数组最大长度
)

type GoDat struct {
	name  string // optional, dat name
	base  []int  // base 表
	check []int  // check 表
	attrs []uint // attr 表
	// 辅助词表, 根据字(rune)来查找该字在base数组中的位置(下标)
	// 如果ascii为true，则不需要辅助词表
	auxiliary    map[rune]int
	revAuxiliary map[int]rune
	// patterns
	pats []string
	// 空闲的位置
	idles int
	// 数组最大长度
	maxLen int
	// options
	nocase bool // 区分大小写
}

// 创建双数组
func CreateGoDat(pats []string, nocase bool) (gd *GoDat, err error) {
	gd = &GoDat{pats: pats}

	gd.Initialize(nocase)
	// gd.dump()
	gd.build()

	return
}

// 初始化
func (gd *GoDat) Initialize(nocase bool) (err error) {
	gd.nocase = nocase
	gd.maxLen = maxArrayLen

	gd.base = make([]int, initArrayLen)
	gd.check = make([]int, initArrayLen)
	gd.attrs = make([]uint, initArrayLen)

	gd.auxiliary = make(map[rune]int)
	gd.revAuxiliary = make(map[int]rune)

	//将base和check组成双向链表
	gd.initLink()
	// 字符串排序
	gd.sort()
	// 创建字符辅助表
	gd.buildAux()

	return
}

// 无冲突版本
// 实现原理:
//
//   1、按列插入
//
func (gd *GoDat) BuildWithoutConflict() (err error) {
	ws := gd.toWords()
	err = gd.ncBuild(ws)
	return
}

// 匹配
// params:
//   opt: options
//      1: exact match
//      2: max match if possible, with gd.attrs
//      3: max match if any pattern found
//      4: min common pattern match
//
func (gd *GoDat) Match(noodle string, opt int) bool {
	var (
		i, rl int
		s, t  int
		base  int
		ch    rune
		code  int
		res   bool = true
	)

	if gd.nocase {
		noodle = strings.ToLower(noodle)
	}

	for i = 0; i < len(noodle); {
		ch, rl = utf8.DecodeRuneInString(noodle[i:])
		i += rl
		code = gd.auxiliary[ch]
		if code == 0 {
			// 不存在
			return false
		}
		base = gd.base[s]

		if base >= 0 {
			t = base + code
		} else if base == DAT_END_POS {
			break
		} else {
			t = -base + code
		}
		if gd.check[t] != s {
			return false
		}

		s = t
	}
	if i != len(noodle) {
		return false
	}

	if gd.base[t] >= 0 {
		if gd.base[t] == 0 {
			fmt.Printf("!!!!!base %d should NOT be 0.\n", t)
		}
		return false
	}
	return res
}
