// Package statistics 统计数据
package statistics

import (
	"fmt"
	"github.com/link1st/go-stress-testing/conf"
	"github.com/link1st/go-stress-testing/server/db"
	"github.com/link1st/go-stress-testing/server/receive"
	"log"
	"math"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/link1st/go-stress-testing/tools"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/link1st/go-stress-testing/model"
)

var (
	// 输出统计数据的时间
	exportStatisticsTime = 1 * time.Second
	p                    = message.NewPrinter(language.English)
	requestTimeList      []uint64 // 所有请求响应时间
	Handle				 = false   //是否已处理清空websocket长连接
	Maxtimemsg 			[]byte
	ReqTime 			uint64
	Reqnum				int64	//统计实际请求数，给websocket使用关闭通道
	Lastsendmessagetime string
	Lastsendmessage string
	SuccessNum     uint64 // 成功处理数，code为0
	FailureNum     uint64 // 处理失败数，code不为0
)

// ReceivingResults 接收结果并处理
// 统计的时间都是纳秒，显示的时间 都是毫秒
// params.Concurrency 并发数
func ReceivingResults(params *conf.Params,ch <-chan *model.RequestResults, wg *sync.WaitGroup,chWebsocket chan int) {
	defer func() {
		wg.Done()
	}()
	fmt.Println("\n")
	var stopChan = make(chan bool)
	// 时间
	var (
		processingTime uint64 // 处理总时间
		requestTime    uint64 // 请求总时间
		maxTime        uint64 // 最大时长
		minTime        uint64 // 最小时长
		//successNum     uint64 // 成功处理数，code为0
		//failureNum     uint64 // 处理失败数，code不为0
		chanIDLen      int    // 并发数
		chanIDs        []uint64       //= make(map[uint64]bool)
		receivedBytes  int64
		mutex          = sync.RWMutex{}
	)
	statTime := uint64(time.Now().UnixNano())//转换毫秒

	// 错误码/错误个数
	var errCode = &sync.Map{}
	// 定时输出一次计算结果，延时一秒后时间
	ticker := time.NewTicker(exportStatisticsTime)
	go func() {
		for {
			select {
			case <-ticker.C:
				endTime := uint64(time.Now().UnixNano())
				mutex.Lock()
				go calculateData(params,processingTime, endTime-statTime, maxTime, minTime, SuccessNum, FailureNum,
					chanIDLen, errCode, receivedBytes,chWebsocket)
				mutex.Unlock()
			case <-stopChan:
				ticker.Stop()
				return
			}
		}
	}()
	header()
	done := false
	for !done{
		select {
		case <- params.Ctx.Done():
			done = true
			goto end
		case data, ok := <- ch:
			//websocket不以发送完为结束，以接收完为结束
			if !ok && params.Request.Form == model.FormTypeWebSocket{
				done = true
				break
			}
			if !ok{
				done = true
				goto end
			}
			mutex.Lock()
			//fmt.Println("处理一条数据", data.ID, data.Time, data.IsSucceed, data.ErrCode)
			processingTime = processingTime + data.Time //计算总时间

			if maxTime <= data.Time {
				maxTime = data.Time
			}
			if minTime == 0 {
				minTime = data.Time
			} else if minTime > data.Time && data.Time != 0 {
				minTime = data.Time
			}
			// 是否请求成功
			if data.IsSucceed == true {
				SuccessNum = SuccessNum + 1
			} else {
				FailureNum = FailureNum + 1
			}
			// 统计错误码
			if value, ok := errCode.Load(data.ErrCode); ok {
				valueInt, _ := value.(int)
				errCode.Store(data.ErrCode, valueInt+1)
			} else {
				errCode.Store(data.ErrCode, 1)
			}
			receivedBytes += data.ReceivedBytes
			//fmt.Println(data.ChanID)
			//通过管道的chid统计起的完成的并发数  chaniDs最开始是空，所以反向判断
			//if _, ok := chanIDs[data.ChanID]; !ok{
			//	chanIDs[data.ChanID] = true
			//	chanIDLen = len(chanIDs)
			//}

			chanIDs = append(chanIDs, data.ChanID)
			chanIDLen = len(chanIDs)

			requestTimeList = append(requestTimeList, data.Time)
			mutex.Unlock()
		}
	}
end:
	// 数据全部接受完成，停止定时输出统计数据
	stopChan <- true
	endTime := uint64(time.Now().UnixNano())
	requestTime = endTime - statTime
	log.Println("发送延时最大的消息：",string(Maxtimemsg))
	log.Println("发送延时时间：",fmt.Sprintf("%d %s",ReqTime/1e6,"ms"))

	log.Printf("\n\n")
	log.Println("*************************  结果 stat  ****************************")
	log.Println("处理协程数量:", params.Concurrency)
	// fmt.Println("处理协程数量:", params.Concurrency, "程序处理总时长:", fmt.Sprintf("%.3f", float64(processingTime/params.Concurrency)/1e9), "秒")
	log.Println("请求总数（并发数*请求数 -c * -n）:", SuccessNum+FailureNum, "程序运行时间:",
		fmt.Sprintf("%.3f", float64(requestTime)/1e9),
		"秒", "successNum:", SuccessNum, "failureNum:", FailureNum)
	printTop(requestTimeList)
	log.Println("*************************  结果 end   ****************************")
	log.Printf("\n\n")

	if params.Request.Form == model.FormTypeWebSocket{
		if params.Iscollectmsg{
			Startrecv_time := time.Unix(receive.Startrecv_time/1e9, 0) // 将时间戳转换为time.Time类型
			Endrecv_time := time.Unix(receive.Endrecv_time/1e9, 0)
			formatted1 := Startrecv_time.Format("2006-01-02 15:04:05") // 使用标准格式化字符串
			formatted2 := Endrecv_time.Format("2006-01-02 15:04:05")
			log.Println("消息接收开始时间：",formatted1)
			log.Println("消息接收结束时间：",formatted2)
		}
		log.Println("最后一条消息发送时间：",Lastsendmessagetime)
		log.Println("最后一条发送消息：",Lastsendmessage)
	}
	if params.Clients ==1 ||params.Clients ==2{
		log.Println("模客户端共计收到消息：",atomic.LoadInt64(&receive.CountMessage))
	}
	if params.Iscollectmsg{
		var sumMessage int64 = 0
		var failclients int64 = 0
		if atomic.LoadInt64(&receive.CountSuccessMessage)>0{
			log.Println("模客户端收到正确消息：",atomic.LoadInt64(&receive.CountSuccessMessage))
			log.Println("模客户端消息平均延时ms：", atomic.LoadInt64(&receive.Runtime)/1e6/atomic.LoadInt64(&receive.CountSuccessMessage))
			log.Println("模客户端消息最大延时ms：", receive.Maxtime/1e6)
			log.Println("模客户端消息最大延时的content：", receive.Maxmsg)
			log.Println("模客户端总接收时间耗时ms：",(receive.ModelrecvEndtime-receive.ModelrecvStarttime)/1e6)
		}
		if params.Clients ==2{
			// 复制 sync.Map 到普通 map
			messageMap := make(map[string]int64)
			receive.ChMessage.Range(func(k, v interface{}) bool {
				messageMap[k.(string)] = v.(int64)
				return true
			})
			for key,value := range messageMap{
				//找出遗漏消息客户端,和csv参数对应有关系
				sumMessage =sumMessage+value
				if params.ExecutionTime == 0 || params.Startsnum == 0 {
					if params.Toclients && uint64(value) != params.Concurrency*params.TotalNumber{
						failclients ++
						log.Println("遗漏消息客户端：",key,value)
					}else if !params.Toclients && params.Clients ==2 && uint64(value) != params.TotalNumber {
						failclients ++
						log.Println("遗漏消息客户端：",key,value)
					}else if !params.Toclients && params.Clients ==1 && uint64(value) != params.Concurrency*params.TotalNumber{
						failclients ++
						log.Println("遗漏消息客户端：",key,value)
					}
				}
			}
		}
		//TPS 请求总数除以处理时间 //处理时间=接收结束时间-发送最早的时间
		if receive.Startrecv_time<receive.Startsend_time{
			receive.Startsend_time = receive.Startrecv_time
		}
		t := receive.Endrecv_time-receive.Startsend_time
		if t==0{
			t=1
		}
		log.Println("总处理时间ms：",t/1e6)
		//和csv参数对应有关系 如果是群toclients一个发送端向多个接收端接收 有转发机制，所以发送端也会收到消息 实际接收消息应该是并发数三次幂倍
		if !params.Toclients {
			log.Println("平均消息处理TPS：",atomic.LoadInt64(&receive.Sum_message)*1e9/t)
		}
		if params.Toclients {
			log.Println("平均消息请求处理TPS：",atomic.LoadInt64(&receive.Sum_message)*int64(params.Concurrency)*1e9/t)
		}
		if atomic.LoadInt64(&receive.Runtime) >0{
			log.Println("平均消息处理瞬时qps：",int64(params.Concurrency) *atomic.LoadInt64(&receive.CountSuccessMessage)*1e9/atomic.LoadInt64(&receive.Runtime))
		}
		if params.Clients ==2 {
			log.Println("有遗漏消息客户端总数：",failclients)
			log.Println("所有客户端接收到的总消息：",sumMessage)
			if params.Toclients{
				//群的发送者也会接收到消息，10个用户每个用户在10个群每个群发1条消息 10*10*10； 再加上模客户端接收的消息 10个用户每个用户在群里接收10条 10*10
				log.Println("本次测试理想化-群消息转发总数：",math.Pow(float64(params.Concurrency), 3)+math.Pow(float64(params.Concurrency), 2))
			}else {
				//群的发送者也会接收到消息，10个用户分别在10个群每个群发1条消息 10*10； 再加上模客户端接收的消息 10个用户每个用户在群里接收1条 10*1
				log.Println("本次测试理想化-群消息转发总数：",math.Pow(float64(params.Concurrency), 2)+math.Pow(float64(params.Concurrency), 1))
			}
		}
		log.Println("总接收时间耗时ms：",(receive.Endrecv_time-receive.Startrecv_time)/1e6)
	}
	if params.Request.Form == model.FormTypeWebSocket{
		db.LocalDb()
		var count int64
		Startsend_time := time.Unix(receive.Startsend_time/1e9-28801, 0)
		//正式
		//Startsend_time := time.Unix(receive.Startsend_time/1e9-1, 0)
		formatted1 := Startsend_time.Format("2006-01-02 15:04:05") // 使用标准格式化字符串
		err := db.WSDb.Model(&db.MsgInfo{}).Where("send_time>=?", formatted1).Count(&count).Error
		if err != nil {
			log.Println("查询出错",err)
		}
		log.Println("db实际收到的发送消息条数：",count)
	}
}

// printTop 排序后计算 top 90 95 99
func printTop(requestTimeList []uint64) {
	if requestTimeList == nil {
		return
	}
	all := tools.MyUint64List{}
	all = requestTimeList
	sort.Sort(all)
	log.Println("tp90:", fmt.Sprintf("%.3f", float64(all[int(float64(len(all))*0.90)]/1e6)))
	log.Println("tp95:", fmt.Sprintf("%.3f", float64(all[int(float64(len(all))*0.95)]/1e6)))
	log.Println("tp99:", fmt.Sprintf("%.3f", float64(all[int(float64(len(all))*0.99)]/1e6)))
}

// calculateData 计算数据
func calculateData(params *conf.Params,processingTime, requestTime, maxTime, minTime, successNum, failureNum uint64,
	chanIDLen int, errCode *sync.Map, receivedBytes int64,chWebsocket chan int) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("关闭出现panic", err)
		}
	}()
	if processingTime == 0 {
		processingTime = 1
	}
	var (
		qps              float64
		averageTime      float64
		maxTimeFloat     float64
		minTimeFloat     float64
		requestTimeFloat float64
	)
	//QPS每秒处理请求次数，具体是指发出请求到服务器，并且服务器成功处理并返回结果
	//每秒事务处理能力  QPS(TPS) = 并发数/平均响应时间 请求数*协程数/请求耗时 (每秒) ==》QPS=并发数/(总耗时/请求数) 这里的processingTime是所有请求的时间的集合
	if params.Startsnum <= 1 && params.ExecutionTime ==0{
		//没有循环请求
		//a := float64(processingTime)/1e9 //总请求时间s
		//b := float64(successNum+failureNum)	//总请求数
		//c := b/(a/float64(chanIDLen))
		//qps = c

		qps = float64((successNum*1e9+failureNum*1e9)*params.Concurrency) / float64(processingTime)
	}else {
		//循环请求
		a := float64(processingTime)/1e9 //总请求时间s
		b := float64(successNum+failureNum)	//总请求数
		c := b/(a/float64(params.Concurrency))
		qps = c
	}
	// 平均时长 总耗时/总请求数/并发数 纳秒=>毫秒
	if successNum != 0 && params.Concurrency != 0 {
		averageTime = float64(processingTime) / float64(successNum*1e6+failureNum*1e6)
	}
	// 纳秒=>毫秒
	maxTimeFloat = float64(maxTime) / 1e6
	minTimeFloat = float64(minTime) / 1e6
	requestTimeFloat = float64(requestTime) / 1e9 //总运行时间
	// 打印的时长都为毫秒
	table(successNum, failureNum, errCode, qps, averageTime, maxTimeFloat, minTimeFloat, requestTimeFloat, chanIDLen,
		receivedBytes)
	var count uint64
	if params.Startsnum != 0 {
		num := params.Concurrency/params.Startsnum
		for i := uint64(1) ; i <= num;i++{
			count = count+params.Startsnum*i
		}
	}
	//判断全部执行完 websocket长连接断开
	if params.Request.Form == model.FormTypeWebSocket && params.Request.Keepalive && !Handle  {
		//发送数则为接收数，发送一条接收一条，如果发送完了则为结束
		if !params.Iscollectmsg{
			//阶梯压测
			if params.Startsnum != 0 && successNum+failureNum == count{
				for range chWebsocket{
					Handle = true
				}
			}
			if params.Startsnum == 0 && atomic.LoadInt64(&Reqnum) == int64(params.Concurrency){
				for range chWebsocket {
					Handle = true
				}
			}
		}else {
			//以接收端是否收到为结束
			//fmt.Println(atomic.LoadInt64(&CountClient))
			if params.Clients ==2 && atomic.LoadInt64(&receive.CountClient)==int64(params.Concurrency){
				for range chWebsocket{
					Handle = true
				}
			}
			if params.Clients ==1 && atomic.LoadInt64(&receive.CountClient)==1{
				for range chWebsocket{
					Handle = true
				}
			}
		}
	}
}

// header 打印表头信息
func header() {
	log.Printf("\n\n")
	// 打印的时长都为毫秒 总请数
	log.Println("───────────────────┬─────┬───────┬───────┬───────┬────────┬────────┬────────┬────────┬────────┬────────┬───────")
	log.Println("       reqtime      总耗时  请求数   成功数   失败数     QPS    最长耗时  最短耗时  平均耗时   下载字节   字节每秒   状态码  ")
	log.Println("───────────────────┼─────┼───────┼───────┼───────┼────────┼────────┼────────┼────────┼────────┼────────┬───────")
	return
}

// table 打印表格
func table(successNum, failureNum uint64, errCode *sync.Map,
	qps, averageTime, maxTimeFloat, minTimeFloat, requestTimeFloat float64, chanIDLen int, receivedBytes int64) {
	var (
		speed int64
	)
	if requestTimeFloat > 0 {
		speed = int64(float64(receivedBytes) / requestTimeFloat)
	} else {
		speed = 0
	}
	var (
		receivedBytesStr string
		speedStr         string
	)
	// 判断获取下载字节长度是否是未知
	if receivedBytes <= 0 {
		receivedBytesStr = ""
		speedStr = ""
	} else {
		receivedBytesStr = p.Sprintf("%d", receivedBytes)
		speedStr = p.Sprintf("%d", speed)
	}
	reqTime := time.Now()
	// 定义格式化布局
	layout := "2006-01-02 15:04:05"
	// 将当前时间格式化为指定布局
	formattedTime := reqTime.Format(layout)
	// 打印的时长都为毫秒
	result := fmt.Sprintf("%s│%4.0fs│%7d│%7d│%7d│%8.2f│%8.2f│%8.2f│%8.2f│%8s│%8s│%v",
		formattedTime, requestTimeFloat, chanIDLen, successNum, failureNum, qps, maxTimeFloat, minTimeFloat, averageTime,
		receivedBytesStr, speedStr,
		printMap(errCode))
	log.Println(result)
	return



}

// printMap 输出错误码、次数 节约字符(终端一行字符大小有限)
func printMap(errCode *sync.Map) (mapStr string) {
	var (
		mapArr []string
	)
	errCode.Range(func(key, value interface{}) bool {
		mapArr = append(mapArr, fmt.Sprintf("%v:%v", key, value))
		return true
	})
	sort.Strings(mapArr)
	mapStr = strings.Join(mapArr, ";")
	return
}
