package receive

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/link1st/go-stress-testing/conf"
	"log"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

//设置接收消息者的客户端通道
var (
	clientMap = new(sync.Map)
	ChMessage  =new(sync.Map)
	CountClient	int64 =0 //实际客户端接收数
	CountMessage int64  //模总共收到的消息条数
	CountSuccessMessage int64  //模校验有效的消息条数
	Sum_message int64 //总消息数
	Messages = new(sync.Map) //客户端对应消息
	Runtime int64	//总的运行时间
	Maxtime int64 = 0   //接收最长的延时
	Maxmsg 	string	//延时最长的一条消息
	Startsend_time int64 = 0   //消息发送起始时间
	Endrecv_time int64 = 0	//消息接收结束时间
	Startrecv_time int64 = 0 //消息接收开始时间
	ModelrecvEndtime int64 = 0 //模消息接收结束时间
	ModelrecvStarttime int64 = 0 //模消息接收结束时间
	stopChan = make(chan bool)	//定时打印qps任务
	mutex = sync.RWMutex{}  //时间定时打印锁
	statisticsMessage sync.Mutex //统计map锁
	WgMessage sync.WaitGroup //客户端接收消息
)

func GetClientMap() *sync.Map {
	return clientMap
}
func CloseClient(){
	clientMap.Range(func(_, v interface{}) bool {
		ws := v.(*websocket.Conn)
		ws.Close() //关闭连接
		return true
	})
}
// WebSocketResponseJSON 返回数据结构体，返回值为json
type WebSocketResponseJSON struct {
	MsgType      	  int `json:"msgType"`
	OperationId      string `json:"operationId"`
	Code		   	  int `json:"code"`
	Message			  string	`json:"message"`
	Data 		string	`json:"data"`
}
type MsgdataListJson struct {
	MsgdataList		[]DataJson `json:"msgDataList"`
}

type DataJson struct {
	SendId      	  string `json:"sid"`
	RecvId      string `json:"rid"`
	GroupId		   	  string `json:"gid"`
	Content 			  string	`json:"c"`
	Sendtime	string `json:"st"`
}
//接收到单条消息
func SingleMessage(k string) (*websocket.Conn, []byte,error) {
	//这里需要修改使用接收者进行查看消息，统计收到消息的时长
	conn, ok := clientMap.Load(k)
	if !ok {
		log.Println("未找到客户端")
	}
	c := conn.(*websocket.Conn)
	msg := make([]byte, 0, 1024)
	now := time.Now()
	c.SetReadLimit(128*1024)
	c.SetReadDeadline(now.Add(time.Second * 60)) //超时抛错
	//m, err := c.Read(msg)
	var err error
	_,msg,err = c.ReadMessage()
	if err != nil {
		log.Println("接收数据失败:", err)
		//msg = msg[:m]
		c.Close()
		return conn.(*websocket.Conn),msg,err
	}
	//msg = msg[:m]
	atomic.AddInt64(&CountMessage,1)
	c.Close()
	//全部接收完后再关闭ping
	//client.StopPing(k)
	return conn.(*websocket.Conn),msg,nil
}

//客户端收到的消息存起来
func CollectMessages(k string,wg *sync.WaitGroup,params *conf.Params,clientModle bool,statis bool) () {
	defer wg.Done()
	var (
		wgRun          sync.WaitGroup // 统计时间
		w 			   sync.Mutex
		messagenum     int64 = 0
		stopcollect    = make(chan bool)
	)
	conn, ok := clientMap.Load(k)
	if !ok {
		fmt.Println("未找到客户端")
		return
	}
	c := conn.(*websocket.Conn)
	c.SetReadLimit(128*1024)
	//defer client.StopPing(k)
	//全部接收完才关闭
	defer func() {
		if r := recover(); r != nil {
			// 在发生 panic 时执行特定的处理
			log.Println("接收端发生了panic:", r)
			time.Sleep(time.Millisecond*10)
			//关闭客户端数+1
			atomic.AddInt64(&CountClient,1)
			//关闭模的统计打印
			if clientModle&&statis{
				//如果是超时则关闭接受通道，退出程序
				stopChan <- true
			}
		}
		go func() {
			statisticsMessage.Lock()
			ChMessage.Store(k, atomic.LoadInt64(&messagenum))
			statisticsMessage.Unlock()
		}()
		//接收不到直接关闭连接，避免还在接收造成服务端统计不一致
		c.Close()
	}()
	//作为打印统计的客户端只设置一个
	if clientModle && statis{
		// 定时输出一次计算结果，延时一秒后时间
		ticker := time.NewTicker(1000 * time.Millisecond)
		go func() {
			for {
				select {
				case <-ticker.C:
					mutex.Lock()
					//TPS 处理请求数/处理时间
					//处理时间=接收结束时间-发送最早的时间
					if Endrecv_time>Startsend_time&&!params.Toclients {
						t := Endrecv_time-Startsend_time
						log.Println("消息处理瞬时TPS：",atomic.LoadInt64(&Sum_message)*1e9/t)
					}
					//toclients的计算可能与csv参数相关
					if Endrecv_time>Startsend_time&&params.Toclients {
						t := Endrecv_time-Startsend_time
						log.Println("消息处理瞬时TPS：",atomic.LoadInt64(&Sum_message)*int64(params.Concurrency)*1e9/t)
					}
					if atomic.LoadInt64(&Runtime) >0{
						log.Println("消息处理瞬时qps：",int64(params.Concurrency) * atomic.LoadInt64(&CountSuccessMessage)*1e9/atomic.LoadInt64(&Runtime))
					}
					mutex.Unlock()
				case <-stopChan:
					ticker.Stop() //结束关闭ticker避免资源泄露
					return
				}
			}
		}()
	}
	//接收消息
	go func(c *websocket.Conn,wgRun *sync.WaitGroup,w *sync.Mutex,messagenum *int64,clientModle bool,stopcollect chan bool,params *conf.Params,k string) {
		msg := make([]byte,0, 128*1024)
		firstrecv := true
		for{
			//设置消息接收超时时间
			if firstrecv {
				now := time.Now()
				c.SetReadDeadline(now.Add(time.Second * time.Duration(params.Timeout)*2)) //一次的消息多等一会
				firstrecv = false
			}else {
				now := time.Now()
				c.SetReadDeadline(now.Add(time.Second * time.Duration(params.Timeout))) //超时抛错,还未发送完30秒没数据就超时
			}
			var err error
			_,msg,err = c.ReadMessage() //此处接收消息可能一直在等阻塞
			//log.Println("Message length:", len(msg))
			if err != nil { //接收消息超时进行关闭接收器
				<- stopcollect
				break
			}
			recvtime := time.Now().UnixNano()
			//fmt.Println("recv-----",k,string(msg))
			//messagenum ++
			//模进行计算，非模只统计数量
			if clientModle{
				wgRun.Add(1)
				//这里进入消息统计
				go MessageStats(msg,recvtime,wgRun,w,messagenum,true,stopcollect,params)
				//atomic.AddInt64(&CountMessage,1)
			}else {
				wgRun.Add(1)
				//这里进入消息统计
				go MessageStats(msg,recvtime,wgRun,w,messagenum,false,stopcollect,params)
			}
		}
	}(c,&wgRun,&w,&messagenum,clientModle,stopcollect,params,k)
	select {
	case <- params.Ctx.Done():
	case <- stopcollect:
	}

	//关闭客户端数+1
	atomic.AddInt64(&CountClient, 1)
	//关闭模的统计打印
	if clientModle && statis {
		//如果是超时则关闭接受通道，退出程序
		stopChan <- true
	}
	wgRun.Wait()
}
//消息存起来统计,消息变成了数组
func MessageStats(msg []byte,recvtime int64,wgRun *sync.WaitGroup,w *sync.Mutex,messagenum *int64,
	clientModle bool,stopcollect chan bool,params *conf.Params)  {
	w.Lock()
	defer wgRun.Done()
	defer w.Unlock()
	responseJSON := &WebSocketResponseJSON{} //josn模板
	err := json.Unmarshal(msg, responseJSON) //解析json
	if err != nil{
		log.Println("接收端无法解析json",err,string(msg))
		return
	}
	data := responseJSON.Data
	msgdataListJson := &MsgdataListJson{}                      //josn模板
	err = json.Unmarshal([]byte(data), msgdataListJson) //解析json
	if err != nil{
		log.Println("接收端无法解析jsondatadata",err,string(msg))
		return
	}
	//计算有多少条消息
	atomic.AddInt64(messagenum,int64(len(msgdataListJson.MsgdataList))) //当前客户端条数
	atomic.AddInt64(&Sum_message,int64(len(msgdataListJson.MsgdataList))) //总接收条数
	//计算消息接收结束时间
	if Endrecv_time == 0 || recvtime > Endrecv_time{
		Endrecv_time = recvtime
	}
	//计算消息接收开始时间
	if Startrecv_time == 0 || recvtime < Startrecv_time{
		Startrecv_time = recvtime
	}
	//模才去解析消息
	if clientModle {
		atomic.AddInt64(&CountMessage,int64(len(msgdataListJson.MsgdataList)))
		for i, msgData := range msgdataListJson.MsgdataList {
			//计算时间
			timestampstr := strings.Split(msgData.Content, "_")[0]
			timestamp, err1 := strconv.ParseInt(timestampstr, 10, 64)
			if err1 != nil {
				log.Println("接收端无法解析jsoncontent", err1, msgData.Content)
				return
			}
			//计算接收耗时最大的消息
			rtime := recvtime - timestamp
			if Maxtime == 0 {
				Maxtime = rtime
				Maxmsg = msgData.Content
			} else if rtime > Maxtime {
				Maxtime = rtime
				Maxmsg = msgData.Content+"---"+ "接收时间："+strconv.FormatInt(recvtime, 10)
			}
			//计算消息接收结束时间
			if ModelrecvEndtime == 0 || recvtime > ModelrecvEndtime{
				ModelrecvEndtime = recvtime
			}
			//计算消息接收开始时间
			if ModelrecvStarttime == 0 || recvtime < ModelrecvStarttime{
				ModelrecvStarttime = recvtime
			}
			//取第一条的运行时间，因为在同一个数组里面误差不大
			if i==0{
				atomic.AddInt64(&Runtime,rtime)
			}
			atomic.AddInt64(&CountSuccessMessage, 1)
		}
	}
	//当前接收端接收完毕 关闭接收通道
	if params.ExecutionTime == 0 || params.Startsnum == 0 {
		if params.Toclients && atomic.LoadInt64(messagenum) >= int64(params.Concurrency*params.TotalNumber) {
			stopcollect <- true
		} else if params.Clients == 2 && atomic.LoadInt64(messagenum) >= int64(params.TotalNumber) {
			stopcollect <- true
		} else if params.Clients == 1 && atomic.LoadInt64(messagenum) >= int64(params.Concurrency*params.TotalNumber) {
			stopcollect <- true
		}
	}

}

//消息存起来统计
//func MessageStats1(msg []byte,recvtime int64,wgRun *sync.WaitGroup,w *sync.Mutex,messagenum int64)  {
//	defer wgRun.Done()
//	defer w.Unlock()
//	w.Lock()
//	responseJSON := &WebSocketResponseJSON{} //josn模板
//	err := json.Unmarshal(msg, responseJSON) //解析json
//	if err != nil{
//		fmt.Println("无法解析json",err,string(msg))
//		return
//	}
//	data := responseJSON.Data
//	dataJson := &DataJson{}                      //josn模板
//	err = json.Unmarshal([]byte(data), dataJson) //解析json
//	if err != nil{
//		fmt.Println("无法解析jsondatadata",err,string(msg))
//		return
//	}
//	//计算时间
//	timestampstr := strings.Split(dataJson.Content, "_")[0]
//	timestamp,err1 := strconv.ParseInt(timestampstr,10,64)
//	if err1 != nil{
//		fmt.Println("无法解析jsoncontent",err1,dataJson.Content)
//		return
//	}
//	//计算接收耗时最大的消息
//	rtime := recvtime-timestamp
//	if Maxtime == 0 {
//		Maxtime = rtime
//		Maxmsg = dataJson.Content
//	}else if rtime > Maxtime {
//		Maxtime = rtime
//		Maxmsg = dataJson.Content
//	}
//	//计算总接收时间
//	if RcvstratTime ==0{
//		RcvstratTime = recvtime
//	}
//	if recvtime<RcvstratTime{
//		RcvstratTime = recvtime
//	}
//	if recvtime>RcvendTime{
//		RcvendTime = recvtime
//	}
//	Runtime = rtime+ Runtime
//	atomic.AddInt64(&CountSuccessMessage,1)
//
//}