// Package main go 实现的压测工具
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/link1st/go-stress-testing/conf"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/link1st/go-stress-testing/model"
	"github.com/link1st/go-stress-testing/server"
)

// array 自定义数组参数,切片
type array []string

// String string 方法
func (a *array) String() string {
	return fmt.Sprint(*a)
}

// Set set 方法
func (a *array) Set(s string) error {
	*a = append(*a, s)

	return nil
}

var (
	concurrency uint64 = 1       // 并发数
	totalNumber uint64 = 1       // 请求数(单个并发/协程)
	executionTime  uint64 = 0       // 请求执行时间（循环请求） min
	debugStr           = "false" // 是否是debug
	requestURL         = ""      // 压测的url 目前支持，http/https ws/wss
	path               = ""      // curl文件路径 http接口压测，自定义参数设置
	verify             = ""      // verify 验证方法 在server/verify中 http 支持:statusCode、json webSocket支持:json
	headers     array            // 自定义头信息传递给服务器
	body               = ""      // HTTP POST方式传送数据
	maxCon             = 1       // 单个连接最大请求数
	code               = 200     // 成功状态码
	http2              = false   // 是否开http2.0
	keepalive          = true   // 是否开启长连接
	cpuNumber          = 4       // CUP 核数，默认为一核，一般场景下单核已经够用了
	apptimeout     int64  = 0       // 超时时间，默认不设置
	timeout  uint64 = 60        // 接口超时时间，默认60秒
	interval	uint64 = 0	 	 //阶梯压测时间启动间隔S
	startsnum	uint64 = 0		 //阶梯压测每秒启动数
	clients		int64  = 0   //是否多个客户端 默认0，1为一个，2为多个
	issenders	       = "false"   //是否多个发送端
	iscollectmsg	   = "false" // 是否将消息收集起来
	toclients		   = "false" //默认每个发送者只向一个群或人发送
	redirect		   = true
	connectionmode	 int64  = 1 // 1:顺序建立长链接 2:并发建立长链接
	variable		   = "" //变量通过,号分割 分别为sendtoken,sendid,recvtoken,recvid
)

// 在main函数之前执行的函数，处理类型
func init() {
	flag.Uint64Var(&concurrency, "c", concurrency, "并发数")
	flag.Uint64Var(&totalNumber, "n", totalNumber, "请求数(单个并发/协程)")
	flag.Uint64Var(&executionTime, "t", executionTime, "请求时间min")
	flag.StringVar(&debugStr, "d", debugStr, "调试模式")
	flag.StringVar(&requestURL, "u", requestURL, "压测地址")
	flag.StringVar(&path, "p", path, "curl文件路径")
	flag.StringVar(&verify, "v", verify, "验证方法 http 支持:statusCode、json webSocket支持:json")
	flag.Var(&headers, "H", "自定义头信息传递给服务器 示例:-H 'Content-Type: application/json'")
	flag.StringVar(&body, "data", body, "HTTP POST方式传送数据")
	flag.IntVar(&maxCon, "m", maxCon, "单个host最大连接数")
	flag.IntVar(&code, "code", code, "请求成功的状态码")
	flag.BoolVar(&http2, "http2", http2, "是否开http2.0")
	flag.BoolVar(&keepalive, "k", keepalive, "是否开启长连接")
	flag.IntVar(&cpuNumber, "cpuNumber", cpuNumber, "CUP 核数，默认为一核")
	flag.Int64Var(&apptimeout, "apptimeout", apptimeout, "程序超时时间(持续时间) 单位 秒,默认不设置")
	flag.Uint64Var(&timeout, "timeout", timeout, "接口超时时间(持续时间) 单位 秒,默认不设置")
	flag.Uint64Var(&interval, "i", interval, "阶梯压测时间启动间隔S")
	flag.Uint64Var(&startsnum, "s", startsnum, "阶梯压测每秒启动数，注意需要和并发数整除")
	flag.Int64Var(&clients, "clients", clients, "是否多个客户端,1一个、2多个、0不设置则为单接口")
	flag.StringVar(&issenders, "issenders", issenders, "是否多个发送端")
	flag.StringVar(&iscollectmsg, "iscollectmsg", iscollectmsg, "是否将消息收集起来")
	flag.StringVar(&toclients, "toclients", toclients, "一个发送端是否向多个客户端发送")
	flag.BoolVar(&redirect, "redirect", redirect, "是否重定向")
	flag.Int64Var(&connectionmode, "connectionmode", connectionmode, "发送建立连接方式")
	flag.StringVar(&variable, "variable", variable, "变量通过,号分割 分别为sendtoken,sendid,recvtoken,recvid")
	// 解析参数
	flag.Parse()
}

// main go 实现的压测工具
// 编译可执行文件
//
//go:generate go build main.go
func main() {
	// 保存日志输出log打印
	filename := "go-stress-"+fmt.Sprintf("%d", time.Now().Unix())
	file, err := os.OpenFile("./logs/"+filename+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	// 使用 io.MultiWriter 将日志输出到文件和终端
	multiWriter := io.MultiWriter(file, os.Stdout)
	log.SetOutput(multiWriter)

	//设置cpu核数函数
	runtime.GOMAXPROCS(cpuNumber)
	if concurrency == 0 || totalNumber == 0 || (requestURL == "" && path == "") {
		fmt.Printf("示例: go run main.go -c 1 -n 1 -u https://www.baidu.com/ \n")
		fmt.Printf("压测地址或curl路径必填 \n")
		fmt.Printf("当前请求参数: -c %d -n %d -d %v -u %s \n", concurrency, totalNumber, debugStr, requestURL)
		flag.Usage()
		return
	}
	//阶梯压测启动数判断是否整除
	if startsnum>0&&concurrency%startsnum!=0{
		fmt.Println("每秒启动数需要能与总数整除")
		return
	}
	debug := strings.ToLower(debugStr) == "true"
	isSenders := strings.ToLower(issenders) == "true"
	isCollectmsg := strings.ToLower(iscollectmsg) == "true"
	toClients := strings.ToLower(toclients) == "true"
	//生成请求结构体
	request, err := model.NewRequest(requestURL, verify, code, time.Duration(timeout)*time.Second, debug, path, headers,
		body, maxCon, http2, keepalive,redirect)
	if err != nil {
		fmt.Printf("参数不合法 %v \n", err)
		return
	}
	fmt.Printf("\n 开始启动  并发数:%d 请求数:%d 请求参数: \n", concurrency, totalNumber)
	//打印请求参数
	request.Print()
	// 开始处理，创建上下文
	ctx := context.Background()
	if apptimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(apptimeout)*time.Second)
		defer cancel()
		deadline, ok := ctx.Deadline()
		if ok {
			fmt.Printf(" deadline %s", deadline)
		}
	}
	// 处理 ctrl+c 信号
	ctx, cancelFunc := context.WithCancel(ctx)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		<-c
		cancelFunc()
	}()
	//处理请求
	param := &conf.Params{
		Concurrency :	concurrency,
		TotalNumber :	totalNumber,
		Request		:	request,
		Interval	:	interval,
		Startsnum	:	startsnum,
		Clients		:	clients,
		Issenders		:	isSenders,
		ExecutionTime	:executionTime,
		Timeout     :	timeout,
		Ctx			: 	ctx,
		Iscollectmsg	:	isCollectmsg,
		Toclients   :   toClients,
		ConnectionMode  :  connectionmode,
		Variable	:	variable,
	}
	//读取变量
	err = conf.InitVariable(param)
	if err != nil{
		return
	}
	server.Dispose(param)
	return
}
