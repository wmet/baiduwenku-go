package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gufeijun/baiduwenku/config"
	"github.com/gufeijun/baiduwenku/controller"
	"github.com/gufeijun/baiduwenku/timer"
)

func init() {
	//启用定时任务
	timer.StartTimer()
}

func main() {
	router := gin.Default()
	router.Static("/static", "front-end")        //加载静态文件
	router.LoadHTMLGlob("front-end/html/*.html") //加载网页模板

	//api部分
	router.GET("/baiduspider", controller.GetHomePage)                                        //home页面
	router.GET("/download", controller.HandleDownload)                                        //文件下载
	router.GET("/logout", controller.Logout)                                                  //登出
	router.GET("/hustregister", controller.GetRegisterPage)                                   //用户注册页面
	router.POST("/baiduspider", controller.LogOutput, controller.HandleRequest)               //处理下载请求
	router.POST("/hustregister", controller.FormatCheck, controller.Register)                 //注册
	router.POST("/hustregister/code", controller.LimitTimeMediumware(), controller.HandleMsg) //验证码发送
	router.POST("/husterlogin", controller.Login)                                             //登录

	//启用配置文件中的监听地址与端口
	router.Run(config.SeverConfig.LISTEN_ADDRESS + ":" + config.SeverConfig.LISTEN_PORT)
}
