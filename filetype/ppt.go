package filetype

import (
	"github.com/gufeijun/baiduwenku/utils"
	"os"
	"strconv"
	"strings"
)

func StartPPTSpider(rawurl string)(string,error){
	sl,title,err:=parsePPTRawURL(rawurl)
	if err!=nil{
		return "",err
	}
	filenames:=make([]string,len(sl))
	for i:=0;i<len(sl);i++{
		buf,err:=utils.GetJPG(sl[i])
		if err!=nil{
			continue
		}
		filenames[i]=title+strconv.Itoa(i)+".jpg"
		f,_:=os.Create(filenames[i])
		f.Write(buf)
		f.Close()
	}
	defer func() {
		for _,val:=range filenames{
			os.Remove(val)
		}
	}()
	return title+".zip",utils.ZipFiles(title+".zip",filenames)
}

func parsePPTRawURL(rawurl string)([]string,string,error){
	doc,err:=utils.QuickSpider(rawurl)
	if err!=nil{
		return nil,"",err
	}
	t,err:=utils.QuickRegexp(doc,`'title': '(.*?)',`)
	if err!=nil{
		return nil,"",err
	}
	title:=utils.Gbk2utf8(t[0][1])
	infoURL:="https://wenku.baidu.com/browse/getbcsurl?doc_id="+utils.GetDocID(rawurl)+"&pn=1&rn=99999&type=ppt"
	doc,err=utils.QuickSpider(infoURL)
	doc=strings.Replace(doc,`\/`,`/`,-1)
	res,err:=utils.QuickRegexp(doc,`"zoom":"(.*?)",`)
	if err!=nil{
		return nil,"",err
	}
	s:=make([]string,len(res))
	for i:=0;i<len(res);i++{
		s[i]=res[i][1]
	}
	return s,title,nil
}