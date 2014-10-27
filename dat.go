package godat

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

func assert(cond bool, err string) {
	if cond == false {
		panic(err)
	}
}

func (gd *GoDat) build() (err error) {
	for _, s := range gd.pats {
		if err = gd.insertPattern(s); err != nil {
			return
		}
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
	if len(gd.base) <= 256 {
		for i := 0; i < len(gd.base); i++ {
			fmt.Printf("GoDat array index %d: %d    %d\n", i, gd.base[i], gd.check[i])
		}
	}
	fmt.Println("aux:", len(gd.auxiliary))
	if len(gd.auxiliary) <= 64 {
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
func (gd *GoDat) extend(l int) (err error) {
	orgLen := len(gd.base)
	length := orgLen
	if l == 0 {
		length = length * 2
	} else {
		for length <= l {
			length = length * 2
		}
	}

	if length > gd.maxLen {
		err = fmt.Errorf("Array cannot exceed dat's maxLen: %d", gd.maxLen)
		return
	}

	base := make([]int, length)
	check := make([]int, length)
	attrs := make([]uint, length)

	copy(base, gd.base)
	copy(check, gd.check)
	copy(attrs, gd.attrs)

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
	gd.attrs = attrs

	gd.idles += orgLen

	//fmt.Println("After extend:\nbase:", gd.base, "\ncheck:", gd.check)
	return nil
}

func (gd *GoDat) buildAux() {
	// 字符序号从1开始
	chs := 1
	if gd.auxiliary == nil {
		gd.auxiliary = make(map[rune]int)
	}
	if gd.revAuxiliary == nil {
		gd.revAuxiliary = make(map[int]rune)
	}
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
			//for t = -1 * gd.check[t]; t != gd.check[0] && t < s; {
			//	t = -1 * gd.check[t]
			//}
			for t = s; t < gd.maxLen; t++ {
				if gd.check[t] < 0 {
					break
				}
			}
			if t == gd.maxLen {
				t = gd.check[0]
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

	// 过滤空字符串, 去重
	m := make(map[string]bool)
	pats := make([]string, len(gd.pats))
	cnt := 0
	for i := 0; i < len(gd.pats); i++ {
		pat := gd.pats[i]
		if pat == "" {
			continue
		}
		if gd.nocase { // 转换大小写
			pat = strings.ToLower(pat)
		}
		if ok := m[pat]; !ok {
			m[pat] = true
			pats[cnt] = pat
			cnt++
		}
	}
	gd.pats = pats[0:cnt]

	// 对pattern排序, 节省空间
	sort.Sort(sort.StringSlice(gd.pats))
}

func (gd *GoDat) __find_pos(s, c int) int {
	if gd.idles > 0 {
		start := gd.check[0]
		pos := 0
		min := len(gd.auxiliary)
		_ = min

		for pos = start; pos <= min; {
			pos = -1 * gd.check[pos]
			if pos == start {
				break
			}
		}
		if pos > min {
			//fmt.Printf("__find_pos: s=%d c=%d pos=%d\n", s, c, pos)
			return pos
		}
	}
	//fmt.Printf("__find_pos: s=%d c=%d, extend array\n", s, c)
	if err := gd.extend(0); err != nil {
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

	if pos >= len(gd.base) {
		if err := gd.extend(pos); err != nil {
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

// find new base
// return -1 if cannot found a free pos
func (gd *GoDat) findNewBase(s int, ca []int) (int, int) {
	if gd.idles == 0 {
		if err := gd.extend(0); err != nil {
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
			if err := gd.extend(0); err != nil {
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
				if err := gd.extend(newbase + c); err != nil {
					panic(fmt.Sprintf("findNewBase failed: s=%d, ca=%v, pos=%d, array_len=%d, newbase=%d",
						s, ca, pos, len(gd.base), newbase))
					return -1, 0
				}
				arrLen = len(gd.base)
				// fmt.Printf(">>>>>>>>>>>>>>>>>>extended: newBase=%d, c=%d arrLen=%d\n", newbase, c, arrLen)
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

	panic(fmt.Sprintf("findNewBase failed: s=%d, ca=%v, pos=%d, array_len=%d", s, ca, pos, len(gd.base)))
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
			// fmt.Printf("i=%d, newbase=%d, c=%d, base[%d]=0\n", i, newbase, c, newbase+c)
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
	if oldbase < 0 {
		gd.base[s] = -newbase
	} else {
		gd.base[s] = newbase
	}
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
	// fmt.Printf("find new base: %d for s=%d, c=%d\n", pos, s, c)
	// 重新定位
	gd.reLocate(s, pos, newbase, ca)
	newPos = newbase + c
	return
}

// 增加一个模式
func (gd *GoDat) insertPattern(pat string) (err error) {
	var (
		r          rune
		c, i, s, t = 0, 0, 0, 0
		arrLen     = len(gd.base)
		//patLen     = len(pat)
		exist    bool
		conflict bool
	)

	for i, r = range pat {
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
	}
	// 最后一个字符
	if gd.base[t] > 0 {
		gd.base[t] = -1 * gd.base[t]
	} else if gd.base[t] == 0 {
		gd.base[t] = DAT_END_POS
	}
	return
}

// remove a pattern
func (gd *GoDat) removePattern(pat string) (err error) {
	var (
		r     rune
		i, tl int
		c, pc int
		s, t  int
		code  int
		base  int
		idx   int
	)

	if gd.Match(pat, 0) == false {
		return fmt.Errorf("pattern is not in dat, cannot remove.")
	}

	// remove pattern from dat
	if gd.nocase {
		pat = strings.ToLower(pat)
	}

	r, i = utf8.DecodeRuneInString(pat[0:])
	idx = gd.binSearch(pat[0:i])

	for tl = 0; tl < len(pat); {
		r, i = utf8.DecodeRuneInString(pat[tl:])

		code = gd.auxiliary[r]
		base = gd.base[s]
		if base >= 0 {
			t = base + code
		} else {
			t = -base + code
		}

		c = gd.commonCount(pat[0:tl+i], idx)
		fmt.Printf("  i=%d, tl=%d, r=%v, c=%d, s=%d, t=%d, code=%d, gd.check[s]=%d, gd.check[t]=%d\n", i, tl, string(r), c, s, t, code, gd.check[s], gd.check[t])
		if c <= 1 {
			break
		}

		s = t
		pc = c
		tl += i
	}

	// 该pat是其他模式的公共串
	if tl == len(pat) {
		assert(gd.base[t] < 0, "gd.base[t] should be less than 0:"+fmt.Sprintf("tl=%d, t=%d, base=%d", tl, t, gd.base[t]))
		gd.base[t] = -gd.base[t]
		return nil
	}

	// ab, abc, 删除abc时，将ab的b的base设置为DAT_END_POS
	if pc == 2 && base < 0 {
		gd.base[s] = DAT_END_POS
	}
	// 执行删除
	//tl += i
	for tl < len(pat) {
		r, i = utf8.DecodeRuneInString(pat[tl:])
		code = gd.auxiliary[r]
		//s = t
		base = gd.base[s]
		fmt.Printf("-- i=%d, tl=%d, r=%v, base[s]=%d, s=%d, t=%d, code=%d, gd.check[s]=%d, gd.check[t]=%d\n", i, tl, string(r), base, s, t, code, gd.check[s], gd.check[t])
		if base < 0 {
			// 应该是最后一个字符
			assert(base == DAT_END_POS, "this pos should be last character, base must be DAT_END_POS")
			assert(tl == len(pat), "this pos should be last character")
			t = -base + code
		} else {
			t = base + code
		}
		assert(gd.check[t] == s, fmt.Sprintf("remove pattern %s: pos %d is not next of pos %d, char at: %v", pat, t, s, r))

		assert(t > 0 && t < gd.maxLen, "pos t is invald: "+
			fmt.Sprintf("pat=%s, t=%d, s=%d, base=%d, tl=%d, code=%d, r=%v",
				pat, t, s, base, tl, code, string(r)))
		gd.addLink(s)
		s = t
		tl += i
	}

	// remove pattern from string array
	idx = gd.binSearch(pat)
	gd.pats = append(gd.pats[0:idx], gd.pats[idx+1:]...)
	fmt.Println(gd.pats)

	return nil
}

// return:
//   res = -10: pat有字符不在dat的字符表中，pat不在dat中
//   res = -20: pat不在dat中
//   res = -30: pat字符跳转表错误，pat不在dat中
func (gd *GoDat) removePat(pat string) (res int, err error) {
	var (
		i      int
		r      rune
		rl     int
		s, t   int
		code   int
		patLen int
		idx    int
		base   int
		nxtCnt int
	)

	if pat == "" {
		return
	}
	patLen = len(pat)
	if gd.nocase {
		pat = strings.ToLower(pat)
	}

	for i < patLen {
		r, rl = utf8.DecodeRuneInString(pat[i:])
		code = gd.auxiliary[r]
		if code == 0 {
			res = -10
			return
		}
		base = gd.base[s]
		if base >= 0 {
			t = base + code
		} else {
			if base == DAT_END_POS {
				return
			}
			t = -base + code
		}

		if t > len(gd.base) {
			res = -20
			return
		}
		if gd.check[t] != s {
			res = -30
			err = fmt.Errorf("pat %s NOT in dat", pat)
			return
		}
		i += rl
		s = t
	}
	i -= rl
	if base = gd.base[t]; base > 0 {
		// 模式pat是其他模式的子串，但模式pat本身不存在
		res = -40
		err = fmt.Errorf("pat %s NOT in dat", pat)
		return
	}

	assert(base != 0, "base should never be 0 except base[0]")
	s = gd.check[t]
	gd.attrs[t] = 0
	if base == DAT_END_POS {
		gd.addLink(t)
	} else {
		// 模式pat是其他模式的子串
		assert(base < 0, "base should < 0 here")
		gd.base[t] = -base

		//fmt.Printf("pat %s is part of other pats, just set base positive.\n", pat)
		goto delPat
	}

	// 计算第一个字符在gd.pats的偏移, 使commonPrefix更快
	_, rl = utf8.DecodeRuneInString(pat[0:])
	idx = gd.binSearch(pat[0:rl])

	// 从后向前删除
	r, rl = utf8.DecodeLastRuneInString(pat[0:i])
	i -= rl
	for i >= 0 {
		t = s
		nxtCnt = gd.commonCount(pat[0:i+rl], idx) - 1
		//assert(nxtCnt == gd.nextStateCount(t), fmt.Sprintf("commonCount=%d should equal with nextStateCount=%d, t=%d,i=%d, rl=%d",
		//	nxtCnt, gd.nextStateCount(t), t, i, rl))
		base = gd.base[t]
		if base < 0 {
			// 该节点是另一个模式的结束
			if nxtCnt == 0 {
				gd.base[t] = DAT_END_POS
			}

			goto delPat
		}
		if nxtCnt > 0 {
			goto delPat
		}
		s = gd.check[t]
		gd.addLink(t)

		r, rl = utf8.DecodeLastRuneInString(pat[0:i])
		if rl == 0 {
			goto delPat
		}
		i -= rl
	}

delPat:
	// remove pattern from string array
	idx = gd.binSearch(pat)
	gd.pats = append(gd.pats[0:idx], gd.pats[idx+1:]...)

	return
}

// 计算gd有多少个以subpat为开头的模式
func (gd *GoDat) nextStateCount(s int) int {
	var (
		code int
		cnt  int
		base int
	)
	base = gd.base[s]

	if base == DAT_END_POS {
		return 0
	}
	if base < 0 {
		base = -1 * base
	}
	for _, code = range gd.auxiliary {
		if (base + code) >= len(gd.base) {
			break
		}

		if gd.check[base+code] == s {
			cnt++
		}
	}

	return cnt
}
