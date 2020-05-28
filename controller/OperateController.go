package controller

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gufeijun/baiduwenku/crawl"
	"github.com/gufeijun/baiduwenku/model"
	"github.com/gufeijun/baiduwenku/timer"
	"github.com/gufeijun/baiduwenku/utils"
)

//处理验证码请求
func HandleMsg(c *gin.Context) {
	var (
		user *model.User //用户对象
		code string      //验证码
	)
	c.JSON(http.StatusOK, newSucMsg())
	//读取用户的信息
	c.ShouldBind(&user)

	//生成一个六位随机数字的验证码
	rand.Seed(time.Now().Unix())
	for i := 0; i < 6; i++ {
		code += strconv.Itoa(rand.Intn(10))
	}

	//将验证码信息保存到recorder
	recorder.Add(user.EmailAdd, code)

	//向用户发送验证码
	utils.SendCode(user.EmailAdd, code)
}

//下载请求处理服务器
func HandleRequest(c *gin.Context) {
	var (
		filepath string //最终的文件存储路径
		url      string //用户的请求文档链接
		err      error  //错误信息
		ok       bool
	)
	//读取请求表单中欲下载的百度文档url
	if url, ok = c.GetPostForm("url"); !ok {
		c.JSON(http.StatusOK, newErrMsg(FAILURE_POSTFORM))
		return
	}

	//读取登录状态
	//根据不同登录状态启用不同的下载函数
	switch !model.CheckSession(c) {
	case true:
		filepath, err = normalDownload(url)
		if !strings.Contains(filepath, `https://wkbjcloudbos.bdimg.com`) {
			timer.Timetable[filepath] = time.Now()
			filepath = "/download/?file=" + filepath
		}
	case false:
		user, err1 := model.GetUserInfo(c)
		if err1 != nil {
			c.JSON(http.StatusOK, newErrMsg(err1.Error()))
			return
		}
		filepath, err = advancedDownload(url, user)
	}

	if err != nil {
		c.JSON(http.StatusOK, newErrMsg(err.Error()))
		return
	}
	//向用户发送下载路径
	c.JSON(http.StatusOK, newSucMsg(gin.H{"path": filepath}))
}

//文件下载服务器
func HandleDownload(c *gin.Context) {
	var (
		name     string //文件名
		filesize string //文件的大小
		ok       bool
	)

	//获取用户想下载的文件名
	if name, ok = c.GetQuery("file"); !ok {
		c.String(http.StatusBadRequest, FAILURE_POSTFORM)
		return
	}

	//判断文件是否存在
	fileinfo, err := os.Stat(name)
	if err != nil {
		c.String(http.StatusBadRequest, FAILURE_DOWNLOAD)
		return
	}

	//限制文件大小在50M内
	if fileinfo.Size() > 50<<20 {
		c.String(http.StatusForbidden, LARGE_FILE)
		return
	}

	//防止下载服务器配置文件
	if strings.Contains(name, "config.json") {
		return
	}

	filesize = strconv.FormatInt(fileinfo.Size(), 10)
	c.Writer.WriteHeader(http.StatusOK)
	c.Header("Content-Disposition", "attachment; filename="+name)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", filesize) //设置文件大小
	c.File(name)
}

//未登录用户调用的爬虫函数
func normalDownload(url string) (filepath string, err error) {
	//获取文档格式
	docType, err := utils.GetDocType(url)
	if err != nil {
		return "", errors.New("老夫暂时拿此链接无能为力~（´Д`）")
	}
	switch docType {
	case "txt":
		return crawl.StartTxtSpider(url)
	case "doc":
		return crawl.StartDocSpider(url)
	case "pdf":
		return crawl.StartPdfSpider(url)
	case "ppt":
		return crawl.StartPPTSpider(url)
	default:
		return "", errors.New(fmt.Sprintf("Do Not Support filetype:%s!", docType))
	}
	return
}

//登陆用户下载调用的函数
func advancedDownload(urls string, user *model.User) (filepath string, err error) {
	//如果普通用户没有剩余下载次数
	if user.PermissionCode == 0 && user.Remain == 0 {
		return "", errors.New("今日的三次下载次数用完！")
	}

	//获取文档信息
	infos, ifprofession, err := utils.GetInfos(urls)
	if err != nil {
		return
	}

	//如果当前下载文档为专享文档
	if ifprofession {
		//获取vip专享文档剩余下载次数
		remain, err := utils.GetDownloadTicket()
		if err != nil {
			return "", err
		}

		//普通用户下载vip专享文档受限
		if user.PermissionCode == 0 {
			return "", errors.New("您只有用券文档和共享文档的免费下载权限！")
		}

		if remain == 0 {
			return "", errors.New("无剩余专享文档下载券！")
		}
	}

	//获取文档的真实下载地址
	location, err := utils.Getlocation(infos)
	if err != nil {
		return "", err
	}

	//普通用户下载完成后，今日剩余下载次数减一
	if user.PermissionCode == 0 {
		user.UpdateUser()
	}

	return location, nil
}
