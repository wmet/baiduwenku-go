package config

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var (
	Db          *sql.DB
	SeverConfig Config
)

//配置信息
type Config struct {
	DB_NAME        string //数据库名称
	DB_CONN        string //数据库连接
	LISTEN_ADDRESS string //监听地址
	LISTEN_PORT    string //监听端口
	IMAP_PORT      int    //IMAP服务器的端口
	IMAP_SERVER    string //IMAP服务器
	IMAP_EMAIL     string //IMAP服务器的邮箱
	IMAP_PASSWORD  string //IMAP服务的授权码
	BDUSS          string //百度文库vip账号的cookie
	DOMAIN         string //自己服务器的域名
}

func init() {
	f, err := os.Open("config.json")
	if err != nil {
		panic("无法定位配置文件")
	}
	defer f.Close()

	buf, _ := ioutil.ReadAll(f)
	//解码配置文件
	dec := json.NewDecoder(strings.NewReader(string(buf)))
	if err := dec.Decode(&SeverConfig); err != nil {
		panic("读取配置文件失败")
	}

	//连接数据库
	Db, _ = sql.Open(SeverConfig.DB_NAME, SeverConfig.DB_CONN)
	if err = Db.Ping(); err != nil {
		panic(err)
	}
}
