package filetype

import (
	"github.com/gufeijun/baiduwenku/utils"
	"io/ioutil"
	"os"
	"strings"
)

func StartDocSpider(rawurl string)(string,error){
	sli,title,err:=parseDocRawURL(rawurl)
	if err!=nil{
		return "",err
	}
	//如果已经存在该文件，直接返回
	if _,err:=os.Stat(title+".doc");err==nil{
		return title+".doc",nil
	}
	var str string
	for _,val:=range sli{
		doc,err:=utils.QuickSpider(val)
		if err!=nil{
			return "",err
		}
		res,err:=utils.QuickRegexp(doc,`{"c":"(.*?)".*?"ps":(.*?),`)
		if err!=nil{
			return "",err
		}
		for _,val:=range res{
			//如果ps值不为null则代表文本需要换行
			if val[2]!="null"{
				str+="\n"+utils.UnicodeToUTF(val[1])
			}else{
				str+=utils.UnicodeToUTF(val[1])
			}
		}
	}
	if err:=ioutil.WriteFile(title+".doc",[]byte(str),0666);err!=nil{
		return "",err
	}
	return title+".doc",nil
}

//获取文档名称，以及文档数据url地址(有多个)，切片形式保存
func parseDocRawURL(rawurl string)([]string,string,error){
	doc,err:=utils.QuickSpider(rawurl)
	if err!=nil{
		return nil,"",err
	}
	t,err:=utils.QuickRegexp(doc,`docTitle: '(.*?)',`)
	if err!=nil{
		return nil,"",err
	}
	title:=utils.Gbk2utf8(t[0][1])
	res,err:=utils.QuickRegexp(doc,`https:(.*?).json?(.*?)\\x22}`)
	if err!=nil{
		return nil,"",err
	}
	sli:=make([]string,len(res)/2)
	for i:=0;i<len(res)/2;i++{
		sli[i]=strings.Replace(res[i][0][:len(res[i][0])-5],`\\\`,"",-1)
	}
	return sli,title,nil
}