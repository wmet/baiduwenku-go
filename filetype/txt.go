package filetype

import (
	"errors"
	"github.com/gufeijun/baiduwenku/utils"
	"os"
	"regexp"
	"strings"
)

func StartTxtSpider(rawurl string)(string,error){
	url,title,err:=parseTxtRawURL(rawurl)
	if err!=nil{
		return "",err
	}
	data,err:=utils.QuickSpider(url)
	if err!=nil{
		return "",err
	}
	f,_:=os.Create(title+".txt")
	defer f.Close()
	str:=utils.UnicodeToUTF(data)
	str,err=extract(str)
	if err!=nil{
		return "",err
	}
	f.WriteString(str)
	return title+".txt",nil
}

//处理获得的txt文本
func extract(str string)(e string,err error){
	reg:=regexp.MustCompile(`"c":"(.*?)"`)
	res:=reg.FindAllStringSubmatch(str,-1)
	if len(res)==0{
		return "",errors.New("No Submatch")
	}
	for _,val:=range res{
		temps:=strings.Split(val[1],"\\r\\n")
		for _,v:=range temps{
			if v==""{
				continue
			}
			e+=v+"\r\n"
		}
	}
	return
}

func parseTxtRawURL(rawurl string)(string,string,error){
	//dom为静态网页的源代码
	dom,err:=utils.QuickSpider(rawurl)
	if err!=nil{
		return "","",err
	}
	//获取文章的docID
	res,err:=utils.QuickRegexp(rawurl,`view/(.*?).html`)
	if err!=nil{
		return "","",err
	}
	docID:=res[0][1]
	//获取文章标题
	res,err=utils.QuickRegexp(dom,` 'title': '(.*?)',`)
	if err!=nil{
		return "","",err
	}
	title:=utils.Gbk2utf8(res[0][1])
	//获取文件格式
	res,err=utils.QuickRegexp(dom,`'docType': '(.*?)',`)
	if err!=nil{
		return "","",err
	}
	docType:=res[0][1]
	//文档的页数
	res,err=utils.QuickRegexp(dom,`'totalPageNum': '(.*?)',`)
	if err!=nil{
		return "","",err
	}
	totalPageNum:=res[0][1]
	//文章信息链接用来获取md5sum和rsign
	docInfoURl:="https://wenku.baidu.com/api/doc/getdocinfo?callback=cb&doc_id="+docID
	body,err:=utils.QuickSpider(docInfoURl)
	if err!=nil{
		return "","",err
	}
	//获取MD5sum
	res,err=utils.QuickRegexp(body,`md5sum":"&(.*?)"`)
	if err!=nil{
		return "","",err
	}
	md5sum:=res[0][1]
	//获取rsign
	res,err=utils.QuickRegexp(body,`rsign":"(.*?)"`)
	if err!=nil{
		return "","",err
	}
	rsign:=res[0][1]
	fmtUrl:="https://wkretype.bdimg.com/retype/text/"
	fmtUrl=fmtUrl+docID+"?"+md5sum+"&callback=cb&pn=1&rn="+totalPageNum+"&type="+docType+"&rsign="+rsign+"&_=1588768641115"
	return fmtUrl,title,nil
}


