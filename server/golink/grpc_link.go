// Package golink 连接
package golink

import (
	"context"
	"github.com/link1st/go-stress-testing/conf"
	"sync"
	"time"

	"github.com/link1st/go-stress-testing/helper"
	pb "github.com/link1st/go-stress-testing/proto"

	"github.com/link1st/go-stress-testing/model"
	"github.com/link1st/go-stress-testing/server/client"
)

// Grpc grpc 接口请求
func Grpc(params *conf.Params, chanID uint64, ch chan<- *model.RequestResults, wg *sync.WaitGroup, ws *client.GrpcSocket,
	variableData *conf.Variable) {
	defer func() {
		wg.Done()
	}()
	defer func() {
		_ = ws.Close()
	}()
	clientGo(params,chanID, ch,nil,ws, variableData)
	return
}

// grpcRequest 请求
func grpcRequest(chanID uint64, ch chan<- *model.RequestResults, i uint64, request *model.Request,
	ws *client.GrpcSocket) {
	var (
		startTime = time.Now()
		isSucceed = false
		errCode   = model.HTTPOk
	)
	// 需要发送的数据
	conn := ws.GetConn()
	if conn == nil {
		errCode = model.RequestErr
	} else {
		// TODO::请求接口示例
		c := pb.NewApiServerClient(conn)
		var (
			ctx = context.Background()
			req = &pb.Request{
				UserName: request.Body,
			}
		)
		rsp, err := c.HelloWorld(ctx, req)
		// fmt.Printf("rsp:%+v", rsp)
		if err != nil {
			errCode = model.RequestErr
		} else {
			// 200 为成功
			if rsp.Code != 200 {
				errCode = model.RequestErr
			} else {
				isSucceed = true
			}
		}
	}
	requestTime := uint64(helper.DiffNano(startTime))
	requestResults := &model.RequestResults{
		Time:      requestTime,
		IsSucceed: isSucceed,
		ErrCode:   errCode,
	}
	requestResults.SetID(chanID, i)
	ch <- requestResults
}
