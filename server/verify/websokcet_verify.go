// Package verify 校验
package verify

import (
	"encoding/json"
	"fmt"
	"github.com/link1st/go-stress-testing/model"
	"github.com/link1st/go-stress-testing/server/receive"
	"log"
)

// WebSocketJSON 通过返回的Body 判断
// 返回示例: {"msgType":2001,"operationId":"","code":0,"message":"","data":"{\"send_id\":\"ub2iuxkqab5k\",\"recv_id\":\"aqm78tq2mpa\",\"server_msg_id\":\"c3ofh5bpgt6\",\"session_type\":1,\"content_type\":101,\"content\":\"I sent a message\",\"seq\":883,\"send_time\":1693879595}"}
// code 取body中的返回code
func WebSocketJSON(request *model.Request, seq string, msg []byte) (code int, isSucceed bool) {
	responseJSON := &receive.WebSocketResponseJSON{}//josn模板
	err := json.Unmarshal(msg, responseJSON) //解析json
	if err != nil{
		code = model.ParseError
		fmt.Printf("请求结果解析json失败 json.Unmarshal msg:%s err:%v", string(msg), err)
	}
	data := responseJSON.Data
	msgdataListJson := &receive.MsgdataListJson{}                      //josn模板
	err = json.Unmarshal([]byte(data), msgdataListJson) //解析json
	if err != nil{
		code = model.ParseError
		fmt.Printf("请求结果解析json失败 json.Unmarshal msg:%s err:%v", string(msg), err)
	}

	c := responseJSON.Code
	// body 中code返回0为返回数据成功
	if c != 0 {
		//fmt.Println(seq)
		code = model.ParseError
	}
	cl := false
	for _, msgData := range msgdataListJson.MsgdataList {
		if msgData.Content == seq{
			cl = true
		}

	}
	if cl{
		code = model.HTTPOk
		isSucceed = true
	}else {
		log.Println("发送端检验seq不一致",string(msg))
		//fmt.Println(string(msg))
		code = model.ParseError
	}


	// 开启调试模式
	if request.GetDebug() {
		fmt.Printf("请求结果 seq:%s body:%s \n", seq, string(msg))
	}
	return
}
