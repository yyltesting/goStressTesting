// Package golink 连接
package golink

import (
	"github.com/link1st/go-stress-testing/conf"
	"github.com/link1st/go-stress-testing/model"
	"github.com/link1st/go-stress-testing/server/client"
	"net/http"
	"sync"
)

// HTTP 请求
func HTTP(params *conf.Params, chanID uint64, ch chan<- *model.RequestResults, wg *sync.WaitGroup,variableData *conf.Variable) {
	defer func() {
		wg.Done()
	}()
	clientGo(params,chanID, ch,nil,nil,variableData)
	return
}

// sendList 多个接口分步压测  现在暂时只有一个接口
func sendList(chanID uint64,ch chan<- *model.RequestResults,i uint64, requestList []*model.Request,csvxtoken string,issenders bool) (isSucceed bool, errCode int, requestTime uint64,
	contentLength int64) {
	errCode = model.HTTPOk
	for _, request := range requestList {
		//发送请求
		succeed, code, u, length := send(chanID, request,csvxtoken,issenders)
		isSucceed = succeed
		errCode = code
		requestTime = requestTime + u
		contentLength = contentLength + length
		if succeed == false {
			break
		}
	}
	requestResults := &model.RequestResults{
		Time:          requestTime,
		IsSucceed:     isSucceed,
		ErrCode:       errCode,
		ReceivedBytes: contentLength,
	}
	requestResults.SetID(chanID, i)
	ch <- requestResults
	return
}

// send 发送一次请求
func send(chanID uint64, request *model.Request,csvxtoken string,issenders bool) (bool, int, uint64, int64) {
	var (
		// startTime = time.Now()
		isSucceed     = false
		Code          = model.HTTPOk
		contentLength = int64(0)
		err           error
		resp          *http.Response
		requestTime   uint64
	)
	newRequest := getRequest(request)
	//这里可以进行可变参
	resp, requestTime, err = client.HTTPRequest(chanID, newRequest,csvxtoken,issenders)
	if err != nil {
		Code = model.RequestErr // 如果报错则请求错误
	} else {
		contentLength = resp.ContentLength
		// 验证请求是否成功
		Code, isSucceed = newRequest.GetVerifyHTTP()(newRequest, resp)
	}
	return isSucceed, Code, requestTime, contentLength
}
