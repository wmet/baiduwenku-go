package crawl

import (
	"github.com/gufeijun/baiduwenku/utils"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
)

func StartPPTSpider(rawurl string)(string,error){
	//如果是vip免费文档直接调用第二种下载方式
	if loction,ok:=utils.PrePrecess(rawurl);ok{
		return loction,nil
	}

	//ch用于存放图片url
	ch:=make(chan string,10)

	//limit限制开启go程的数目为10
	limit:=make(chan interface{},10)

	title,err:=parsePPTRawURL(rawurl,ch)
	if err!=nil{
		return "",err
	}

	//如果已经存在该文件，直接返回
	if _,err:=os.Stat(title+".zip");err==nil{
		return title+".zip",nil
	}

	//i用于记录有多少张图片
	i:=0

	var wg sync.WaitGroup
	for url:=range ch{
		wg.Add(1)
		//limit满时阻塞
		limit<-struct{}{}

		go func(wg *sync.WaitGroup,i int,url string) {
			defer wg.Done()
			buf,_:=utils.GetJPG(url)
			ioutil.WriteFile(title+strconv.Itoa(i)+".jpg",buf,0666)
			//go程完成任务，则释放一个空位
			<-limit
		}(&wg,i,url)
		i++
	}
	//将图片的文件名存入切片
	filenames:=make([]string,i)
	for j:=0;j<i;j++{
		filenames[j]=title+strconv.Itoa(j)+".jpg"
	}
	//同步，等待图片处理完
	wg.Wait()

	//打包图片
	if err:=utils.ZipFiles(title+".zip",filenames);err!=nil{
		return "",err
	}

	//图片打包完成后删除所有图片
	go func() {
		for _,val:=range filenames{
			os.Remove(val)
		}
	}()

	//将图片打包为zip文件
	return title+".zip",nil
}

//获取所有图片的url以及文件的名称
func parsePPTRawURL(rawurl string,ch chan<- string)(string,error){
	//发起http请求
	doc,err:=utils.QuickSpider(rawurl)
	if err!=nil{
		return "",err
	}

	//获取文档标题
	t,err:=utils.QuickRegexp(doc,`'title': '(.*?)',`)
	if err!=nil{
		return "",err
	}
	title:=utils.Gbk2utf8(t[0][1])

	//利用go程来得到多个图片的url，文章title先返回给父程
	go func(ch chan<- string,rawurl string) {
		defer close(ch)
		infoURL:="https://wenku.baidu.com/browse/getbcsurl?doc_id="+utils.GetDocID(rawurl)+"&pn=1&rn=99999&type=ppt"
		doc,err:=utils.QuickSpider(infoURL)
		if err!=nil{
			return
		}
		doc=strings.Replace(doc,`\/`,`/`,-1)
		res,_:=utils.QuickRegexp(doc,`"zoom":"(.*?)",`)
		for i:=0;i<len(res);i++{
			//将图片url存入管道，让父程进行处理
			ch<-res[i][1]
		}
	}(ch,rawurl)
	return title,nil
}