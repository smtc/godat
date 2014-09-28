package godat

import (
	//"strings"
	"unicode/utf8"
)

// 回溯
// 获取数组s位置的所有下一状态的runes
func (gd *GoDat) backtrace(s int) []rune {
	res := ""
	prevPos := gd.check[s]
	for prevPos > 0 {
		prevBase := gd.base[prevPos]
		code := 0
		if prevBase > 0 {
			code = s - prevBase
		} else if prevBase != DAT_END_POS {
			code = s - (-1 * prevBase)
		}
		if code <= 0 {
			panic("code value invalid")
		}
		r := gd.revAuxiliary[code]
		res += string(r)
	}
	// 反转字符串
	runes := []rune(res)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return gd.prefixCount(string(runes))
}

// 计算gd的patterns中以p开头的字符串的数量
func (gd *GoDat) prefixCount(p string) (runes []rune) {
	patLen := len(gd.pats)
	// 二分查找最接近的位置
	pos := gd.binSearch(p) + 1
	plen := len(p)
	rm := make(map[rune]bool)
	for pos < patLen && len(gd.pats[pos]) > plen && gd.pats[pos][0:plen] == p {
		pat := gd.pats[pos]
		r, _ := utf8.DecodeRuneInString(pat[plen:])

		if _, ok := rm[r]; !ok {
			rm[rune(r)] = true
			runes = append(runes, rune(r))
		}
		pos++
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

	//return -1
	return right
}
