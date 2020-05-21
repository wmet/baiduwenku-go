package filetype

import (
	"github.com/gufeijun/baiduwenku/utils"
	"io/ioutil"
	"os"
	"strings"
)

func StartDocSpider(rawurl string)(string,error){
	//ch用于存放文档数据url
	ch:=make(chan string,10)

	title,err:=parseDocRawURL(rawurl,ch)
	if err!=nil{
		return "",err
	}
	//如果已经存在该文件，直接返回
	if _,err:=os.Stat(title+".doc");err==nil{
		return title+".doc",nil
	}

	var str string
	for url:=range ch{
		doc,err:=utils.QuickSpider(url)
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

func parseDocRawURL(rawurl string,ch chan<- string)(string,error){
	doc,err:=utils.QuickSpider(rawurl)
	if err!=nil{
		return "",err
	}
	t,err:=utils.QuickRegexp(doc,`docTitle: '(.*?)',`)
	if err!=nil{
		return "",err
	}
	title:=utils.Gbk2utf8(t[0][1])
	res,err:=utils.QuickRegexp(doc,`https:(.*?).json?(.*?)\\x22}`)
	if err!=nil{
		return "",err
	}

	go func(){
		for i:=0;i<len(res)/2;i++{
			ch<-strings.Replace(res[i][0][:len(res[i][0])-5],`\\\`,"",-1)
		}
		close(ch)
	}()
	return title,nil
}