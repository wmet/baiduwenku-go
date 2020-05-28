package utils

import (
	"github.com/axgle/mahonia"
	"strconv"
	"strings"
)
/*
		一些文本编码转换的函数
 */

//GBK转utf-8
func Gbk2utf8(src string)string{
	srcCoder := mahonia.NewDecoder("gbk")
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder("utf-8")
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}

//unicode转utf-8
func UnicodeToUTF(s string)string{
	sli:=strings.Split(s,"\\u")
	str:=sli[0]
	for _, v := range sli[1:]{
		if len(v) <4 {
			str+=v
			continue
		}
		temp, _ := strconv.ParseInt(v[:4], 16, 32)
		str+=string(temp)
		if len(v)>4{
			str+=v[4:]
		}
	}
	return str
}