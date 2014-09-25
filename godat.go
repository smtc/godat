package godat

import (
	"fmt"
	"log"
	"sort"
	"strings"
)

// double array trie algorithm implement in golang
// go实现的双数组

var (
	minReserve      int     = 5
	maxReserve      int     = 64
	initArrayLen    int     = 1024
	maxArrayLen     int     = 2 ^ 20
	defReserveRatio float32 = 1.1
)

type GoDat struct {
	name  string // optional, dat name
	base  []int  // base 表
	check []int  // check 表
	// 辅助词表, 根据字(rune)来查找该字在base数组中的位置(下标)
	// 如果ascii为true，则不需要辅助词表
	auxiliary map[rune]int
	// patterns
	pats []string
	// 下一个将要使用的位置
	nextPos int
	idles   int
	// options
	nocase    bool // 区分大小写
	ascii     bool // 仅处理ascii
	minResv   int
	maxResv   int
	maxLen    int
	resvRatio float32
}

// 创建双数组
func CreateGoDat(pats []string, opts map[string]interface{}) (gd *GoDat, err error) {
	gd.SetOptions(opts)
	if gd.ascii {
		for _, pat := range pats {
			if err = gd.add(pat); err != nil {
				log.Printf("CreateGoDat: pattern %s is invalid, maybe it contains non-ascii characters, ignored.\n", pat)
			}
		}
	} else {
		gd.pats = pats
	}
	err = gd.Build()
	return
}

// dump godat
func (gd *GoDat) dump() {
	if gd.name != "" {
		fmt.Println("dat " + gd.name + ":")
	}
	fmt.Println("options: nocase=%v ascii=%v\n", gd.nocase, gd.ascii)
	if !gd.ascii {
		fmt.Printf("dat auxiliary table: total = %d\n", len(gd.auxiliary))
	}
	fmt.Println("base  array:", gd.base)
	fmt.Println("check array:", gd.check)
}

// 增加一个模式
func (gd *GoDat) add(pat string) error {
	if gd.ascii {
		// 检查pat是否全为ascii字符
	}
	gd.pats = append(gd.pats, pat)
	return nil
}

func (gd *GoDat) SetOptions(opts map[string]interface{}) {
	var (
		i     int
		f     float32
		b, ok bool
	)

	gd.minResv = minReserve
	gd.maxResv = maxReserve
	gd.maxLen = maxArrayLen
	gd.resvRatio = defReserveRatio

	for key, v := range opts {
		k := strings.ToLower(key)
		switch k {
		case "nocase":
			if b, ok = v.(bool); ok {
				gd.nocase = b
			} else {
				log.Println("SetOptions: option nocase should be type bool")
			}
		case "ascii":
			if b, ok = v.(bool); ok {
				gd.ascii = b
			} else {
				log.Println("SetOptions: option ascii should be type bool")
			}
		case "minResv":
			if i, ok = v.(int); ok {
				gd.minResv = i
			} else {
				log.Println("SetOptions: option minResv should be type int")
			}
		case "maxResv":
			if i, ok = v.(int); ok {
				gd.minResv = i
			} else {
				log.Println("SetOptions: option maxResv should be type int")
			}
		case "resvRatio":
			if f, ok = v.(float32); ok {
				gd.resvRatio = f
			} else {
				log.Println("SetOptions: option resvRatio should be type float32")
			}
		default:
			log.Println("SetOptions: ignore unknown options " + key)
		}
	}
	// 设置默认选项nocase为true
	if opts["nocase"] == nil {
		gd.nocase = true
	}
}

// 扩大数组长度
func (gd *GoDat) extend() (err error) {
	length := len(gd.base) * 2
	if length > gd.maxLen {
		err = fmt.Errorf("Array cannot exceed dat's maxLen: %d", gd.maxLen)
		return
	}

	base := make([]int, length)
	check := make([]int, length)

	copy(base, gd.base)
	copy(check, gd.check)

	gd.base = base
	gd.check = check

	return nil
}

func (gd *GoDat) buildAux() {
	if gd.auxiliary == nil {
		gd.auxiliary = make(map[rune]int)
	}

	// 字符序号从1开始
	chs := 1

	for _, pat := range gd.pats {
		for _, ch := range pat {
			if _, ok := gd.auxiliary[ch]; !ok {
				gd.auxiliary[ch] = chs
				chs++
			}
		}
	}
}

func abs(i int) int {
	if i == 0 {
		return 0
	}
	if i < 0 {
		return -i
	}
	return i
}

// 创建
func (gd *GoDat) Build() (err error) {
	gd.base = make([]int, initArrayLen)
	gd.check = make([]int, initArrayLen)
	gd.auxiliary = make(map[rune]int)

	// 创建字符辅助表
	gd.buildAux()

	// 对pattern排序, 节省空间
	sort.Sort(sort.StringSlice(gd.pats))

	// 第一个非首字母将要插入的位置
	gd.nextPos = len(gd.pats)

	return
}

// 获取s的下一个字符r是否存在，存在返回下标，不存在返回-1
func (gd *GoDat) nextState(s, c int) int {
	// gd.check[0] should always be 0
	if gd.check[s] < 0 {
		return -1
	}

	// 检查状态转移表
	// c := gd.auxiliary[r]
	t := gd.base[s] + c
	if gd.check[t] != s {
		return -1
	}

	return t
}

// 增加一个模式
func (gd *GoDat) buildPattern(pat string) {
	arrLen := len(gd.base)
	patLen := len(pat)

	for i, r := range pat {

	}
}

// 无冲突版本
// 实现原理:
//   1、按列插入
//   2、加入链表中，如果一个字符串已经全部加入到dat中，把这个字符串删除
func (gd *GoDat) buildWithoutConflict() {

}

// 匹配
//
func (gd *GoDat) Match(noodle string, opt ...int) bool {
	return false
}