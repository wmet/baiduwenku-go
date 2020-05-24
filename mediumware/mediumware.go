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
	code, ok := c.GetPostForm("code")
	if !ok {
		c.JSON(200, gin.H{
			"status": 0,
			"err":    "验证码不能为空",
		})
		c.Abort()
		return
	}
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
	if ok, _ := regexp.MatchString("^([a-z0-9_\\.-]+)@([\\da-z\\.-]+)\\.([a-z\\.]{2,6})$", user.EmailAdd); !ok {
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
	if code != config.VerificationCode[user.EmailAdd].Code{
		c.JSON(http.StatusForbidden,gin.H{
			"status":0,
			"err":"验证码不正确",
		})
		c.Abort()
		return
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
		if ok, _ := regexp.MatchString("^([a-z0-9_\\.-]+)@([\\da-z\\.-]+)\\.([a-z\\.]{2,6})$", user.EmailAdd); !ok {
			c.JSON(200, gin.H{
				"status": 0,
				"err":   "请输入正确邮箱格式！",
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
