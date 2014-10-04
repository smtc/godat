package godat

import (
	"fmt"
	"sort"
	"unicode/utf8"
)

// 无冲突build双数组
type Words []rune

//
// 去重
func (gd *GoDat) toWords() []Words {
	var (
		cnt int
		m   = make(map[string]bool)
	)
	ws := make([]Words, len(gd.pats))
	for _, pat := range gd.pats {
		if _, ok := m[pat]; ok {
			continue
		}
		m[pat] = true
		ru := make(Words, 0)
		for len(pat) > 0 {
			r, size := utf8.DecodeRuneInString(pat)
			pat = pat[size:]
			ru = append(ru, r)
		}
		ws[cnt] = ru
		cnt++
	}

	return ws[0:cnt]
}

func (gd *GoDat) ncBuild(ws []Words) error {
	type buildStack struct {
		endRow int
		column int
		s      int
	}
	var (
		s      = 0
		t      = 0
		cursor = 0
		codes  []int
		stacks = make([]buildStack, 0)
		stack  buildStack
		err    error
	)

	fmt.Println(len(gd.pats))
	//fmt.Println(gd.pats)
	//for i, pat := range gd.pats {
	//	fmt.Println(i, pat)
	//}
	// 首字符
	codes = gd.prefixRune(ws, -1, 0)
	stacks = append(stacks, buildStack{len(codes), 0, 0})
	if err = gd.insertCodes(0, codes); err != nil {
		return err
	}

	for cursor < len(ws) {
		w := ws[cursor]
		stack = stacks[len(stacks)-1]
		if stack.endRow == 1 {
			stacks = stacks[0 : len(stacks)-1]
		} else {
			stacks[len(stacks)-1].endRow--
		}
		//fmt.Printf("cursor: %d string to be inserted: %v stack: %v stacks: %v base[stack.s]: %d\n", cursor, w, stack, stacks, gd.base[stack.s])
		w = w[stack.column:]
		if gd.base[stack.s] < 0 {
			s = -1*gd.base[stack.s] + gd.auxiliary[w[0]]
		} else {
			//println("cursor:", cursor, "w length:", len(w), stack.column, len(ws[cursor]), string(ws[cursor]))
			s = gd.base[stack.s] + gd.auxiliary[w[0]]
		}
		if s >= len(gd.base) {
			if err = gd.extend(0); err != nil {
				return err
			}
		}
		// s是当前字符的位置, t是下一个字符的位置
		for i, r := range w {
			// codes 不包括r
			codes = gd.prefixRune(ws[cursor:], i+stack.column, r)

			//fmt.Printf("    i=%d w=%v s=%d t=%d r=%v(%d) codes=%v\n", i, w, s, t, r, gd.auxiliary[r], codes)
			if i == len(w)-1 {
				// 所有模式中，无与当前相同的前缀
				// 最后一个字符
				if len(codes) == 0 {
					gd.base[s] = DAT_END_POS
				} else {
					gd.base[s] = -gd.base[s]
					stacks = append(stacks, buildStack{len(codes), stack.column + i + 1, s})
					if err = gd.insertCodes(s, codes); err != nil {
						return err
					}
				}

				//fmt.Printf("    i=%d s=%d t=%d r=%v  codes=%v stack.column=%d\n", i, s, t, r, codes, stack.column)
				//fmt.Println("   ", gd.base, "\n   ", gd.check)
				//fmt.Println("    stacks:", len(stacks), stacks)
				break
			} else {
				allcodes := codes
				nextChar := gd.auxiliary[w[i+1]]
				if len(codes) == 0 {
					allcodes = []int{nextChar}
				} else {
					if i != len(w)-1 {
						allcodes = append(codes, nextChar) //gd.auxiliary[r])
					}
					stacks = append(stacks, buildStack{len(codes), stack.column + i + 1, s})
				}
				if err = gd.insertCodes(s, allcodes); err != nil {
					return err
				}
				t = gd.base[s] + nextChar //gd.auxiliary[r]
			}

			//fmt.Printf("    i=%d s=%d t=%d r=%v  codes=%v stack.column=%d\n", i, s, t, r, codes, stack.column)
			//fmt.Println("   ", gd.base, "\n   ", gd.check)
			//fmt.Println("    stacks:", len(stacks), stacks)

			s = t
		}

		cursor++
		if cursor%1000 == 0 {
			fmt.Printf("build %d patterns\n", cursor)
		}

	}
	return nil
}

// 比较两个slice中的每一个元素是否都相同
// 长度相同
func equalSlic(src, dst Words) bool {
	for i, _ := range src {
		if src[i] != dst[i] {
			return false
		}
	}
	return true
}

// ws已经排过序，且ws的第一条记录应该是符合ws[startPos]==r的
// ws中有多少条记录ws[startPos]==r
// endRow: 记录条数
// codes: 符合条件的Words的下一个rune对应的code
// 特殊情况：
//  startPos = -1:
//  返回值endRow=1，codes=[]，只有ws[0]符合条件，且ws[0]以r结尾
func (gd *GoDat) prefixRune(ws []Words, startPos int, r rune) []int {
	var (
		i        int
		w        Words
		nextChar rune
		rm       = make(map[rune]bool)
		codes    = make([]int, 0)
	)

	if len(ws) <= 1 {
		return codes
	}

	if startPos != -1 && startPos <= len(ws[0])-2 {
		nextChar = ws[0][startPos+1]
	}
	//fmt.Printf("    +++prefixRune: ws=%v startPos=%v nextChar=%v r=%v\n", ws, startPos, nextChar, r)

	prefix := ws[0]
	if startPos != -1 {
		i = 1
	}
	for ; i < len(ws); i++ {
		w = ws[i]
		if len(w) < startPos+1 || (startPos >= 0 && equalSlic(w[0:startPos+1], prefix[0:startPos+1]) == false) {
			break
		}
		if len(w) > startPos+1 {
			r := w[startPos+1]
			if r == nextChar {
				continue
			}
			if _, ok := rm[r]; !ok {
				rm[r] = true
				codes = append(codes, gd.auxiliary[w[startPos+1]])
			}
		}
	}

	return codes
}

// codes已经排好序
// 为code中的每一个值，选择一个位置
// 确定s位置的base值
// 类似于findNewBase
func (gd *GoDat) insertCodes(s int, codes []int) error {
	var (
		start = gd.check[0]
		pos   int
		base  int
		found bool
		err   error
	)
	if len(codes) == 0 {
		return nil
	}

	sort.Sort(sort.IntSlice(codes))

	if gd.idles == 0 {
		if err = gd.extend(0); err != nil {
			return err
		}
	}

	for pos = start; ; pos = -1 * gd.check[pos] {
		if -1*gd.check[pos] == start {
			if err = gd.extend(0); err != nil {
				return err
			}
		}
		base = pos - codes[0]
		if (s == 0 && base < 0) || (s > 0 && base <= 0) {
			continue
		}
		found = true
		for i := 1; i < len(codes); i++ {
			if base+codes[i] >= len(gd.check) {
				if err = gd.extend(base + codes[i]); err != nil {
					return err
				}
			}
			if gd.check[base+codes[i]] >= 0 {
				found = false
				break
			}
		}
		if found {
			break
		}
	}

	// found
	//fmt.Println("    ----base:", base, ", s:", s, ", codes:", codes, len(gd.base))
	gd.base[s] = base
	for _, c := range codes {
		//fmt.Println("    --------delLink:", base+c, gd.base[base+c], gd.check[base+c])
		gd.delLink(base + c)
		gd.check[base+c] = s
		gd.base[base+c] = 0
	}
	return nil
}
