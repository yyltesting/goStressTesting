// Package client http 客户端
package client

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/link1st/go-stress-testing/model"
	httplongclinet "github.com/link1st/go-stress-testing/server/client/http_longclinet"
	"golang.org/x/net/http2"

	"github.com/link1st/go-stress-testing/helper"
)

// logErr err
var logErr = log.New(os.Stderr, "", 0)
var endtime uint64

// HTTPRequest HTTP 请求
// method 方法 GET POST
// url 请求的url
// body 请求的body
// headers 请求头信息
// timeout 请求超时时间
func HTTPRequest(chanID uint64, request *model.Request,csvxtoken string ,issenders bool) (resp *http.Response, requestTime uint64, err error) {
	method := request.Method
	url := request.URL
	body := request.GetBody()
	timeout := request.Timeout
	//headers := request.Headers
	headers := make(map[string]string)

	for key, value := range request.Headers {
		headers[key]=value
	}

	//请求数据组装
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return
	}

	// 在req中设置Host，解决在header中设置Host不生效问题
	if _, ok := headers["Host"]; ok {
		req.Host = headers["Host"]
	}
	// 设置默认为utf-8编码
	//if _, ok := headers["Content-Type"]; !ok {
	//	if headers == nil {
	//		headers = make(map[string]string)
	//	}
	//	headers["Content-Type"] = "application/x-www-form-urlencoded; charset=utf-8"
	//}
	//CSV参数替换,多客户端请求token
	if issenders{
		headers["x-token"] = csvxtoken
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	var client *http.Client
	if request.Keepalive {
		client = httplongclinet.NewClient(chanID, request)
		startTime := time.Now()
		//获取返回数据
		//fmt.Println(req)
		resp, err = client.Do(req)
		//请求时间截取
		requestTime = uint64(helper.DiffNano(startTime))
		if err != nil {
			//fmt.Println(err)
			logErr.Println("请求失败:", err)

			return
		}
		//fmt.Println("resp",resp)
		return
	} else {
		req.Close = true
		tr := &http.Transport{}
		if request.HTTP2 {
			// 使用真实证书 验证证书 模拟真实请求
			tr = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
			}
			if err = http2.ConfigureTransport(tr); err != nil {
				return
			}
		} else {
			// 跳过证书验证
			tr = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
		}

		client = &http.Client{
			Transport: tr,
			Timeout:   timeout,
		}
		if !request.Redirect {
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}
		}
		startTime := time.Now()
		resp, err = client.Do(req)
		requestTime = uint64(helper.DiffNano(startTime))
		if err != nil {
			logErr.Println("请求失败:", err)

			return
		}
		return
	}
}
