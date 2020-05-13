package config

import (
	"database/sql"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

//Db 定义全局db
var (
	Db    *sql.DB
	err   error
	Mutex sync.Mutex
	SeverConfig Config
	VerificationCode map[string]M = make(map[string]M)
)

const(
	NOT_REGISTERED=1
	WRONG_PASSWORD=2
	PERMISSION_PASSWORD=3
)

type M struct{
	Code string
	Time time.Time
}

type Config struct {
	DB_NAME string
	DB_CONN string
	LISTEN_ADDRESS string
	LISTEN_PORT string
	IMAP_PORT int
	IMAP_SERVER string
	IMAP_EMAIL string
	IMAP_PASSWORD string
	BDUSS string
	DOMAIN string
}

func init() {
	f, _ := os.Open("config.json")
	buf, _ := ioutil.ReadAll(f)
	dec := json.NewDecoder(strings.NewReader(string(buf)))
	if err := dec.Decode(&SeverConfig); err != nil {
		log.Fatal("读取配置文件失败")
		os.Exit(1)
	}
	f.Close()
	Db, _ = sql.Open(SeverConfig.DB_NAME, SeverConfig.DB_CONN)
	if err = Db.Ping(); err != nil {
		log.Fatal(err)
	}
}
