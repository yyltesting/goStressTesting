// Package golink 连接
package golink

import (
	"encoding/json"
	"fmt"
	"github.com/link1st/go-stress-testing/conf"
	"github.com/link1st/go-stress-testing/server/receive"
	"github.com/link1st/go-stress-testing/server/statistics"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/link1st/go-stress-testing/helper"
	"github.com/link1st/go-stress-testing/model"
	"github.com/link1st/go-stress-testing/server/client"
)


var (
	countmessage uint64 //当前已发送条数
)


// WebSocket webSocket go link 发送数据逻辑
func WebSocket(params *conf.Params, chanID uint64, ch chan<- *model.RequestResults,
	wg *sync.WaitGroup,ws *client.WebSocket, chWebsocket chan int, variableData *conf.Variable) {
	defer func() {
		wg.Done() //组完成
	}()
	defer func() {
		_ = ws.Close() //关闭连接
		//client.StopPing(csvxtoken)
	}()

	defer func() {
		if err := recover(); err != nil {
			log.Println("出现panic", err)
		}
	}()

	clientGo(params,chanID, ch,ws,nil, variableData)
	atomic.AddInt64(&statistics.Reqnum,1)
	// 保持连接 等待接收端全部接收完了再关闭
	if params.Request.Keepalive {
		// 保持连接
		//chWaitFor := make(chan int, 0) //无缓冲的通道阻塞的方式保持连接
		//chWaitFor <- 1
		select {
		case <-params.Ctx.Done():
		case chWebsocket <- 1:
		}
	}
	return
}

// webSocketRequest 请求
func webSocketRequest(params *conf.Params,chanID uint64, ch chan<- *model.RequestResults, i uint64,
	ws *client.WebSocket, variableData *conf.Variable) {
	//可能ping和写入时并发报错 重新执行该方法
	defer func() {
		if r := recover(); r != nil {
			// 在发生 panic 时执行特定的处理
			fmt.Println("发生了并发 panic:", r)
			time.Sleep(time.Millisecond*10)
			webSocketRequest(params,chanID, ch, i, ws, variableData)
		}
	}()
	var (
		startTime   = time.Now()
		isSucceed   = false
		Code        = model.HTTPOk
		msg         []byte
		requestTime uint64
	)
	//fmt.Println(csvxtoken,csvsendid,csvrecvid)
	// 需要发送的数据,此处seq需要组装发送的消息内容 唯一
	seq := fmt.Sprintf("%d_%s_%s_%d", startTime.UnixNano(),variableData.SendId,variableData.RecvId,chanID)
	operationId := fmt.Sprintf("%d_%d_%s", startTime.UnixNano(),chanID,variableData.RecvId)
	clientMsgId := fmt.Sprintf("%d_%d_%s", startTime.UnixNano(),chanID,variableData.RecvId)
	msgStr := `
	{
	 "msgType": 1003,
	 "operationId": "`+operationId+`",
	 "data": "{\"sid\":\"` + variableData.SendId + `\",\"gid\":\"` + variableData.RecvId + `\",\"sty\":2,\"cty\":101,\"cmid\":\"`+clientMsgId+`\",\"c\":\"` + seq + `\"}"
	}`
	//err := ws.Write([]byte(`
	//{
	//  "msgType": 1003,
	//  "token": "` + csvxtoken + `",
	//  "sendId": "` + csvsendid + `",
	//  "operationId": "`+operationId+`",
	//  "data": "{\"sendId\":\"` + csvsendid + `\",\"recvId\":\"` + csvrecvid + `\",\"sessionType\":1,\"contentType\":102,\"clientMsgId\":\"`+clientMsgId+`\",\"content\":\"{\\\"oFileId\\\":\\\"cdzj9tjihlh\\\",\\\"localId\\\":\\\"661\\\",\\\"width\\\":3200,\\\"height\\\":1440}\"}"
	//}`)) //发送媒体图片
	//err := error(nil)
	err := ws.Write([]byte(msgStr),params.Timeout) //发送
	if err != nil {
		log.Println("请求错误",err)
		Code = model.RequestErr   // 请求错误
		requestTime = 0 	//错误的时间不统计
	} else {
		//如果开启debug打印发送的消息
		if params.Request.GetDebug(){
			// 定义格式化布局
			lay := "2006-01-02 15:04:05"
			// 将当前时间格式化为指定布局
			reqtime := time.Now().Format(lay)
			fmt.Println(reqtime+msgStr)
		}
		//如果设定一个客户端接收数据，数据会很大解析数据很慢影响工具真实统计，所以不作为参考，在利用协程统计
		if params.Iscollectmsg{
			msg, err = ws.Read(params.Timeout) //发送者接收服务器的消息
			//msg, err = ws.Read() //发送者接收服务器的消息
			//fmt.Println("send------",string(msg))
			atomic.AddUint64(&countmessage,1)
			requestTime = uint64(helper.DiffNano(startTime)) //纳秒转换
			if err != nil {
				log.Println("发送端接收消息错误",err)
				Code, isSucceed =  model.ParseError,false
			}else {
				//发送消息的最后一条相关信息
				if atomic.LoadUint64(&countmessage)==params.Concurrency*params.TotalNumber&& params.ExecutionTime==0&&params.Startsnum==0{
					// 获取当前时间
					currentTime := time.Now()
					// 定义格式化布局
					layout := "2006-01-02 15:04:05"
					// 将当前时间格式化为指定布局
					statistics.Lastsendmessagetime = currentTime.Format(layout)
					statistics.Lastsendmessage = msgStr
				}
				//Code, isSucceed = request.GetVerifyWebSocket()(request, seq, msg) //校验数据
				responseJSON := &receive.WebSocketResponseJSON{} //josn模板
				err := json.Unmarshal(msg, responseJSON) //解析json
				if err != nil{
					fmt.Println("无法解析json",err,string(msg))
					Code, isSucceed =  model.ParseError,false
				}
				if responseJSON.Code == 0 {
					Code, isSucceed =  model.HTTPOk,true
				}else {
					fmt.Println("send-responseJSON--",responseJSON)
					Code, isSucceed =  model.ParseError,false
				}
			}
		}else if params.Clients == 0{
			//只管发，也没有客户端统计
			msg, err = ws.Read(params.Timeout) //发送者接收服务器的消息
			//msg, err = ws.Read() //发送者接收服务器的消息
			//fmt.Println(string(msg))
			requestTime = uint64(helper.DiffNano(startTime)) //纳秒转换
			if err != nil {
				log.Println("发送端接收消息错误",err)
				Code, isSucceed =  model.ParseError,false
			}else {
				Code, isSucceed = params.Request.GetVerifyWebSocket()(params.Request, seq, msg) //校验数据
				//Code, isSucceed =  model.HTTPOk,true
			}
		}else {
				//接收端只可能收到一条消息
				_,msg,err = receive.SingleMessage(variableData.RecvxToken) //取多个客户端的消息
				//这里的时间需要使用接收消息的时间减发送消息的时间
				requestTime = uint64(helper.DiffNano(startTime)) //纳秒转换
				if err != nil {
					log.Println("发送端接收消息错误",err)
					Code, isSucceed =  model.ParseError,false
				}else {
					Code, isSucceed = params.Request.GetVerifyWebSocket()(params.Request, seq, msg) //校验数据
				}
			}
		}

	//失败不统计时间
	if err != nil{
		requestTime = 0
	}
	//计算接收消息的最长时间
	if statistics.ReqTime == 0 {
		statistics.ReqTime  = requestTime
		statistics.Maxtimemsg = msg
	}else if requestTime > statistics.ReqTime  && err == nil{
		statistics.ReqTime  = requestTime
		statistics.Maxtimemsg = msg
	}

	requestResults := &model.RequestResults{
		Time:      requestTime,
		IsSucceed: isSucceed,
		ErrCode:   Code,
	}
	requestResults.SetID(chanID, i)
	ch <- requestResults
}
