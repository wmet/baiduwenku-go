package filetype

import (
	"github.com/gufeijun/baiduwenku/utils"
	"os"
	"strings"
)

func StartDocSpider(rawurl string)(string,error){
	sli,title,err:=parseDocRawURL(rawurl)
	if err!=nil{
		return "",err
	}
	var str string
	for _,val:=range sli{
		doc,err:=utils.QuickSpider(val)
		if err!=nil{
			return "",err
		}
		res,err:=utils.QuickRegexp(doc,`{"c":"(.*?)".*?,"y":(.*?),.*?"ps":(.*?),`)
		if err!=nil{
			return "",err
		}
		for _,val:=range res{
			if val[3]!="null"{
				str+="\n"+utils.UnicodeToUTF(val[1])
			}else{
				str+=utils.UnicodeToUTF(val[1])
			}
		}
	}
	f,err:=os.Create(title+".doc")
	if err!=nil{
		return "",err
	}
	f.WriteString(str)
	f.Close()
	return title+".doc",nil
}

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