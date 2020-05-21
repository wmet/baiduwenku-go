package utils

import (
	"archive/zip"
	"errors"
	"github.com/axgle/mahonia"
	"github.com/go-gomail/gomail"
	"github.com/gufeijun/baiduwenku/config"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func Gbk2utf8(src string)string{
	srcCoder := mahonia.NewDecoder("gbk")
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder("utf-8")
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}

//正则的封装
func QuickRegexp(raw string,patten string) ([][]string,error){
	reg:=regexp.MustCompile(patten)
	res:=reg.FindAllStringSubmatch(raw,-1)
	if len(res)==0{
		return nil,errors.New("No Submatch")
	}
	return res,nil
}

//爬虫的封装
func QuickSpider(url string)(string,error){
	cli:=&http.Client{}
	req,err:=http.NewRequest("GET",url,nil)
	if err!=nil{
		return "",err
	}
	req.Header.Set("User-Agent","Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36")
	resp,err:=cli.Do(req)
	if err!=nil{
		return "",err
	}
	defer resp.Body.Close()
	buf,err:=ioutil.ReadAll(resp.Body)
	if err!=nil{
		return "",err
	}
	return string(buf),nil
}

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

//获取文档的id
func GetDocID(rawurl string)string{
	res,_:=QuickRegexp(rawurl,`view/(.*?).html`)
	return res[0][1]
}

//下载图片
func GetJPG(url string) ([]byte,error){
	cli:=&http.Client{}
	req,err:=http.NewRequest("GET",url,nil)
	if err!=nil{
		return nil,err
	}
	req.Header.Set("User-Agent","Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36")
	resp,err:=cli.Do(req)
	if err!=nil{
		return nil,err
	}
	defer resp.Body.Close()
	buf,err:=ioutil.ReadAll(resp.Body)
	if err!=nil{
		return nil,err
	}
	return buf,nil
}

//打包文件
func ZipFiles(filename string, files []string) error {
	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()
	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()
	for _, file := range files {
		if err = addFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

func addFileToZip(zipWriter *zip.Writer, filename string) error {
	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = filename
	header.Method = zip.Deflate
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

//SendCode 发送验证码
func SendCode(emailadd string, code string) {
	m := gomail.NewMessage()

	m.SetAddressHeader("From", config.SeverConfig.IMAP_EMAIL/*"发件人地址"*/, "百度文库下载平台") // 发件人

	m.SetHeader("To", m.FormatAddress(emailadd, "收件人")) // 收件人

	m.SetHeader("Subject", "来自百度文库下载平台的验证码") // 主题

	m.SetBody("text/plain", "您的验证码为:"+code) // 正文

	d := gomail.NewPlainDialer(config.SeverConfig.IMAP_SERVER,  config.SeverConfig.IMAP_PORT, config.SeverConfig.IMAP_EMAIL, config.SeverConfig.IMAP_PASSWORD) // 发送邮件服务器、端口、发件人账号、发件人密码
	if err := d.DialAndSend(m); err != nil {
		log.Println("发送失败", err)
		return
	}
	log.Println("Have sent email to "+emailadd)
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
	res,err:=QuickRegexp(url,`view/(.*?).html`)
	infos[0]=res[0][1]  //文档id
	if err!=nil{
		return
	}
	res,err=QuickRegexp(doc,`'docType': '(.*?)',`)
	if err!=nil{
		return
	}
	filetype:=res[0][1] //文档类型
	res,err=QuickRegexp(doc,` 'title': '(.*?)',`)
	if err!=nil{
		return
	}
	title:=Gbk2utf8(res[0][1])
	infos[1]=title+"."+filetype//文档名称
	res,err=QuickRegexp(doc,`"downloadToken" value="(.*?)"`)
	if err!=nil{
		return
	}
	infos[2]=res[0][1]  //下载的Token
	res,err=QuickRegexp(doc,`'professionalDoc': '(.*?)'`)
	if err!=nil{
		return
	}
	ifprofession=res[0][1]=="1"  //是否是专享文档
	return
}