package godat

import (
	"fmt"
	"unicode/utf8"
)

// 回溯
// 获取数组s位置的所有下一状态的runes
// endRune: 截至rune，0获取所有
func (gd *GoDat) backtrace(s int, endRune rune) []int {
	if s == 0 {
		return gd.firstChar(endRune)
	}

	if gd.base[s] == DAT_END_POS {
		return []int{}
	}

	tracePoint := s
	_ = tracePoint
	res := ""
	for {
		prevPos := gd.check[s]
		prevBase := gd.base[prevPos]
		code := 0
		if prevBase > 0 {
			code = s - prevBase
		} else if prevBase != DAT_END_POS {
			code = s - (-1 * prevBase)
		} else {
			// prev base 不应该为DAT_END_POS，如果prev base==DAT_END_POS, 就不会产生冲突，也就不需要backtrace
			panic("prevBase should NOT be DAT_END_POS")
		}
		if code <= 0 {
			fmt.Printf("s=%d, prevPos=%d, prevBase=%d, code=%d, res=%s\n", s, prevPos, prevBase, code, res)
			fmt.Println(gd.base, "\n", gd.check)
			panic("code value invalid")
		}
		r := gd.revAuxiliary[code]
		res += string(r)

		// gd.check[i] == 0，则到了字符串头部
		if prevPos <= 0 {
			break
		}
		s = prevPos
	}
	// 反转字符串
	runes := []rune(res)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	children := gd.prefixCount(string(runes), endRune)
	ca := make([]int, len(children))
	for i, r := range children {
		ca[i] = gd.auxiliary[r]
	}
	//fmt.Println("backtrase: s=", tracePoint, "string=", string(runes), children, ca)
	return ca
}

// 获取所有字符串的首字符
func (gd *GoDat) firstChar(endRune rune) []int {
	rm := make(map[rune]bool)
	runes := make([]rune, 0)
	for _, pat := range gd.pats {
		r, _ := utf8.DecodeRuneInString(pat[0:])
		if _, ok := rm[r]; !ok {
			rm[r] = true
			runes = append(runes, r)
		}
		if r == endRune {
			break
		}
	}
	ca := make([]int, len(runes))
	for i, r := range runes {
		ca[i] = gd.auxiliary[r]
	}
	return ca
}

// 计算gd的patterns中以p开头的字符串的数量
func (gd *GoDat) prefixCount(p string, endRune rune) (runes []rune) {
	patLen := len(gd.pats)
	plen := len(p)
	rm := make(map[rune]bool)

	// 二分查找最接近的位置
	for pos := gd.binSearch(p); pos < patLen; pos++ {
		pat := gd.pats[pos]
		if len(pat) <= plen {
			continue
		}
		if pat[0:plen] != p {
			break
		}

		r, _ := utf8.DecodeRuneInString(pat[plen:])

		if _, ok := rm[r]; !ok {
			rm[rune(r)] = true
			runes = append(runes, rune(r))
		}
		if r == endRune {
			break
		}
	}

	return
}

func (gd *GoDat) binSearch(p string) int {
	var (
		left   = 0
		right  = len(gd.pats)
		middle = 0
	)
	//如果这里是int right = n 的话，那么下面有两处地方需要修改，以保证一一对应：
	//1、下面循环的条件则是while(left < right)
	//2、循环内当array[middle]>value 的时候，right = mid
	//循环条件，适时而变
	for left < right {
		middle = left + ((right - left) >> 1) //防止溢出，移位也更高效。同时，每次循环都需要更新。

		if gd.pats[middle] > p {
			right = middle //right赋值，适时而变
		} else if gd.pats[middle] < p {
			left = middle + 1
		} else {
			return middle
		}
		//可能会有读者认为刚开始时就要判断相等，但毕竟数组中不相等的情况更多
		//如果每次循环都判断一下是否相等，将耗费时间
	}

	return right
}

// idx 从patterns[idx]开始搜索
// 获取以pattern[0:N]为公共子串的个数
func (gd *GoDat) commonCount(pattern string, idx int) int {
	var (
		l     int = len(pattern)
		start int
		cnt   int
		rm    = make(map[rune]bool)
	)
	for start = idx; start < len(gd.pats); start++ {
		pat := gd.pats[start]
		if len(pat) < l {
			break
		}
		if pat[0:l] != pattern {
			break
		}
		r, _ := utf8.DecodeRuneInString(pat[l:])
		if _, ok := rm[r]; !ok {
			cnt++
			rm[r] = true
		}
	}

	return cnt
}
