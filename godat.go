package godat

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// double array trie algorithm implement in golang
// go实现的双数组

const DAT_END_POS = -2147483648

var (
	initArrayLen int = 64        // 数组初始长度
	maxArrayLen  int = (1 << 24) // 数组最大长度
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
	gd.nocase = nocase

	gd.initialize()
	gd.dump()
	gd.build()

	return
}

func (gd *GoDat) build() (err error) {
	//fmt.Println(gd.pats)
	for i, s := range gd.pats {
		if err = gd.buildPattern(s); err != nil {
			return
		}
		fmt.Println(i, "+++++++After insert", s)
		//fmt.Println("base array:", gd.base)
		//fmt.Println("check array:", gd.check)
	}
	return
}

// dump godat
func (gd *GoDat) dump() {
	if gd.name != "" {
		fmt.Println("dat " + gd.name + ":")
	}
	fmt.Printf("options: nocase=%v\n", gd.nocase)
	fmt.Printf("array length = %d, idles = %d\n", len(gd.base), gd.idles)
	if len(gd.base) <= 1024 {
		for i := 0; i < len(gd.base); i++ {
			fmt.Printf("GoDat array index %d: %d    %d\n", i, gd.base[i], gd.check[i])
		}
	}
	fmt.Println("aux:", len(gd.auxiliary))
	if len(gd.auxiliary) <= 256 {
		for k, v := range gd.auxiliary {
			fmt.Printf("    %s(%v): %d\n", string(k), k, v)
		}
		fmt.Println("reverse aux:", gd.revAuxiliary)
	}
	fmt.Println("patterns:", len(gd.pats))
}

// 增加一个模式
func (gd *GoDat) add(pat string) error {
	gd.pats = append(gd.pats, pat)
	return nil
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

	//fmt.Println("Before extend:\nbase:", gd.base, "\ncheck:", gd.check)

	gd.base = base
	gd.check = check
	gd.idles += orgLen

	//fmt.Println("After extend:\nbase:", gd.base, "\ncheck:", gd.check)
	return nil
}

func (gd *GoDat) buildAux() {
	// 字符序号从1开始
	chs := 1

	// 给每个字符标定序号
	for _, pat := range gd.pats {
		for _, ch := range pat {
			if gd.nocase {
				ch = unicode.ToLower(ch)
			}
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
	//fmt.Println("addLink:", s)
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

		//fmt.Println("delLink:", s, gd.base[s], gd.check[s])
		//fmt.Println("before delLink: base:", gd.base)
		//fmt.Println("before delLink: check:", gd.check)
		gd.base[-1*gd.check[s]] = gd.base[s]
		gd.check[-1*gd.base[s]] = gd.check[s]
		//fmt.Println("after delLink: base:", gd.base)
		//fmt.Println("after delLink: check:", gd.check)
	}
	gd.idles--
}

func (gd *GoDat) sort() {
	if len(gd.pats) == 0 {
		return
	}

	// 转换大小写, 过滤空字符串
	if gd.nocase {
		pats := make([]string, len(gd.pats))
		cnt := 0
		for i := 0; i < len(gd.pats); i++ {
			pat := gd.pats[i]
			if pat == "" {
				continue
			}
			pats[cnt] = strings.ToLower(pat)
			cnt++
		}
		gd.pats = pats[0:cnt]
	}

	// 对pattern排序, 节省空间
	sort.Sort(sort.StringSlice(gd.pats))

	if !gd.nocase {
		// 过滤空字符串
		pats := gd.pats
		for i, pat := range pats {
			if pat == "" {
				if i == len(pats)-1 {
					gd.pats = []string{}
					return
				}
				gd.pats = pats[i+1:]
			} else {
				break
			}
		}
	}
}

// 创建
func (gd *GoDat) initialize() (err error) {
	gd.maxLen = maxArrayLen

	gd.base = make([]int, initArrayLen)
	gd.check = make([]int, initArrayLen)
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

func (gd *GoDat) __find_pos(s, c int) int {
	if gd.idles > 0 {
		start := gd.check[0]
		pos := 0
		min := len(gd.auxiliary)
		_ = min

		for pos = start; pos <= c; {
			pos = -1 * gd.check[pos]
			if pos == start {
				break
			}
		}
		if pos > c {
			//fmt.Printf("__find_pos: s=%d c=%d pos=%d\n", s, c, pos)
			return pos
		}
	}
	//fmt.Printf("__find_pos: s=%d c=%d, extend array\n", s, c)
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
		pos = gd.base[0] + c
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
			if pos = gd.__find_pos(s, c); pos > 0 {
				gd.base[s] = c - pos
			}
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

	if gd.check[pos] >= 0 {
		// gd.check [pos]不可能=0
		conflict = true
	}

	return
}

//
func (gd *GoDat) nextPos(s, code int) (t int, isEnd bool) {

	if gd.base[s] == DAT_END_POS {
		// 结束状态
		return -1, true
	}
	if s == 0 {
		t = gd.base[0] + code
	} else {
		if gd.base[s] > 0 {
			t = gd.base[s] + code
		} else {
			t = -1*gd.base[s] + code
			isEnd = true
		}
	}

	if gd.check[t] != s {
		t = -1
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
	//chs := 1
	//if len(ca) > 0 {
	//	chs = ca[len(ca)-1]
	//}
	for ; ; pos = -1 * gd.check[pos] {
		if -1*gd.check[pos] == start {
			if err := gd.extend(); err != nil {
				return -1, 0
			}
		}
		//fmt.Printf("find new base loop: pos=%d, c=%d\n", pos, ca[0])
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
			if (newbase + c) >= arrLen {
				if err := gd.extend(); err != nil {
					fmt.Printf("findNewBase failed: s=%d, ca=%v, pos=%d, base array len=%d newbase=%d\nbase: %v\ncheck: %v\n",
						s, ca, pos, len(gd.base), newbase, gd.base, gd.check)
					panic("findNewBase failed inner")
					return -1, 0
				}
				arrLen = len(gd.base)
				fmt.Printf(">>>>>>>>>>>>>>>>>>extended: newBase=%d, c=%d arrLen=%d\n", newbase, c, arrLen)
			}
			if gd.check[newbase+c] >= 0 {
				found = false
				//fmt.Println("breaked at", newbase, c)
				break
			}
		}
		if found {
			return pos, newbase
		}
	}

	fmt.Printf("findNewBase failed: s=%d, ca=%v, pos=%d, base array len=%d\nbase: %v\ncheck: %v\n",
		s, ca, pos, len(gd.base), gd.base, gd.check)
	panic("findNewBase failed")
	return -1, 0
}

// re-locate
// s --> b
// newbase = pos - ca[0]
func (gd *GoDat) reLocate(s, pos, newbase int, ca []int) {
	oldbase := gd.base[s]
	//fmt.Printf("!!!!!!!!reLocate for s=%d, oldBase=%d, newBase=%d, pos=%d, ca=%v\n", s, oldbase, newbase, pos, ca)

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
			if oldbase < 0 {
				oldPos = -oldbase + c
			}

			gd.delLink(newPos)

			gd.base[newPos] = gd.base[oldPos]
			gd.check[newPos] = s

			// 下一跳的check值更新
			//fmt.Printf("  to be update pos %d, base=%d next children.\n", oldPos, gd.base[oldPos])
			if gd.base[oldPos] != DAT_END_POS {
				subca := gd.backtrace(oldPos, 0)
				//fmt.Printf("    sub char-array for old pos %d: %v\n", oldPos, subca)
				for _, child := range subca {
					//fmt.Printf("    set check [%d] to %d oldPos=%d child=%d\n",
					//	gd.base[oldPos]+child, newPos, oldPos, child)
					if gd.base[oldPos] > 0 {
						gd.check[gd.base[oldPos]+child] = newPos
					} else {
						gd.check[-1*gd.base[oldPos]+child] = newPos
					}
				}
			}

			// 旧状态节点变为空闲
			gd.addLink(oldPos)
			//fmt.Printf("i=%d c=%d add idle pos %d, del idle pos %d\n\n", i, c, oldPos, newPos)
		}
	}
	gd.base[s] = newbase
}

// 解除冲突 resolve conflicts
//
func (gd *GoDat) resolve(s, c int) (newPos int, err error) {
	// 根据字符得到编码
	ca := gd.backtrace(s, gd.revAuxiliary[c])
	//fmt.Printf("backtrace %d: %v\n", s, ca)

	// find new base with ca
	pos, newbase := gd.findNewBase(s, ca)
	if pos < 0 {
		return -1, fmt.Errorf("find new base for %d, %d failed.", s, c)
	}
	fmt.Printf("find new base: %d for s=%d, c=%d\n", pos, s, c)
	// 重新定位
	gd.reLocate(s, pos, newbase, ca)
	newPos = newbase + c
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
		//fmt.Printf("inserted: ch=%v, code=%d, s=%d[base=%d], t=%d, exist=%v, conflict=%v\n", string(r), c, s, gd.base[s], t, exist, conflict)

		if t < 0 {
			return fmt.Errorf("build pattern %s(%d) failed: origal array length=%d, now array length=%d",
				pat, i, arrLen, len(gd.base))
		}

		if exist == false {
			// exist==false的情况下, conflict有可能为true, 也有可能为false
			if conflict {
				// resolve已经把c加入到数组中, 且base设为0
				t, err = gd.resolve(s, c)
				if err != nil {
					return
				}
			} else {
				gd.delLink(t)
				// base值由下一个确定
				gd.base[t] = 0
				gd.check[t] = s
			}
		}
		s = t

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
//   0、排序
//   1、按列插入
//   2、加入链表中，如果一个字符串已经全部加入到dat中，把这个字符串删除
func (gd *GoDat) buildWithoutConflict() {

}

// 匹配
//
func (gd *GoDat) Match(noodle string) bool {
	res := true
	s := 0
	t := 0
	fmt.Println("Match: " + noodle)
	for _, ch := range noodle {
		code := gd.auxiliary[ch]
		t, _ = gd.nextPos(s, code)
		if t > 0 {
			fmt.Printf("    ch=%s, code=%d, t=%d, base[%d]=%d, check[%d]=%d\n", string(ch), code, t, t, gd.base[t], t, gd.check[t])
		} else {
			fmt.Printf("    ch=%s, code=%d, t=%d\n", string(ch), code, t)
		}
		if t < 0 {
			res = false
			break
		}
		s = t
	}
	return res
}
