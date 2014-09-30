package godat

import (
	"bufio"
	//"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"testing"
)

func TestBuildDat(t *testing.T) {
	pats := []string{
		/*"abcd",
		"aa", "aaa",
		"ad",
		"aceiw",
		"djkafiew",
		"asd", "aglmnqioew",
		"bbbbb",
		"c",
		"bbbbbbae",
		"http://www.sina.cn", "alpha", "aaa", "zzbc", "fals", "hi!", "ab", "cc", "ca", "sets",
		"wow", "baa", "ma", "mm",
		"how", "bcefd", "apple", "google", "ms", "tencent", "baidu", "axon",
		"全國假日旅遊部際協調會議辦公室", "中华人民共和国全国人民代表大会",
		"中華人民共和國全國人民代表大會", "全国高等教育自学考试指导委员会",
		"全國高等教育自學考試指導委員會", "数两千零七十一万六千零三十七户",
		"數兩千零七十一萬六千零三十七戶", "劳动和社会保障部劳动科学研究所",
		"勞動和社會保障部勞動科學研究所", "联合国教科文组织世界遗产委员会",
		"聯合國教科文組織世界遺產委員會", "中国人民政治协商会议全国委员会",
		"中國人民政治協商會議全國委員會", "积石山保安族东乡族撒拉族自治县",
		"積石山保安族東鄉族撒拉族自治縣", "香港特別行政區基本法起草委員會",
		"香港特别行政区基本法起草委员会", "八千一百三十七万七千二百三十六口",
		"八千一百三十七萬七千二百三十六口", "第九屆全國人民代表大會常務委員會",
		"第九届全国人民代表大会常务委员会", "劳动和社会保障部职业技能鉴定中",
		*/
		"1号店",
		"1號店",
		"4S店",
		"4s店",
		"AA制",
		"AB型",
		"AT&T",
		"A型",
		"A座",
		"A股",
		"A輪",
		"A轮",
		"BB机",
		"BB機",
		"BP机",
		"BP機",
		"B型",
		"B座",
		"B股",
		"B超",
		"B輪",
		"B轮",
	}
	gd, err := CreateGoDat(pats, true)
	if err != nil {
		t.Fatal("create dat failed:", err)
	}
	gd.dump()

	for _, pat := range gd.pats {
		if gd.Match(pat) == false {
			t.Fatal("match should be true:" + pat)
		}
		//if gd.Match(pat+"!") == true {
		//	t.Fatal("match should be false")
		//}
	}
}

func TestBuildDict(t *testing.T) {
	gd := GoDat{}
	f, err := os.Open("./dictionary.txt")
	if err != nil {
		t.Fatal("open dictionary file failed!\n")
	}
	rd := bufio.NewReader(f)
	for line, err := rd.ReadString(byte('\n')); err == nil || err != io.EOF; line, err = rd.ReadString(byte('\n')) {
		segs := strings.Split(line, " ")
		if len(segs) != 3 {
			continue
		}
		gd.add(segs[0])
	}
	gd.initialize()
	gd.dump()
	gd.build()
}

func testExtend(t *testing.T) {
	gd := GoDat{}
	gd.base = make([]int, initArrayLen)
	gd.check = make([]int, initArrayLen)
	gd.maxLen = 1024
	gd.initLink()
	gd.dump()
	if err := gd.extend(); err != nil {
		t.Fatal(err)
	}
	gd.dump()
}

func testBinSearch(t *testing.T) {
	gd := GoDat{
		pats: []string{"aaa", "zzbc", "fals", "hi!", "ab", "cc", "ca", "sets",
			"abcd", "wow", "baa", "ma", "mm",
			"how", "bcefd", "apple", "google", "ms", "tencent", "baidu", "axon"},
	}
	sort.Sort(sort.StringSlice(gd.pats))

	for i := 0; i < len(gd.pats); i++ {
		pat := gd.pats[i]
		s1 := gd.binSearch(pat)
		if i != s1 {
			t.Fatal("bin search failed at", i, s1, "\n")
		}
	}
}
