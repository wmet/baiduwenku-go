package timer

import (
	"log"
	"os"
	"time"

	"github.com/gufeijun/baiduwenku/model"
)

/*
	定时任务
*/

//存放下载的文档以及下载的时间
var Timetable map[string]time.Time = make(map[string]time.Time)

func StartTimer() {
	go timer1()
	go timer2()
}

//timer1 定时器，爬虫下载的文件120分钟后删除后删除，精度不高，最大有60分钟偏差
func timer1() {
	for {
		time.Sleep(60 * time.Minute)
		for key, val := range Timetable {
			sub := int(time.Since(val).Minutes())
			if sub > 120 {
				os.Remove(key)
			}
		}
	}
}

//Timer2 定时器，每天凌晨12点重置用户的剩余下载次数
func timer2() {
	for {
		now := time.Now()
		next := now.Add(time.Hour * 24)
		next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
		t := time.NewTimer(next.Sub(now))
		<-t.C
		if err := model.UpdateAll(); err != nil {
			log.Println(err)
		}
	}
}
