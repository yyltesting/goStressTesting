// Package client webSocket 客户端
package client

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/link1st/go-stress-testing/conf"
	"github.com/link1st/go-stress-testing/helper"
	"github.com/link1st/go-stress-testing/server/receive"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	connRetry = 3 // 建立连接重试次数
)

var (
	Stopping = make(map[string]bool)
	mu sync.Mutex
	connections int64	//连接数
	mutex = sync.RWMutex{}	//打印连接任务
	stopChan = make(chan bool)	//定时打印连接数
	uri	string
	randomNumbers []int //模客户端
	statisnu    int	//模客户端取一个进行统计tps
)
// WebSocket webSocket
type WebSocket struct {
	conn    *websocket.Conn
	URLLink string
	URL     *url.URL
	IsSsl   bool
	HTTPHeader map[string]string
}

func CreatWebsocket(conn *websocket.Conn) (ws *WebSocket) {
	ws = &WebSocket{
		conn: conn,
	}
	return ws
}
func CreatWebsocketClients(params *conf.Params){
	defer func() {
		stopChan <- true
	}()
	//解析URL
	u, err := url.Parse(params.Request.URL)
	// 解析失败
	if err != nil {
		log.Fatal("url解析失败：", err)
		return
	}
	uri = u.Scheme+"://"+u.Host
	//若发送给多个客户端多个消息需要取20%的抽样客户端作为统计
	if params.Iscollectmsg{
		if params.Concurrency>=10 {
			// 将 uint 类型转换为 float64 类型进行计算
			resultFloat := float64(params.Concurrency) * 0.2
			// 将浮点数结果转换为整数
			resultInt := int(resultFloat)
			randomNumbers = helper.RandNums(int(params.Concurrency),resultInt)
			statisnu = randomNumbers[0]
		}else {
			randomNumbers = append(randomNumbers,1)
			statisnu = randomNumbers[0]
		}

	}

	// 定时输出一次计算结果，延时一秒后时间
	ticker := time.NewTicker(5000 * time.Millisecond)
	go func() {
		for {
			select {
			case <-ticker.C:
				mutex.Lock()
				if connections>0{
					fmt.Println("当前连接数",atomic.LoadInt64(&connections))
				}
				mutex.Unlock()
			case <-stopChan:
				ticker.Stop()
				return
			}
		}
	}()

	//建立连接后并发，建立发送端连接
	if params.ConnectionMode == 1 &&params.Issenders{
		connections = 0
		fmt.Println("正在建立发送端连接...")
		//大于一千的连接数，每次按1000个连接通过协程进行连接
		if params.Concurrency > 1000{
			var wgw sync.WaitGroup
			for z := 1;z <= int(params.Concurrency)/1000;z++{
				wgw.Add(1)
				go func(loop int) {
					defer wgw.Done()
					for x:=uint64((loop-1)*1000);x<1000*uint64(loop);x++ {
						sendTokenlist := conf.CsvList.Tokenlist[x]
						token := sendTokenlist[0]
						conn,err := Dial(uri+"/?t="+token+"&u="+token+"&p=4","","",params.Request.Headers)
						if err != nil {
							fmt.Println("无法建立WebSocket连接：", err)
							// 注意:时间间隔太短会出现连接失败的报错 默认连接时长:20毫秒(公网连接)
							time.Sleep(1 * time.Millisecond)
							break
						}
						receive.GetClientMap().Store(token,conn)
						atomic.AddInt64(&connections,1)
						time.Sleep(1 * time.Millisecond)
					}
				}(z)
				wgw.Wait()
			}
		}else {
			for x:=uint64(0);x<params.Concurrency;x++ {
				sendTokenlist := conf.CsvList.Tokenlist[x]
				token := sendTokenlist[0]
				conn,err := Dial(uri+"/?t="+token+"&u="+token+"&p=4","","",params.Request.Headers)
				if err != nil {
					fmt.Println("无法建立WebSocket连接：", err)
					// 注意:时间间隔太短会出现连接失败的报错 默认连接时长:20毫秒(公网连接)
					time.Sleep(1 * time.Millisecond)
					break
				}
				receive.GetClientMap().Store(token,conn)
				atomic.AddInt64(&connections,1)
				time.Sleep(1 * time.Millisecond)
			}
		}

		fmt.Println("发送端连接建立完成...")
		fmt.Println("共成功连接：",atomic.LoadInt64(&connections))
		//发送端连接数不正确则关闭连接，关闭程序
		if params.Concurrency != uint64(atomic.LoadInt64(&connections))  {
			receive.CloseClient()
			fmt.Println("建立客户端不完整，请重启。。。")
			return
		}
	}
	//建立WebSocket连接,生成不同客户端
	if params.Clients == 2 {
		connections = 0
		fmt.Println("正在建立客户端连接...")
		if params.Concurrency > 1000{
			var wgw sync.WaitGroup
			for z := 1;z <= int(params.Concurrency)/1000;z++{
				wgw.Add(1)
				go func(loop int) {
					defer wgw.Done()
					for x:=uint64((loop-1)*1000);x<1000*uint64(loop);x++ {
						tokenData := conf.CsvList.Recvtokenlist[x]
						token := tokenData[0]
						conn,err := Dial(uri+"/?t="+token+"&u="+token+"&p=4","","",params.Request.Headers)
						if err != nil {
							fmt.Println("无法建立WebSocket连接：", err)
							// 注意:时间间隔太短会出现连接失败的报错 默认连接时长:20毫秒(公网连接)
							time.Sleep(1 * time.Millisecond)
							continue
						}
						receive.GetClientMap().Store(token,conn)
						atomic.AddInt64(&connections,1)
						//go client.Ping(uri,conn,token,"CLIENT")
						//判断是否是群聊客户端收多条消息
						if params.Iscollectmsg{
							receive.WgMessage.Add(1)
							m := false
							//抽样百分之20的人进行模计算
							for _,nu:= range randomNumbers{
								if x== uint64(nu)-1&&x==uint64(statisnu)-1{
									m = true
									go receive.CollectMessages(token,&receive.WgMessage,params,true,true)
								}else if x== uint64(nu)-1&&x!=uint64(statisnu)-1 {
									m = true
									go receive.CollectMessages(token,&receive.WgMessage,params,true,false)
								}
							}
							if !m {
								go receive.CollectMessages(token,&receive.WgMessage,params,false,false)
							}
							//go receive.CollectMessages(token,&wgMessage,params,true,true)
						}
						time.Sleep(1 * time.Millisecond)
					}
				}(z)
				wgw.Wait()
			}
		}else {
			for x:=uint64(0);x<params.Concurrency;x++ {
				tokenData := conf.CsvList.Recvtokenlist[x]
				token := tokenData[0]
				conn,err := Dial(uri+"/?t="+token+"&u="+token+"&p=4","","",params.Request.Headers)
				if err != nil {
					fmt.Println("无法建立WebSocket连接：", err)
					// 注意:时间间隔太短会出现连接失败的报错 默认连接时长:20毫秒(公网连接)
					time.Sleep(1 * time.Millisecond)
					continue
				}
				receive.GetClientMap().Store(token,conn)
				atomic.AddInt64(&connections,1)
				//go client.Ping(uri,conn,token,"CLIENT")
				//判断是否是群聊客户端收多条消息
				if params.Iscollectmsg{
					receive.WgMessage.Add(1)
					m := false
					//抽样百分之20的人进行模计算
					for _,nu:= range randomNumbers{
						if x== uint64(nu)-1&&x==uint64(statisnu)-1{
							m = true
							go receive.CollectMessages(token,&receive.WgMessage,params,true,true)
						}else if x== uint64(nu)-1&&x!=uint64(statisnu)-1 {
							m = true
							go receive.CollectMessages(token,&receive.WgMessage,params,true,false)
						}
					}
					if !m {
						go receive.CollectMessages(token,&receive.WgMessage,params,false,false)
					}
					//go receive.CollectMessages(token,&wgMessage,params,true,true)
				}
				time.Sleep(1 * time.Millisecond)
			}
		}

		fmt.Println("共成功连接：",atomic.LoadInt64(&connections))
		fmt.Println("客户端连接建立完成...")
		if params.Concurrency != uint64(atomic.LoadInt64(&connections))  {
			receive.CloseClient()
			fmt.Println("建立客户端不完整，请重启。。。")
			return
		}
	}
	//一个客户端接收
	if params.Clients == 1 {
		conn,err :=Dial(uri+"/?t="+conf.CsvList.RecvxToken+"&u="+conf.CsvList.RecvxToken+"&p=4","","",params.Request.Headers)
		if err != nil {
			fmt.Println("无法建立WebSocket客户端连接：", err)
			return
		}
		receive.GetClientMap().Store(conf.CsvList.RecvxToken,conn)
		//go client.Ping(uri,conn,token,"CLIENT")
		receive.WgMessage.Add(1)
		if params.Iscollectmsg{
			go receive.CollectMessages(conf.CsvList.RecvxToken,&receive.WgMessage,params,true,true)
		}
		fmt.Println("已成功建立客户端连接")
	}


}
// 解析WS协议URL
func NewWebSocket(urlLink string) (ws *WebSocket) {
	var isSsl bool
	if strings.HasPrefix(urlLink, "wss://") {
		isSsl = true
	}
	//解析URL
	u, err := url.Parse(urlLink)
	// 解析失败
	if err != nil {
		panic(err)
	}
	ws = &WebSocket{
		URLLink: urlLink,
		URL:     u,
		IsSsl:   isSsl,
		HTTPHeader: make(map[string]string),
	}
	return
}
// 复写 websocket库的 Dial 方法 ,增加 httpheader 设置功能
func Dial(url, protocol, origin string, httpHeader map[string]string) (ws *websocket.Conn, err error) {
	var conn *websocket.Conn
	if len(origin)>0{
		httpHeader["origin"] = origin
	}
	convertedHeader := make(http.Header)
	for key, value := range httpHeader {
		convertedHeader.Add(key, value)
	}
	//重试三次建立连接
	for i := 0; i < connRetry; i++ {
		conn,_,err = websocket.DefaultDialer.Dial(url,convertedHeader)
		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, err
	}
	return  conn,nil
}

// GetConn 建立连接方法
func (w *WebSocket) GetConn(csvxtoken string) (ws *WebSocket,err error) {
	var (
		conn *websocket.Conn
		i    int
		urllink string
		u     *url.URL
	)
	if len(csvxtoken) != 0{
		// 修改参数值
		urllink = w.URLLink
		//解析URL
		u, err = url.Parse(urllink)
		// 解析失败
		if err != nil {
			panic(err)
		}
		q := u.Query()
		q.Set("t", csvxtoken)
		q.Set("u", csvxtoken)
		// 将修改后的参数重新设置到URL中
		u.RawQuery = q.Encode()
		urllink = u.String()
	}
	if len(csvxtoken) != 0 {
		conn, err = Dial(urllink, "", w.getOrigin(),w.HTTPHeader)
	}else {
		conn, err = Dial(w.getLink(), "", w.getOrigin(),w.HTTPHeader)
	}
	w.conn = conn
	ws = &WebSocket{
		conn : conn,
		URLLink : urllink,
		URL  :   u,
		IsSsl  : w.IsSsl,
	}
	if err != nil {
		fmt.Println("GetConn 建立连接失败", i, err)
	}
	return ws, nil
}
// getLink 获取连接
func (w *WebSocket) getLink() (link string) {
	return w.URLLink
}

func (w *WebSocket) SetHeader(head map[string]string) {
	w.HTTPHeader = head
}

// getOrigin 获取源连接
func (w *WebSocket) getOrigin() (origin string) {
	origin = "http://"
	if w.IsSsl {
		origin = "https://"
	}
	origin = fmt.Sprintf("%s%s/", origin, w.URL.Host)
	return
}

// Close 关闭ws连接
func (w *WebSocket) Close() (err error) {
	if w == nil {
		return
	}
	if w.conn == nil {
		return
	}
	return w.conn.Close()
}



// Write 发送数据
func (w *WebSocket) Write(body []byte,timeout uint64) (err error) {
	if w.conn == nil {
		err = errors.New("未建立连接")
		return
	}
	now := time.Now()
	w.conn.SetWriteDeadline(now.Add(time.Second * time.Duration(timeout))) //超时抛错
	//_, err = w.conn.Write(body)
	err = w.conn.WriteMessage(websocket.TextMessage,body)
	if err != nil {
		fmt.Println("发送数据失败:", err)
		return
	}
	return
}

// Read 接收数据
func (w *WebSocket) Read(timeout uint64) (msg []byte, err error) {
	if w.conn == nil {
		err = errors.New("未建立连接")
		return
	}
	now := time.Now()
	w.conn.SetReadDeadline(now.Add(time.Second * time.Duration(timeout))) //超时抛错
	msg = make([]byte, 1024)
	//n, err := w.conn.Read(msg)
	_,n,err:= w.conn.ReadMessage()
	msg = n
	if err != nil {
		fmt.Println("接收数据失败:", err)
		return nil, err
	}
	//return msg[:n], nil
	return msg, nil
}

//ping消息
func Ping(uri string ,conn *websocket.Conn, token string,client string)(){
	//可能ping和写入时并发报错 重新执行该方法
	defer func() {
		if r := recover(); r != nil {
			// 在发生 panic 时执行特定的处理
			fmt.Println("发生了并发 panic:", r)
			time.Sleep(time.Second)
			go Ping(uri,conn,token,client)
		}
	}()
	// 加锁
	mu.Lock()
	Stopping[token]=false
	mu.Unlock()
	for {
		mu.Lock()
		stop := Stopping[token]
		mu.Unlock()
		if stop{
			break
		}
		err := conn.WriteMessage(websocket.PingMessage, []byte("ping"))
		if err != nil {
			//出错重连
			fmt.Println("pingerror",client,token,err)
			for r := 0; r < 3; r++{
				conn,_,err = websocket.DefaultDialer.Dial(uri+"/?t="+token+"&u="+token+"&p=4",nil)
				if err == nil {
					break
				}
				time.Sleep(time.Second*2)
			}

		}
		time.Sleep(time.Second*2)

	}

}
func StopPing(token string)(){
	// 加锁
	mu.Lock()
	Stopping[token] = true
	mu.Unlock()
}