package godat

import (
	"strings"
)

//回溯

func (gd *GoDat) backtrace(s int) string {
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

	rev := ""
	for i = len(res) - 1; i >= 0; i-- {
		rev += res[i]
	}

	return rev
}

func (gd *GoDat) howmany() {

}
