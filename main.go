package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gufeijun/baiduwenku/filetype"
	"github.com/gufeijun/baiduwenku/utils"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var a map[string]time.Time = make(map[string]time.Time)

func main(){
	go Timer()
	router := gin.Default()
	router.Static("/static", "front-end")
	router.LoadHTMLGlob("front-end/html/*")
	router.GET("/baiduspider", func(c *gin.Context) {
		c.HTML(http.StatusOK,"home.html",nil)
	})
	router.POST("/baiduspider",func(c *gin.Context){
		url,ok:=c.GetPostForm("url")
		if !ok{
			c.JSON(http.StatusOK,gin.H{
				"status":"0",
				"err":"Can Not Parse URL!",
			})
			return
		}
		filepath,err:=spider(url)
		a[filepath]=time.Now()
		if err!=nil{
			c.JSON(http.StatusOK,gin.H{
				"status":"0",
				"err":err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK,gin.H{
			"status":"1",
			"path":filepath,
		})
	})
	router.GET("/download", func(c *gin.Context) {
		name,ok:=c.GetQuery("file")
		if !ok{
			c.String(http.StatusBadRequest,"illegal!")
			return
		}
		f,err:=os.Open(name)
		if err!=nil{
			c.String(http.StatusBadRequest,"No Such File!")
			return
		}
		defer f.Close()
		buf,_:=ioutil.ReadAll(f)
		c.Writer.WriteHeader(http.StatusOK)
		c.Header("Content-Disposition", "attachment; filename="+name)
		c.Header("Content-Type", "application/octet-stream")
		c.Writer.Write(buf)
		//go os.Remove(name)
	})
	router.Run("0.0.0.0:9999")
}

func spider(url string)(string,error) {
	docType,err:=utils.GetDocType(url)
	if err!=nil{
		fmt.Println(err)
		return "",errors.New("老夫暂时拿此链接无能为力~（´Д`）")
	}
	switch docType {
	case "txt":
		return filetype.StartTxtSpider(url)
	case "doc":
		return filetype.StartDocSpider(url)
	case "pdf":
		return filetype.StartPdfSpider(url)
	case "ppt":
		return filetype.StartPPTSpider(url)
	default:
		return "",errors.New(fmt.Sprintf("Do Not Support filetype:%s!",docType))
	}
	return "",nil
}

//定时器
func Timer(){
	for{
		time.Sleep(10*time.Minute)
		m,_:=time.ParseDuration("-1m")
		now:=time.Now()
		for key,val:=range a{
			val=val.Add(10*m)
			if now.After(val){
				os.Remove(key)
			}
		}
	}
}






