package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/gufeijun/baiduwenku/config"
	"github.com/gufeijun/baiduwenku/utils"
	"net/http"
	"strings"
)

//获取注册界面
func GetRegisterPage(c *gin.Context){
	c.HTML(http.StatusOK,"regist.html",nil)
}

//获取home页面
func GetHomePage(c *gin.Context){
	var emailadd string

	//从用户的请求体中读出cookie
	//如果能读出cookie则把用户名称返回给前端
	cookie, _ := c.Request.Cookie("sessionid")
	if cookie!=nil{
		sessionid := cookie.Value
		query := "select emailadd from hustsessions where sessionid=?"
		row := config.Db.QueryRow(query, sessionid)
		row.Scan(&emailadd)
		emailadd=strings.Split(emailadd,"@")[0]
	}

	///获取剩余的专享vip下载券，显示给前端
	remain,_:=utils.GetDownloadTicket()

	//模板引擎渲染
	c.HTML(http.StatusOK,"home.html",struct {
		Emailadd string
		Remain int
	}{
		emailadd,
		remain,
	})
}