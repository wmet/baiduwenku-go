package utils

import (
	"errors"
	"github.com/gufeijun/baiduwenku/config"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

/*
		一些用于获取下载文档信息的函数
 */

//获取文件的类别
func GetDocType(url string)(string,error){
	doc,err:=QuickSpider(url)
	if err!=nil{
		return "",err
	}
	res,err:=QuickRegexp(doc,`'docType': '(.*?)',`)
	if err!=nil{
		return "",err
	}
	return res[0][1],nil
}

//获取文档的id
func GetDocID(rawurl string)string{
	res,_:=QuickRegexp(rawurl,`view/(.*?).html`)
	return res[0][1]
}

//还剩多少专享文档下载券
func GetDownloadTicket()(num int,err error){
	client:=&http.Client{}
	req,err:=http.NewRequest("GET","https://wenku.baidu.com/customer/interface/getuserdownloadticket",nil)
	if err!=nil{
		return
	}
	req.Header.Set("User-Agent","Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")
	cookie:=&http.Cookie{
		Name: "BDUSS",
		Value: config.SeverConfig.BDUSS,
	}
	req.AddCookie(cookie)
	resp,err:=client.Do(req)
	if err!=nil{
		return
	}
	defer resp.Body.Close()
	buf,err:=ioutil.ReadAll(resp.Body)
	if err!=nil{
		return
	}
	reg:=regexp.MustCompile(`"pro_download_ticket":(.*?),"`)
	res:=reg.FindAllStringSubmatch(string(buf),-1)
	if len(res)==0{
		return 0,errors.New("Can Not Get Ticket Information!")
	}
	return strconv.Atoi(res[0][1])
}

//获取文档的信息
func GetInfos(url string)(infos []string,ifprofession bool,err error){
	infos=make([]string,3)
	cli:=&http.Client{}
	req,err:=http.NewRequest("GET",url,nil)
	req.Header.Set("User-Agent","Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36")
	cookie:=&http.Cookie{
		Name: "BDUSS",
		Value: config.SeverConfig.BDUSS,
	}
	req.AddCookie(cookie)
	resp,err:=cli.Do(req)
	if err!=nil{
		return
	}
	defer resp.Body.Close()
	buf,_:=ioutil.ReadAll(resp.Body)
	doc:=string(buf)

	//获取文档id
	res,err:=QuickRegexp(url,`view/(.*?).html`)
	infos[0]=res[0][1]
	if err!=nil{
		return
	}

	//获取文档的类型
	res,err=QuickRegexp(doc,`'docType': '(.*?)',`)
	if err!=nil{
		return
	}
	filetype:=res[0][1]

	//获取文档的标题
	res,err=QuickRegexp(doc,` 'title': '(.*?)',`)
	if err!=nil{
		return
	}
	title:=Gbk2utf8(res[0][1])
	infos[1]=title+"."+filetype

	//获取下载token
	res,err=QuickRegexp(doc,`"downloadToken" value="(.*?)"`)
	if err!=nil{
		return
	}
	infos[2]=res[0][1]

	//是否是专享文档
	res,err=QuickRegexp(doc,`'professionalDoc': '(.*?)'`)
	if err!=nil{
		return
	}
	ifprofession=res[0][1]=="1"

	return
}

//IsVIPfreeDoc 判断该文档是否为vip免费文档
func IsVIPfreeDoc(url string)(ok bool,err error){
	docID:=GetDocID(url)
	url="https://wenku.baidu.com/user/interface/getvipfreedoc?doc_id="+docID
	doc,err:=QuickSpider(url)
	if err!=nil{
		return
	}
	return !strings.Contains(doc,"false"),nil
}

