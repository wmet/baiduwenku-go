package utils

import (
	"archive/zip"
	"errors"
	"io"
	"log"
	"os"
	"regexp"

	"github.com/go-gomail/gomail"
	"github.com/gufeijun/baiduwenku/config"
)

/*
	公用的工具函数
*/

//正则的封装
func QuickRegexp(raw string, patten string) ([][]string, error) {
	reg := regexp.MustCompile(patten)
	res := reg.FindAllStringSubmatch(raw, -1)
	if len(res) == 0 {
		return nil, errors.New("No Submatch")
	}
	return res, nil
}

//打包文件
func ZipFiles(filename string, files []string) error {
	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()
	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()
	for _, file := range files {
		if err = addFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

func addFileToZip(zipWriter *zip.Writer, filename string) error {
	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = filename
	header.Method = zip.Deflate
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

//SendCode 发送验证码
func SendCode(emailadd string, code string) {
	m := gomail.NewMessage()

	m.SetAddressHeader("From", config.SeverConfig.IMAP_EMAIL /*"发件人地址"*/, "百度文库下载平台") // 发件人

	m.SetHeader("To", m.FormatAddress(emailadd, "收件人")) // 收件人

	m.SetHeader("Subject", "来自百度文库下载平台的验证码") // 主题

	m.SetBody("text/plain", "您的验证码为:"+code) // 正文

	d := gomail.NewPlainDialer(config.SeverConfig.IMAP_SERVER, config.SeverConfig.IMAP_PORT, config.SeverConfig.IMAP_EMAIL, config.SeverConfig.IMAP_PASSWORD) // 发送邮件服务器、端口、发件人账号、发件人密码
	if err := d.DialAndSend(m); err != nil {
		log.Println("发送失败", err)
		return
	}
	log.Println("Have sent email to " + emailadd)
}

func PrePrecess(url string) (location string, ok bool) {
	ok, err := IsVIPfreeDoc(url)
	if err != nil || !ok {
		return
	}
	infos, _, err := GetInfos(url)
	if err != nil {
		return
	}
	location, err = Getlocation(infos)
	return location, err == nil
}
