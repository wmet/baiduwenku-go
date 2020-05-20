package filetype

import (
	"github.com/gufeijun/baiduwenku/utils"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func StartPPTSpider(rawurl string)(string,error){
	sl,title,err:=parsePPTRawURL(rawurl)
	if err!=nil{
		return "",err
	}
	//如果已经存在该文件，直接返回
	if _,err:=os.Stat(title+".zip");err==nil{
		return title+".zip",nil
	}
	filenames:=make([]string,len(sl))
	for i:=0;i<len(sl);i++{
		buf,err:=utils.GetJPG(sl[i])
		if err!=nil{
			continue
		}
		ioutil.WriteFile(filenames[i],buf,0666)
		filenames[i]=title+strconv.Itoa(i)+".jpg"
	}
	//文件打包好后删除
	defer func() {
		for _,val:=range filenames{
			os.Remove(val)
		}
	}()
	//将图片打包为zip文件
	return title+".zip",utils.ZipFiles(title+".zip",filenames)
}

//获取所有图片的url以及文件的名称
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