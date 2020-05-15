package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gufeijun/baiduwenku/config"
	"github.com/gufeijun/baiduwenku/filetype"
	"github.com/gufeijun/baiduwenku/mediumware"
	"github.com/gufeijun/baiduwenku/model"
	"github.com/gufeijun/baiduwenku/utils"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var a map[string]time.Time = make(map[string]time.Time)

func main(){
	go Timer()
	router := gin.Default()
	router.Static("/static", "front-end")
	router.LoadHTMLGlob("front-end/html/*.html")
	//主页面
	router.GET("/baiduspider", func(c *gin.Context) {
		cookie, _ := c.Request.Cookie("sessionid")
		var emailadd string
		if cookie!=nil{
			sessionid := cookie.Value
			query := "select emailadd from hustsessions where sessionid=?"
			row := config.Db.QueryRow(query, sessionid)
			row.Scan(&emailadd)
			emailadd=strings.Split(emailadd,"@")[0]
		}
		remain,_:=utils.GetDownloadTicket()
		c.HTML(http.StatusOK,"home.html",struct {
			Emailadd string
			Remain int
		}{emailadd,remain})
	})
	//文件下载api
	router.POST("/baiduspider",func(c *gin.Context){
		url,ok:=c.GetPostForm("url")
		if !ok{
			c.JSON(http.StatusOK,gin.H{
				"status":"0",
				"err":"Can Not Parse URL!",
			})
			return
		}
		var filepath string
		var err error
		//根据不同登录状态启用不同的函数
		if !model.CheckSession(c){
			filepath,err=spider(url)
			filepath="/download/?file="+filepath
		}else{
			filepath,err=advancedDownload(url)
		}
		if err!=nil{
			c.JSON(http.StatusOK,gin.H{
				"status":"0",
				"err":err.Error(),
			})
			return
		}
		a[filepath]=time.Now()
		c.JSON(http.StatusOK,gin.H{
			"status":"1",
			"path":filepath,
		})
	})
	//文件下载
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
	})
	//用户注册页面
	router.GET("/hustregister",func(c *gin.Context){
		c.HTML(http.StatusOK,"regist.html",nil)
	})
	//向用户邮箱发送验证码
	router.POST("/hustregister/code", mediumware.LimitTimeMediumware(),func(c *gin.Context) {
		c.JSON(http.StatusOK,gin.H{
			"status":1,
		})
		var user *model.User
		c.ShouldBind(&user)
		//生成一个六位随机数字的验证码
		var code string
		rand.Seed(time.Now().Unix())
		for i := 0; i < 6; i++ {
			code += strconv.Itoa(rand.Intn(10))
		}
		config.VerificationCode[user.EmailAdd] = config.M{Code: code,Time: time.Now()}
		//向用户发送验证码
		utils.SendCode(user.EmailAdd, code)
	})
	//注册api
	router.POST("/hustregister",mediumware.FormatCheck,func(c *gin.Context){
		var user *model.User
		code, ok := c.GetPostForm("code")
		if !ok {
			c.JSON(200, gin.H{
				"status": 0,
				"err":    "验证码不能为空",
			})
			return
		}
		if err := c.ShouldBind(&user); err != nil {
			c.JSON(200, gin.H{
				"status": 0,
				"err":    "表单解析错误！",
			})
			return
		}
		if !user.HaveRegistered() {
			if code == config.VerificationCode[user.EmailAdd].Code {
				if err:=user.AddUser();err!=nil{
					fmt.Println(err)
				}
				c.JSON(200, gin.H{
					"status": 1,
				})
				delete(config.VerificationCode, user.EmailAdd)
			} else {
				c.JSON(200, gin.H{
					"status": 0,
					"err":    "验证码错误!",
				})
			}
		}
	})
	//登录api
	router.POST("/husterlogin", func(c *gin.Context) {
		var user *model.User
		if err:=c.ShouldBind(&user);err!=nil{
			c.JSON(http.StatusOK,gin.H{
				"status":0,
			})
			return
		}
		if p:=user.CheckLogin();p==config.WRONG_PASSWORD{
			c.JSON(http.StatusOK,gin.H{
				"status":0,
				"err":config.WRONG_PASSWORD,
			})
			return
		}else if p==config.NOT_REGISTERED{
			c.JSON(http.StatusOK,gin.H{
				"status":0,
				"err":config.NOT_REGISTERED,
			})
			return
		}
		sessionid:=model.NewSessionID(user.EmailAdd)
		c.SetCookie("sessionid", sessionid, 2592000, "/", config.SeverConfig.DOMAIN, false,true)
		c.JSON(200, gin.H{
			"status": "1",
		})
	})
	//登出
	router.GET("/logout", func(c *gin.Context) {
		c.SetCookie("sessionid", "nil", -1, "/", config.SeverConfig.DOMAIN,false, true)
		c.Redirect(http.StatusFound, "/baiduspider")
	})
	router.Run(config.SeverConfig.LISTEN_ADDRESS+":"+config.SeverConfig.LISTEN_PORT)
}

//未登录用户调用的爬虫函数
func spider(url string)(filepath string,err error) {
	docType,err:=utils.GetDocType(url)
	if err!=nil{
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
	return
}

//登陆用户下载调用的函数
func advancedDownload(urls string)(filepath string,err error){
	infos,ifprofession,err:=utils.GetInfos(urls)
	if err!=nil{
		return
	}
	if ifprofession{
		remain,err:=utils.GetDownloadTicket()
		if err!=nil{
			return "",err
		}
		if remain==0{
			return "",errors.New("无剩余专享文档下载券！")
		}
	}
	client:=&http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse //停止重定向，直接把下载连接发送给用户，节省服务器带宽
		},
	}
	val:=url.Values{
		//"ct": {"20008"},
		"doc_id": {infos[0]},
		//"retType": {"newResponse"}, //用券文档暂时不需要
		//"sns_type": {""},
		"storage": {"1"},
		//"useTicket": {"0"}, //用券文档测试不需要
		//"target_uticket_num": {"0"}, //用券文档暂时不需要
		"downloadToken": {infos[2]},
		//"sz": {"37097"},
		//"v_code": {"0"},
		//"v_input": {"0"},
		"req_vip_free_doc": {"1"}, //用券文档暂时不需要
	}
	req,err:=http.NewRequest("POST","https://wenku.baidu.com/user/submit/download",strings.NewReader(val.Encode()))
	if err!=nil{
		return
	}
	req.Header.Add("User-Agent","Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
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
	return resp.Header.Get("Location"),nil
}

//Timer 定时器，爬虫下载的文件十分钟后删除
func Timer(){
	for{
		time.Sleep(10*time.Minute)
		for key,val:=range a{
			sub:=int(time.Since(val).Minutes())
			if sub>10{
				os.Remove(key)
			}
		}
	}
}

