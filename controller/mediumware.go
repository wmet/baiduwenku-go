package controller

import (
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gufeijun/baiduwenku/model"
)

//FormatCheck 检测用户的注册表单是否合法
func FormatCheck(c *gin.Context) {
	var (
		user *model.User //用户信息
		code string      //验证码
		ok   bool
	)

	//读取出提交表单中的验证码
	code, ok = c.GetPostForm("code")
	if !ok {
		c.AbortWithStatusJSON(http.StatusOK, newErrMsg(EMPTY_CODE))
	}

	//读取用户提交的信息
	if err := c.ShouldBind(&user); err != nil {
		c.AbortWithStatusJSON(http.StatusOK, newErrMsg(err.Error()))
	}

	//检验用户设置密码的合法性
	if ok, _ = regexp.MatchString("\\s", user.Password); ok || len(user.Password) < 8 {
		c.AbortWithStatusJSON(http.StatusOK, newErrMsg(ILLEGAL_PASSWORD))
	}

	//检验邮箱格式的正确性
	if ok, _ = regexp.MatchString("^([a-z0-9_\\.-]+)@([\\da-z\\.-]+)\\.([a-z\\.]{2,6})$", user.EmailAdd); !ok {
		c.AbortWithStatusJSON(http.StatusOK, newErrMsg(ERR_FORMATION))
	}

	//检验邮箱是否已经注册
	if user.HaveRegistered() {
		c.AbortWithStatusJSON(http.StatusOK, newErrMsg(HAVE_REGISTERED))
	}

	//判断用户提交的验证码是否正确
	if code != recorder.GetCode(user.EmailAdd) || code == "" {
		c.AbortWithStatusJSON(http.StatusOK, newErrMsg(WRONG_CODE))
	}
}

//LimitTimeMediumware 限制发送邮箱的次数，防止验证码接口被滥用
func LimitTimeMediumware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user *model.User
		//解析用户
		if err := c.ShouldBind(&user); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, newErrMsg(err.Error()))
		}

		//邮箱格式检验
		if ok, _ := regexp.MatchString("^([a-z0-9_\\.-]+)@([\\da-z\\.-]+)\\.([a-z\\.]{2,6})$", user.EmailAdd); !ok {
			c.AbortWithStatusJSON(http.StatusOK, newErrMsg(ERR_FORMATION))
		}

		//检验是否已经注册
		if user.HaveRegistered() {
			c.AbortWithStatusJSON(http.StatusOK, newErrMsg(HAVE_REGISTERED))
		}

		//获取上次向该邮箱发送的时间
		t, ok := recorder.GetTime(user.EmailAdd)
		if !ok {
			c.Next()
			return
		}

		//计算距上次发送经过的时间,小于60秒则返回错误信息
		sub := int(time.Since(t).Seconds())
		if sub < 60 {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"status": 0, "err": "请求太频繁，请于" + strconv.Itoa(60-sub) + "秒后尝试!"})
		}
	}
}

//日志记录使用情况
func LogOutput(c *gin.Context) {
	go func(remote string) {
		f, _ := os.OpenFile("log.output", os.O_CREATE|os.O_APPEND, 0666)
		defer f.Close()
		log.SetOutput(f)
		log.SetPrefix("[用户]")
		log.Printf("IP:  %s", remote)
	}(c.Request.RemoteAddr)
}
