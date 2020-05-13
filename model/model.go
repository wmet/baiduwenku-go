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
}

type Session struct {
	SessionID string
	Emailadd string
}


//AddUser 添加一个用户
func (user *User) AddUser() error {
	query := "insert into hustusers(emailadd,password) values(?,?)"
	_, err := config.Db.Exec(query, user.EmailAdd, user.Password)
	return err
}

//UpdateUser 修改一个用户的密码
func (user *User) UpdateUser(newpsd string) (err error) {
	query := "update hustusers set password=? where id=? "
	if _, err = config.Db.Exec(query, newpsd, user.ID); err != nil {
		return
	}
	//更改密码并清除保存的session信息
	query = "delete from hustsessions where userid=?"
	_, err = config.Db.Exec(query, user.ID)
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


