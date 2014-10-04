godat
=====

double array trie algorithm golang

# 构建
## 动态构建(在词条较多时，构建时间很长)

动态构建在词条较多时，耗时很长。测试使用本目录下dictionary.txt中的58万多条数据构建时，需要20分钟以上。

`
	func CreateGoDat(pats []string, nocase bool) (gd *GoDat, err error)
`


## 无冲突构建

无冲突构建很快，测试使用本目录下dictionary.txt中的58万多条数据构建时，不到2分钟即可完成

`   
	gd := GoDat{pats: pats}
	
	gd.Initialize(true)
	
	gd.BuildWithoutConflict()
`

# 查找

`
	func (gd *GoDat) Match(noodle string) bool
`
