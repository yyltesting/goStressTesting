package golink

import (
	"github.com/link1st/go-stress-testing/conf"
	"github.com/link1st/go-stress-testing/model"
	"github.com/link1st/go-stress-testing/server/client"
	"time"
)
const (
	firstTime    = 1 * time.Second // 连接以后首次请求数据的时间
	intervalTime = 1 * time.Millisecond // 发送数据的时间间隔
)
//
func clientGo(params *conf.Params,chanID uint64, ch chan<- *model.RequestResults,
	ws *client.WebSocket, gws *client.GrpcSocket,
	variableData *conf.Variable){
	// 暂停1秒,服务端建立完通讯以后有一些异步事件需要处理 避免建立完成以后就发送数据
	t := time.NewTimer(firstTime)
	//以请求时间为条件，在某段时间内已知循环请求 比如100并发持续请求5分钟
	var (
		i uint64
		list []string

	)
	//计算应该结束的时间
	runTime := params.ExecutionTime * 60
	endTime := time.Now().Add(time.Duration(runTime)*time.Second)
	for {
		//select是一种go可以处理多个通道之间的机制
		select {
		case <- params.Ctx.Done():
			goto end
		case <- t.C: //读数据
			t.Reset(intervalTime)
			// 请求
			switch params.Request.Form {
			case model.FormTypeHTTP:
				//获取请求列表
				requestList := getRequestList(params.Request)
				//发送请求
				sendList(chanID,ch,i,requestList,variableData.SendToken,params.Issenders)
			case model.FormTypeWebSocket:
				//一个发送端给多个客户端发送，向每个客户端发送一次
				if params.Toclients{
					for i:=uint64(0);i<params.Concurrency;i++{
						list = conf.CsvList.Recvidlist[i]
						variableData.RecvId = list[0]
						webSocketRequest(params,chanID, ch, i,ws, variableData)
					}
				}else {
					webSocketRequest(params,chanID, ch, i, ws, variableData)
				}
			case model.FormTypeGRPC:
				grpcRequest(chanID, ch, i, params.Request, gws)
			default:
				goto end
			}
			// 结束条件 判断当前时间是否大于结束时间
			if params.ExecutionTime > 0&&time.Now().After(endTime){
				goto end
			}
			if params.ExecutionTime ==0 {
				// 结束条件 判断当前请求是否大于请求数
				i = i + 1
				if i >= params.TotalNumber {
					goto end
				}
			}
		}
	}
	end:
		t.Stop() //停止再次请求通道
	return
}