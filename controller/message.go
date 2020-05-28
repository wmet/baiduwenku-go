package controller

import (
	"github.com/gin-gonic/gin"
	"time"
)

type m struct {
	Code string			//验证码
	Time time.Time		//时间
}

type MessageRecorder map[string]m

//建立一个全局的recorder，记录验证码信息
var recorder MessageRecorder=make(map[string]m)

const(
	FAILURE_POSTFORM = "错误表单数据！"
	FAILURE_DOWNLOAD = "无此文件！"
	LARGE_FILE		 = "目标文件大于50M，禁止下载！"
	HAVE_REGISTERED  = "邮箱已经被注册！"
	ERR_FORMATION	 = "邮箱格式不正确！"
	EMPTY_CODE		 = "空验证码！"
	ILLEGAL_PASSWORD = "密码不小于8位且不能有空格!"
	WRONG_CODE		 = "验证码错误!"
)

//记录验证码信息
func (this MessageRecorder) Add(email string,code string){
	this[email]=m{
		Code: code,
		Time: time.Now(),
	}
}

//获取上一次邮件时间
func (this MessageRecorder)GetTime(email string) (time.Time,bool){
	m,ok:=recorder[email]
	return m.Time,ok
}

//获取验证码
func (this MessageRecorder)GetCode(email string)string{
	m:=recorder[email]
	return m.Code
}

//删除
func (this MessageRecorder)Delete(email string){
	delete(this,email)
}

//生成错误讯息
func newErrMsg(errmsg string)gin.H{
	return gin.H{
		"status":0,
		"err":errmsg,
	}
}

//生成成功讯息
func newSucMsg(h ...gin.H)gin.H{
	if h==nil{
		return gin.H{
			"status":1,
		}
	}
	h[0]["status"]=1
	return h[0]
}