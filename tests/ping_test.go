package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"testing"
	"time"
)

func TestPing(t *testing.T) {
	//url := "ws://10.8.200.13:12345?u=ub014pick9h4&t=1"  //服务器地址
	url := "wss://dev-gateway-habox.wss1.cn?t=2oYlH6GLAQt5xiIvxaLaCwodGzT&p=5"  //服务器地址
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal(err)
	}
	ws.SetPongHandler(func(appData string) error {
		fmt.Println("recv pong msg",appData)
		return nil
	})
	go func() {
		for {
			err := ws.WriteMessage(websocket.PingMessage, []byte("ping"))
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("ping...")
			time.Sleep(time.Second*2)
		}
	}()
	time.Sleep(time.Second*1)
	//oid := fmt.Sprintf("%d", time.Now().UnixNano())
	//go func() {
	//	for {
//			err = ws.WriteMessage(websocket.TextMessage, []byte(`{
//   "data": "{\"content\":\"v\",\"sendId\":\"ud7oegiwcfo6\",\"sessionType\":2,\"groupId\":\"dq71azd0r84\",\"contentType\":101}",
//   "msgType": 1003,
//   "operationId": "`+oid+`"
//}`))
//			if err != nil {
//				log.Fatal(err)
//			}
//			fmt.Println("send msg...")
//			time.Sleep(time.Second*5)
	//	}
	//}()

	for {
		_, data, err := ws.ReadMessage()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("receive: ", string(data))
	}
}