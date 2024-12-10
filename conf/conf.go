package conf

import (
	"errors"
	"fmt"
	"github.com/link1st/go-stress-testing/helper"
	"github.com/link1st/go-stress-testing/model"
	"strings"
)
import "context"

type	Params	struct{
	Concurrency uint64       // 并发数
	TotalNumber uint64      // 请求数(单个并发/协程)
	Request		*model.Request
	Interval	uint64	//阶梯压测时间启动间隔S
	Startsnum	uint64	//阶梯压测每秒启动数，注意需要和并发数整除
	Clients		int64	//是否多个客户端,1一个、2多个、0不设置则为单接口
	Issenders   bool	//是否多个发送端
	ExecutionTime	uint64  //请求时间min
	Ctx 		context.Context
	Iscollectmsg		bool //是否需要将消息收集起来统计
	Toclients		bool //一个发送端是否向多个客户端发送
	Timeout  uint64 //接收超时时间
	ConnectionMode int64	// 1:顺序建立长链接 2:并发建立长链接
	Variable	string	//变量通过,号分割 分别为sendtoken,sendid,recvtoken,recvid
}

//CSV全局变量
type Csv struct {
	Tokenlist   [][]string
	Sendidlist  [][]string
	Recvtokenlist	[][]string
	Recvidlist	[][]string
	SendToken string
	SendId string
	RecvId string
	RecvxToken string
}
//可变变量
type Variable struct {
	SendToken string
	SendId string
	RecvxToken string
	RecvId string
}
var(
	tokenlist   [][]string
	sendidlist  [][]string
	recvtokenlist	[][]string
	recvidlist	[][]string
	sendToken string
	sendId string
	recvxToken string
	recvId string
	CsvList Csv
)
func InitVariable(params *Params)(err error) {
	data := strings.Split(params.Variable,",")
	if len(data) != 4 {
		//采用定制变量必须有四个参数 可以为空 aid,,,
		return nil
	}
	sendToken = data[0]
	sendId =data[1]
	recvxToken =data[2]
	recvId =data[3]

	//是否开启CSV文件读取参数
	fmt.Println("正在读取参数csv文件.......")
	if params.Issenders{
		if strings.Contains(sendToken, "csv"){
			//token
			tokenlist = helper.OpenCsv("./csv/"+sendToken) //发送者token/http请求token
		}else {
			return errors.New("多个发送端请配置csv参数")
		}
		//发送者id
		if params.Request.Form == model.FormTypeWebSocket && strings.Contains(sendId, "csv"){
			sendidlist = helper.OpenCsv("./csv/"+sendId) //发送者id
		}
	}
	if params.Request.Form == model.FormTypeWebSocket{
		//可通过DescOpenCsv方法倒叙读取
		//高并发id user_info.csv
		//recvId = "dqh8t95lh3o"//"dxma8dmchfr"//性能测试群
		if strings.Contains(recvxToken, "csv"){
			recvtokenlist = helper.OpenCsv("./csv/"+recvxToken) //接收者token
		}
		if strings.Contains(recvId, "csv"){
			recvidlist = helper.OpenCsv("./csv/"+recvId) //接收者id
		}
	}

	CsvList = Csv{
		Tokenlist   :tokenlist,
		Sendidlist  :sendidlist,
		Recvtokenlist	:recvtokenlist,
		Recvidlist	:recvidlist,
		SendToken :sendToken,
		SendId :sendId,
		RecvxToken :recvxToken,
		RecvId :recvId,
	}

	fmt.Println("读取文件完成.......")
	return nil
}
// 提取csv参数
func GetValueFromCsvOrDefault(confValue string, list [][]string, i uint64) string {
	if strings.Contains(confValue, "csv") {
		return list[i][0]
	}
	return confValue
}