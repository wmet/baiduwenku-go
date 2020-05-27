package model

import (
	"github.com/gin-gonic/gin"
	"github.com/gufeijun/baiduwenku/config"
	uuid "github.com/satori/go.uuid"
)

type User struct{
	ID             int    `json:"id" form:"id"`
	EmailAdd       string `json:"emailadd" form:"emailadd"`
	Password       string `json:"password" form:"password"`
	PermissionCode int
	Remain int
}

type Session struct {
	SessionID string
	Emailadd string
}

//AddUser 添加一个用户
func (user *User) AddUser() error {
	query := "insert into hustusers(emailadd,password,permissioncode) values(?,?,?)"
	_, err := config.Db.Exec(query, user.EmailAdd, user.Password,user.PermissionCode)
	return err
}

//UpdateUser 修改用户的剩余下载次数
func (user *User) UpdateUser()error {
	query := "update hustusers set remain=? where id=? "
	_, err := config.Db.Exec(query, user.Remain-1, user.ID)
	return err
}

//UpdateAll 更新所有用户的剩余下载次数
func UpdateAll()error{
	query:="update hustusers set remain=3"
	_,err:=config.Db.Exec(query)
	return err
}

//HaveRegistered 判断能否注册
func (user *User) HaveRegistered() bool {
	query := "select id from hustusers where emailadd=?"
	row := config.Db.QueryRow(query, user.EmailAdd)
	var id int
	return row.Scan(&id)==nil
}

//判断一个登录用户是否合法
func (user *User)CheckLogin() int{
	query := "select password from hustusers where emailadd=?"
	row:=config.Db.QueryRow(query,user.EmailAdd)
	var password string
	if err:=row.Scan(&password);err!=nil{
		return config.NOT_REGISTERED
	}
	if password!=user.Password{
		return config.WRONG_PASSWORD
	}
	return config.PERMISSION_PASSWORD
}

//新建一个会话id
func NewSessionID(emailadd string) string {
	u:= uuid.NewV4()
	uuid := u.String()
	//不存在则更改
	if !sessionExisted(emailadd){
		query := "insert into hustsessions(emailadd,sessionid) values(?,?)"
		config.Db.Exec(query, emailadd, uuid)
	} else {
		query := "update hustsessions set sessionid=? where emailadd=?"
		config.Db.Exec(query, uuid, emailadd)
	}
	return uuid
}

func sessionExisted(emailadd string)bool{
	query := "select sessionid from hustsessions where emailadd = ?"
	row := config.Db.QueryRow(query, emailadd)
	var sessionid string
	return row.Scan(&sessionid)==nil
}

//CheckSession 检查服务端是否保存客户端的session信息
func CheckSession(c *gin.Context)bool{
	cookie, err := c.Request.Cookie("sessionid")
	if err != nil {
		return false
	}
	sessionid := cookie.Value
	query := "select userid from hustsessions where sessionid=?"
	row := config.Db.QueryRow(query, sessionid)
	var userid int
	return row.Scan(&userid)!=nil
}

//GetUserInfo  获取用户信息
func GetUserInfo(c *gin.Context)(user *User,err error){
	var u User
	cookie, err := c.Request.Cookie("sessionid")
	if err != nil {
		return
	}
	sessionid:=cookie.Value
	query:="select hustusers.permissioncode,hustusers.remain,hustusers.id from hustusers inner join hustsessions on hustusers.emailadd=hustsessions.emailadd where hustsessions.sessionid=?"
	row:=config.Db.QueryRow(query,sessionid)
	err=row.Scan(&u.PermissionCode,&u.Remain,&u.ID)
	user=&u
	return
}