package utils

import (
	"archive/zip"
	"errors"
	"github.com/axgle/mahonia"
	"io"
	"io/ioutil"
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

func QuickRegexp(raw string,patten string) ([][]string,error){
	reg:=regexp.MustCompile(patten)
	res:=reg.FindAllStringSubmatch(raw,-1)
	if len(res)==0{
		return nil,errors.New("No Submatch")
	}
	return res,nil
}

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

func GetDocID(rawurl string)string{
	res,_:=QuickRegexp(rawurl,`view/(.*?).html`)
	return res[0][1]
}

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

func ZipFiles(filename string, files []string) error {
	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()
	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()
	for _, file := range files {
		if err = AddFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

func AddFileToZip(zipWriter *zip.Writer, filename string) error {
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




