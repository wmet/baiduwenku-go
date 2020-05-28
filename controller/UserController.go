package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/gufeijun/baiduwenku/config"
	"github.com/gufeijun/baiduwenku/model"
	"net/http"
	"strings"
)

//处理用户的注册表单
func Register(c *gin.Context){
	var user *model.User

	//解析用户的提交信息
	if err := c.ShouldBind(&user); err != nil {
		c.JSON(http.StatusOK,newErrMsg(FAILURE_POSTFORM))
		return
	}

	//查看是否是华科邮箱，进而赋予不同的权限
	//普通人权限为0，huster为1
	emailTail:=strings.Split(user.EmailAdd,"@")[1]
	if emailTail=="hust.edu.cn"{
		user.PermissionCode=1
	}

	//将注册信息写入数据库
	if err:=user.AddUser();err!=nil{
		c.JSON(http.StatusBadRequest,newErrMsg(err.Error()))
		return
	}

	//像用户发送成注册功信息
	c.JSON(http.StatusOK, newSucMsg())

	//删除验证码表中的记录
	recorder.Delete(user.EmailAdd)
}

//登录
func Login(c *gin.Context){
	var user *model.User

	//解析用户的提交表单
	if err:=c.ShouldBind(&user);err!=nil{
		c.JSON(http.StatusOK,newErrMsg(err.Error()))
		return
	}

	//检验登录是否成功
	if p:=user.CheckLogin();p!=""{
		c.JSON(http.StatusOK,newErrMsg(p))
		return
	}

	//给该用户分配一个session
	sessionid:=model.NewSessionID(user.EmailAdd)
	c.SetCookie("sessionid", sessionid, 2592000, "/", config.SeverConfig.DOMAIN, false,true)

	//发送成功登录信息
	c.JSON(200, newSucMsg())
}

//登出
func Logout(c *gin.Context){
	c.SetCookie("sessionid", "nil", -1, "/", config.SeverConfig.DOMAIN,false, true)
	c.Redirect(http.StatusFound, "/baiduspider")
}