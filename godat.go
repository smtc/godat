package godat

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"unicode"
)

// double array trie algorithm implement in golang
// go实现的双数组

const DAT_END_POS = -2147483648

var (
	minReserve      int     = 5
	maxReserve      int     = 64
	initArrayLen    int     = 32
	maxArrayLen     int     = (1 << 20)
	defReserveRatio float32 = 1.1
)

type GoDat struct {
	name  string // optional, dat name
	base  []int  // base 表
	check []int  // check 表
	// 辅助词表, 根据字(rune)来查找该字在base数组中的位置(下标)
	// 如果ascii为true，则不需要辅助词表
	auxiliary    map[rune]int
	revAuxiliary map[int]rune
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
	gd = &GoDat{}
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
	fmt.Printf("options: nocase=%v ascii=%v\n", gd.nocase, gd.ascii)
	if !gd.ascii {
		fmt.Printf("dat auxiliary table: total = %d\n", len(gd.auxiliary))
	}
	fmt.Println("base  array:", gd.base)
	fmt.Println("check array:", gd.check)
	fmt.Println("aux:", gd.auxiliary)
	fmt.Println("reverse aux:", gd.revAuxiliary)
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
	orgLen := len(gd.base)
	length := orgLen * 2
	fmt.Printf("extend array from %d to %d\n", orgLen, length)
	if length > gd.maxLen {
		err = fmt.Errorf("Array cannot exceed dat's maxLen: %d", gd.maxLen)
		return
	}

	base := make([]int, length)
	check := make([]int, length)

	copy(base, gd.base)
	copy(check, gd.check)

	// 将base和check数组的新元素连接起来
	for i := orgLen; i < length; i++ {
		base[i] = -(i - 1)
		check[i] = -(i + 1)
	}
	if check[0] == 0 {
		// 原来已经没有空位置
		check[0] = orgLen
		base[orgLen] = -(length - 1)
		check[length-1] = -orgLen
	} else {
		// 将扩展出的位置与原来的空闲位置连接起来
		first := check[0]
		last := -1 * base[first]

		base[orgLen] = -last
		check[last] = -orgLen

		base[first] = -length + 1
		check[length-1] = -first
	}

	gd.base = base
	gd.check = check
	gd.idles += orgLen

	fmt.Println("After extend:\nbase:", gd.base, "\ncheck:", gd.check)
	return nil
}

func (gd *GoDat) buildAux() {
	// 字符序号从1开始
	chs := 1

	// 给每个字符标定序号
	for _, pat := range gd.pats {
		for _, ch := range pat {
			if _, ok := gd.auxiliary[ch]; !ok {
				gd.auxiliary[ch] = chs
				gd.revAuxiliary[chs] = ch
				chs++
			}
		}
	}
}

// 将base和check数组组成双向链表，借鉴小稳的C实现原理
func (gd *GoDat) initLink() {
	// 数组位置0预留
	gd.base[0] = 0
	gd.check[0] = 1

	// 数组从1到length-1的各个位置连接起来
	length := len(gd.base)
	for i := 1; i < length; i++ {
		gd.base[i] = -(i - 1)
		gd.check[i] = -(i + 1)
	}
	gd.base[1] = -(length - 1)
	gd.check[length-1] = -1

	gd.idles = length - 1
}

// 节点用完后, 调用addLink使该节点变成idle
func (gd *GoDat) addLink(s int) {
	if gd.idles == 0 {
		gd.check[0] = s
		gd.base[s] = -s
		gd.check[s] = -s
	} else {
		t := gd.check[0]
		// 找到s的下一个空节点位置
		if s > t {
			for t = -1 * gd.check[t]; t != gd.check[0] && t < s; {
				if t < 0 || t > len(gd.check) {
					fmt.Println("base:", gd.base)
					fmt.Println("check:", gd.check)
					fmt.Println("addLink: s=", s, "t=", t)
					panic("invalid t")
				}
				t = -1 * gd.check[t]
			}
		}

		gd.base[s] = gd.base[t]
		gd.check[-1*gd.base[t]] = -1 * s

		gd.base[t] = -1 * s
		gd.check[s] = -1 * t

		if s < gd.check[0] {
			gd.check[0] = s
		}
	}

	gd.idles++
}

// 使用节点前调用delLink删除节点
func (gd *GoDat) delLink(s int) {
	if gd.idles <= 0 {
		return
	}

	if gd.idles == 1 {
		gd.check[0] = 0
	} else {
		// 是首节点的话, 修改check[0]指向下一个空闲的节点
		if s == gd.check[0] {
			gd.check[0] = -1 * gd.check[s]
		}

		fmt.Println("delLink:", s, gd.check[s])
		fmt.Println("check:", gd.check)
		gd.base[-1*gd.check[s]] = gd.base[s]
		gd.check[-1*gd.base[s]] = gd.check[s]
	}
	gd.idles--
}

// 创建
func (gd *GoDat) Build() (err error) {
	gd.base = make([]int, initArrayLen)
	gd.check = make([]int, initArrayLen)
	gd.auxiliary = make(map[rune]int)
	gd.revAuxiliary = make(map[int]rune)

	//将base和check组成双向链表
	gd.initLink()
	// 创建字符辅助表
	gd.buildAux()

	// 对pattern排序, 节省空间
	sort.Sort(sort.StringSlice(gd.pats))
	gd.dump()

	for _, s := range gd.pats {
		if err = gd.buildPattern(s); err != nil {
			return
		}

		fmt.Println("After insert", s, ":\nbase array:", gd.base)
		fmt.Println("check array:", gd.check)
	}

	return
}

func (gd *GoDat) __find_pos(s, c int) int {
	if gd.idles > 0 {
		start := gd.check[0]
		pos := 0
		min := len(gd.auxiliary)

		for pos = start; pos <= min; {
			pos = -1 * gd.check[pos]
			if pos == start {
				break
			}
		}
		if pos > min {
			return pos
		}
	}
	if err := gd.extend(); err != nil {
		return -1
	}
	return len(gd.base) / 2
}

// 为s的下一个状态c找到一个位置
// 返回：
//		pos: c的位置, -1 为失败
//      exist: pos位置是否有字符c
//      conflict: 是否冲突
func (gd *GoDat) findPos(s, c int) (pos int, exist, conflict bool) {
	if s == 0 {
		pos = c
	} else {
		if gd.base[s] == 0 {
			// base[s] == 0 该位置是一个中间状态，其base值由找到的位置和c来决定
			// 为c找到位置
			pos = gd.__find_pos(s, c)
			if pos > 0 {
				gd.base[s] = pos - c
			}
			return
		} else if gd.base[s] > 0 {
			// 只有base[0] = 0
			pos = gd.base[s] + c
		} else if gd.base[s] == DAT_END_POS {
			// base[s] == DAT_END_POS 该位置是一个结束位置
			// 为c找到位置
			pos = gd.__find_pos(s, c)
			return
		} else {
			// base[s] < 0
			// base[s]是一个结束字符，且该字符后有其他字符
			pos = -1*gd.base[s] + c
		}
	}

	for pos >= len(gd.base) {
		if err := gd.extend(); err != nil {
			pos = -1
			fmt.Println("extend failed:", err)
			return
		}
	}

	if gd.check[pos] == s {
		exist = true
		return
	}

	if gd.check[pos] < 0 {
		//pos = gd.__find_pos(s, c)
	} else {
		// gd.check [pos]不可能=0
		conflict = true
	}

	return
}

// find new base
// return -1 if cannot found a free pos
func (gd *GoDat) findNewBase(s int, ca []int) (int, int) {
	if gd.idles == 0 {
		if err := gd.extend(); err != nil {
			return -1, 0
		}
	}

	start := gd.check[0]
	pos := start
	arrLen := len(gd.base)
	chs := len(gd.auxiliary)
	for {
		for ; -1*gd.check[pos] != start; pos = -1 * gd.check[pos] {
			// new base 的最小位置
			if pos <= chs {
				continue
			}
			found := true
			newbase := pos - ca[0]

			// 0位置跳过
			i := 1
			for ; i < len(ca); i++ {
				c := ca[i]
				if !(newbase+c < arrLen && gd.check[newbase+c] < 0) {
					found = false
					break
				}
			}
			if found {
				return pos, newbase
			}
			if newbase+ca[i] >= arrLen {
				break
			}
		}
		if err := gd.extend(); err != nil {
			return -1, 0
		}
		pos = arrLen
	}

	return -1, 0
}

// re-locate
// s --> b
// newbase = pos - ca[0]
func (gd *GoDat) reLocate(s, pos, newbase int, ca []int) {
	oldbase := gd.base[s]
	fmt.Println("reLocate:", ca)

	for i, c := range ca {
		// 新加入的状态
		if i == len(ca)-1 {
			fmt.Printf("i=%d, newbase=%d, c=%d\n", i, newbase, c)
			gd.delLink(newbase + c)
			gd.check[newbase+c] = s
			gd.base[newbase+c] = 0
		} else {
			// s的下一跳字符状态由oldPos迁移到newPos
			newPos := newbase + c
			oldPos := oldbase + c

			gd.delLink(newPos)

			gd.base[newPos] = gd.base[oldPos]
			gd.check[newPos] = s

			// 下一跳的check值更新
			if gd.base[oldPos] != DAT_END_POS {
				ca := gd.backtrace(oldPos, gd.revAuxiliary[c])
				for _, child := range ca {
					fmt.Printf("    set check [%d] to %d oldPos=%d child=%d\n",
						gd.base[oldPos]+child, newPos, oldPos, child)
					gd.check[gd.base[oldPos]+child] = newPos
				}
			}

			// 旧状态节点变为空闲
			gd.addLink(oldPos)
			fmt.Printf("i=%d c=%d add idle pos %d, del idle pos %d\n\n", i, c, oldPos, newPos)
		}
	}
	gd.base[s] = newbase
}

// 解除冲突 resolve conflicts
//
func (gd *GoDat) resolve(s, c int) (err error) {
	// 根据字符得到编码
	ca := gd.backtrace(s, gd.revAuxiliary[c])
	fmt.Printf("backtrace %d: %v\n", s, ca)

	// find new base with ca
	pos, newbase := gd.findNewBase(s, ca)
	if pos < 0 {
		return fmt.Errorf("find new base for %d, %d failed.", s, c)
	}
	fmt.Printf("find new base: %d for s=%d, c=%d\n", pos, s, c)
	// 重新定位
	gd.reLocate(s, pos, newbase, ca)
	return
}

// 增加一个模式
func (gd *GoDat) buildPattern(pat string) (err error) {
	var (
		r          rune
		c, i, s, t = 0, 0, 0, 0
		arrLen     = len(gd.base)
		patLen     = len(pat)
		exist      bool
		conflict   bool
	)

	for i, r = range pat {
		if gd.nocase {
			r = unicode.ToLower(r)
		}

		c = gd.auxiliary[r]
		t, exist, conflict = gd.findPos(s, c)
		fmt.Printf("insert character %v, index %d, s=%d, t=%d, exist=%v, conflict=%v\n", string(r), c, s, t, exist, conflict)

		if t < 0 {
			return fmt.Errorf("build pattern %s(%d) failed: origal array length=%d, now array length=%d",
				pat, i, arrLen, len(gd.base))
		}

		if exist {
			s = t
		} else {
			// exist==false的情况下, conflict有可能为true, 也有可能为false
			if conflict {
				err = gd.resolve(s, c)
				if err != nil {
					return
				}
			} else {
				gd.delLink(t)
				// base值由下一个确定
				gd.base[t] = 0
				gd.check[t] = s
				s = t
			}
		}

		if i == patLen-1 {
			// r为最后一个字符
			if gd.base[t] > 0 {
				gd.base[t] = -1 * gd.base[t]
			} else if gd.base[t] == 0 {
				gd.base[t] = DAT_END_POS
			}
		}
	}

	return
}

// 无冲突版本
// 实现原理:
//   1、按列插入
//   2、加入链表中，如果一个字符串已经全部加入到dat中，把这个字符串删除
func (gd *GoDat) buildWithoutConflict() {

}

// 匹配
//
func (gd *GoDat) Match(noodle string) bool {
	return false
}
