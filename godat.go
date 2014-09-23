package godat

import (
	"fmt"
	"log"
	"strings"
)

// double array trie algorithm implement in golang
// go实现的双数组

var (
	minReserve      int     = 5
	maxReserve      int     = 64
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
	// options
	nocase    bool // 区分大小写
	ascii     bool // 仅处理ascii
	minResv   int
	maxResv   int
	resvRatio float32
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

// 创建
func (gd *GoDat) Build() (err error) {
	return
}

// 全自动创建过程
func CreateGoDat(pats []string, opts map[string]interface{}) (gd *GoDat, err error) {
	gd.SetOptions(opts)
	if gd.ascii {
		for _, pat := range pats {
			if err = gd.addPattern(pat); err != nil {
				log.Printf("CreateGoDat: pattern %s is invalid, maybe it contains non-ascii characters, ignored.\n", pat)
			}
		}
	} else {
		gd.pats = pats
	}
	err = gd.Build()
	return
}

// 动态增加一个模式
func (gd *GoDat) AddPattern(pat string) {

}

func (gd *GoDat) Match(noodle string) bool {

}
