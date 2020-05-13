package mediumware

import (
	"github.com/gin-gonic/gin"
	"github.com/gufeijun/baiduwenku/config"
	"github.com/gufeijun/baiduwenku/model"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

//FormatCheck 传输User数据的格式检验中间件
func FormatCheck(c *gin.Context) {
	var user *model.User
	if err := c.ShouldBind(&user); err != nil {
		c.JSON(200, gin.H{
			"status": 0,
			"err":    err.Error(),
		})
		c.Abort()
		return
	}
	if ok, _ := regexp.MatchString("\\s", user.Password); ok {
		c.JSON(200, gin.H{
			"status": 0,
			"err":    "密码不能有空格",
		})
		c.Abort()
		return
	}
	if ok, _ := regexp.MatchString("^([a-z0-9_\\.-]+)@hust.edu.cn", user.EmailAdd); !ok {
		c.JSON(200, gin.H{
			"status": 0,
			"err":    "邮箱格式有误!",
		})
		c.Abort()
		return
	}
	if len(user.Password) < 8 {
		c.JSON(200, gin.H{
			"status": 0,
			"err":    "密码不少于8位",
		})
		c.Abort()
		return
	}
	if user.HaveRegistered() {
		c.JSON(200, gin.H{
			"status": 0,
			"err":    "邮箱已经被注册",
		})
		c.Abort()
		return
	}
}

//VeryfyMediumware 验证用户是否登录的中间件
func VeryfyMediumware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//解析cookie
		if !model.CheckSession(c) {
			c.String(http.StatusForbidden, "请先登录！")
			c.Abort()
			return
		}
	}
}

//LimitTimeMediumware 限制发送邮箱的次数，防止验证码接口被滥用
func LimitTimeMediumware() gin.HandlerFunc{
	return func(c *gin.Context) {
		var user *model.User
		if err := c.ShouldBind(&user); err != nil {
			c.JSON(200, gin.H{
				"status": 0,
				"err":    err.Error(),
			})
			c.Abort()
			return
		}
		if ok, _ := regexp.MatchString("^([a-z0-9_\\.-]+)@hust.edu.cn", user.EmailAdd); !ok {
			c.JSON(200, gin.H{
				"status": 0,
				"err":    "仅限智慧华中大邮箱注册！",
			})
			c.Abort()
			return
		}
		if user.HaveRegistered(){
			c.JSON(http.StatusOK,gin.H{
				"status":0,
				"err":"邮箱已被注册!",
			})
		}
		m,ok:=config.VerificationCode[user.EmailAdd]
		if !ok{
			c.Next()
			return
		}
		sub:=int(time.Since(m.Time).Seconds())
		if sub<60{
			c.JSON(http.StatusOK,gin.H{
				"status":0,
				"err":"请求太频繁，请于"+strconv.Itoa(60-sub)+"秒后尝试!",
			})
			c.Abort()
			return
		}
	}
}
