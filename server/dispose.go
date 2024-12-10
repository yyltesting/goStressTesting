// Package server 压测启动
package server

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/link1st/go-stress-testing/conf"
	"github.com/link1st/go-stress-testing/server/receive"
	"sync"
	"time"

	"github.com/link1st/go-stress-testing/model"
	"github.com/link1st/go-stress-testing/server/client"
	"github.com/link1st/go-stress-testing/server/golink"
	"github.com/link1st/go-stress-testing/server/statistics"
	"github.com/link1st/go-stress-testing/server/verify"
)


// init 注册验证器 初始化方法
func init() {

	// http
	model.RegisterVerifyHTTP("statusCode", verify.HTTPStatusCode)
	model.RegisterVerifyHTTP("json", verify.HTTPJson)
	// webSocket
	model.RegisterVerifyWebSocket("json", verify.WebSocketJSON)
}

var (
	wg          sync.WaitGroup // 发送数据完成，监听发送数据是否完成用
	wgReceiving sync.WaitGroup // 数据处理完成，监听报告用 为了锁住当前报告打印
)
// Dispose 处理函数
func Dispose(params *conf.Params) {
	// 设置接收数据缓存，chan函数设置缓存通道
	ch := make(chan *model.RequestResults, params.Concurrency)
	//设置websocket长连接通道，全部结束后断开
	chWebsocket := make(chan int,0)

	//建立发送端和客户端连接
	if params.Request.Form == model.FormTypeWebSocket{
		client.CreatWebsocketClients(params)
	}

	wgReceiving.Add(1)
	//统计压测结果执行上下文
	go statistics.ReceivingResults(params,ch, &wgReceiving,chWebsocket)

	//计算消息发起最早的时间
	receive.Startsend_time = time.Now().Add(time.Second).UnixNano()
	//阶梯压测 启动延时和循环 每次循环后延时，每次并发数=设置的启动数 循环次数=总并发数/启动数
	if params.Interval>0&&params.Startsnum>0{
		loopnum := params.Concurrency/params.Startsnum
		//通过并发数建立协程
		for j := uint64(1);j<=loopnum;j++{
			for i := uint64(0); i < params.Startsnum*j; i++ {
				ReqGo(i,ch,&wg,chWebsocket,params)
			}
			//延时启动
			time.Sleep(time.Duration(params.Interval) * time.Second)
		}
	}else {
		//通过并发数建立协程
		for i := uint64(0); i < params.Concurrency; i++ {
			ReqGo(i,ch,&wg,chWebsocket,params)
		}
	}

	// 等待所有的数据都发送完成
	wg.Wait()
	receive.WgMessage.Wait()
	// 延时1毫秒 确保数据都处理完成了
	time.Sleep(1 * time.Millisecond)
	//关闭通道
	close(ch)
	// 数据全部处理完成了
	wgReceiving.Wait()
	return
}

func ReqGo(i uint64,ch chan<- *model.RequestResults, wg *sync.WaitGroup,chWebsocket chan int,params *conf.Params)  {
	wg.Add(1)
	var (
		sendToken string
		sendId string
		recvxToken string
		recvId string
	)

	sendToken = conf.GetValueFromCsvOrDefault(conf.CsvList.SendToken, conf.CsvList.Tokenlist, i)
	sendId =  conf.GetValueFromCsvOrDefault(conf.CsvList.SendId, conf.CsvList.Sendidlist, i)
	recvxToken =  conf.GetValueFromCsvOrDefault(conf.CsvList.RecvxToken, conf.CsvList.Recvtokenlist, i)
	recvId =  conf.GetValueFromCsvOrDefault(conf.CsvList.RecvId, conf.CsvList.Recvidlist, i)

	variableData := &conf.Variable{
		SendToken : sendToken,
		SendId :sendId,
		RecvxToken :recvxToken,
		RecvId :recvId,
	}

	switch params.Request.Form {
	case model.FormTypeHTTP:
		go golink.HTTP(params, i, ch, wg,variableData) //i当前并发数索引、ch当前通道、totalNumber请求次数、wg等待组地址、request请求
	case model.FormTypeWebSocket:
		switch params.ConnectionMode {
		case 1:
			var newws *client.WebSocket
			var err error
			// 连接以后再启动协程
			if !params.Issenders{
				//如果未设置变量使用url参数
				ws := client.NewWebSocket(params.Request.URL)
				ws.SetHeader(params.Request.Headers)
				if len(sendToken)==0{
					newws,err = ws.GetConn("")
				}else {
					newws,err = ws.GetConn(sendToken)
				}
				if err != nil {
					fmt.Println("后连接失败:", i, err)
					wg.Done()
					return
				}
			}else {
				//是多个发送端证明已经建立了连接 取出conn
				conn,ok := receive.GetClientMap().Load(sendToken)
				if !ok {
					fmt.Println("未找到客户端")
					wg.Done()
					return
				}
				ws := conn.(*websocket.Conn)
				newws = client.CreatWebsocket(ws)
			}
			//i当前并发数索引、ch当前通道、totalNumber请求次数、wg组地址、request请求、ws连接地址
			go golink.WebSocket(params, i, ch, wg, newws,chWebsocket,variableData)
		case 2:
			// 并发建立长链接
			go func(params *conf.Params, i uint64, ch chan<- *model.RequestResults, wg *sync.WaitGroup, chWebsocket chan int,
				variableData *conf.Variable) {
				// 连接以后再启动协程
				ws := client.NewWebSocket(params.Request.URL) //解析WSURL，创建数据结构实例
				//ws.SetHeader(params.Request.Headers)
				newWs,err := ws.GetConn(variableData.SendToken) //建立连接
				if err != nil {
					fmt.Println("连接失败:", i, err)
					wg.Done()
					return
				}
				//发送数据校验数据
				golink.WebSocket(params, i, ch, wg, newWs,chWebsocket,variableData)

			}(params, i, ch, wg, chWebsocket,variableData)
			// 注意:时间间隔太短会出现连接失败的报错 默认连接时长:20毫秒(公网连接)
			time.Sleep(20 * time.Millisecond)
		default:
			data := fmt.Sprintf("不支持的类型:%d", params.ConnectionMode)
			panic(data)
		}
	case model.FormTypeGRPC:
		// 连接以后再启动协程
		ws := client.NewGrpcSocket(params.Request.URL)
		err := ws.Link()
		if err != nil {
			fmt.Println("连接失败:", i, err)
			wg.Done()
			return
		}
		go golink.Grpc(params, i, ch,  wg, ws,variableData)
	default:
		// 类型不支持
		wg.Done()
	}
}
