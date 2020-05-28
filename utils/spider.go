package utils

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/gufeijun/baiduwenku/config"
)

/*
	爬虫相关的函数
*/

//爬虫的封装
func QuickSpider(url string) (string, error) {
	cli := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36")
	resp, err := cli.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

//下载图片
func GetJPG(url string) ([]byte, error) {
	cli := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36")
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

//获取重定向后的真实下载链接
func Getlocation(infos []string) (location string, err error) {
	//建立http客户端
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse //停止重定向，直接把下载连接发送给用户，节省服务器带宽
		},
	}

	//设置post表单数据
	val := url.Values{
		"doc_id":           {infos[0]},
		"storage":          {"1"},
		"downloadToken":    {infos[2]},
		"req_vip_free_doc": {"1"}, //共享文档应设为0
	}

	req, err := http.NewRequest("POST", "https://wenku.baidu.com/user/submit/download", strings.NewReader(val.Encode()))
	if err != nil {
		return
	}

	//设置请求头
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	//设置cookie
	cookie := &http.Cookie{
		Name:  "BDUSS",
		Value: config.SeverConfig.BDUSS,
	}
	req.AddCookie(cookie)

	//发起请求
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
	location = resp.Header.Get("Location")

	//如果获取重定向地址失败，尝试将"req_vip_free_doc"参数改为1重新请求
	if location == "" {
		val.Set("req_vip_free_doc", "0")
		//更改请求体
		req.Body = ioutil.NopCloser(strings.NewReader(val.Encode()))
		resp, err = client.Do(req)
		if err != nil {
			return
		}
		resp.Body.Close()

		location = resp.Header.Get("Location")
		if location == "" {
			return "", errors.New("下载出错，未知原因！")
		}
	}
	return
}
