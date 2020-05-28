package crawl

import (
	"github.com/gufeijun/baiduwenku/utils"
	"io/ioutil"
	"os"
	"strings"
)

func StartDocSpider(rawurl string)(string,error){
	//ch用于存放文档数据url
	ch:=make(chan string,10)

	//如果是vip免费文档直接调用第二种下载方式
	if loction,ok:=utils.PrePrecess(rawurl);ok{
		return loction,nil
	}

	title,err:=parseDocRawURL(rawurl,ch)
	if err!=nil{
		return "",err
	}
	//如果已经存在该文件，直接返回
	if _,err:=os.Stat(title+".doc");err==nil{
		return title+".doc",nil
	}

	var str string

	//遍历管道
	for url:=range ch{
		//对url进行请求
		doc,err:=utils.QuickSpider(url)
		if err!=nil{
			return "",err
		}

		//正则匹配数据
		res,err:=utils.QuickRegexp(doc,`{"c":"(.*?)".*?"y":(.*?),.*?"ps":(.*?),`)
		if err!=nil{
			return "",err
		}

		//pre_y记录上一行的纵坐标
		pre_y:=res[0][2]
		for _,val:=range res{
			//三种情况要换行，不要问我怎么写出这坨翔一样的东西，想死的心都有了
			//1、如果ps值为{"_enter":1}则代表文本需要换行
			if val[3]==`{"_enter":1}`{
				str+=utils.UnicodeToUTF(val[1])+"\n"
			}else{
				//2、str最后一位为" "且该行与上一行的y坐标不同则换行
				//3、str最后一位为换行符，倒数第3位为" "则换行
				if len(str)>1&&str[len(str)-1:]==" "&&val[2]!=pre_y{
					str+="\n"
				}else if len(str)>2&&str[len(str)-1:]=="\n"&&str[len(str)-3:len(str)-2]==" "{
					str+="\n"
				}
				str+=utils.UnicodeToUTF(val[1])
			}
			pre_y=val[2]
		}
	}

	//解决多个转义字符以及双引号输出问题
	str=strings.Replace(str,`\/`,`/`,-1)
	str=strings.Replace(str,"\\","\"",-1)

	if err:=ioutil.WriteFile(title+".doc",[]byte(str),0666);err!=nil{
		return "",err
	}
	return title+".doc",nil
}

func parseDocRawURL(rawurl string,ch chan<- string)(string,error){
	//发起http请求
	doc,err:=utils.QuickSpider(rawurl)
	if err!=nil{
		return "",err
	}

	//获取文档标题
	t,err:=utils.QuickRegexp(doc,`docTitle: '(.*?)',`)
	if err!=nil{
		return "",err
	}
	title:=utils.Gbk2utf8(t[0][1])

	//正则匹配最终数据的url
	res,err:=utils.QuickRegexp(doc,`https:(.*?).json?(.*?)\\x22}`)
	if err!=nil{
		return "",err
	}

	go func(){
		for i:=0;i<len(res)/2;i++{
			//存入管道，交给父程处理
			ch<-strings.Replace(res[i][0][:len(res[i][0])-5],`\\\`,"",-1)
		}
		close(ch)
	}()
	return title,nil
}