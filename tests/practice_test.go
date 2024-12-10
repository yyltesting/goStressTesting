package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestSync(t *testing.T){
	a := []int{1,2,3,4,5,6,7}
	b := len(a)-1
	c := 0
	for c <= b{
		x := a[b]
		a[b] = a[c]
		a[c] = x

		c++
		b--
	}
	fmt.Println(a)


	var wg sync.WaitGroup
	wg.Add(2)

	numCh := make(chan int)
	charCh := make(chan string)

	q := []string{"A","B","C","D"}
	go func() {
		defer wg.Done()
		for i := 1; i <= 4; i++ {
			numCh <- i
			fmt.Print(<-charCh)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i <= len(q)-1; i++ {
			fmt.Print(<-numCh)
			charCh <- q[i]
		}
	}()

	wg.Wait()
	fmt.Println()
}

func TestCh(t *testing.T){
	ch := make(chan int,0)

	// 启动 10 个 goroutine 向通道发送数据
	for i := 0; i < 10; i++ {
		go func() {
			ch <- 1
		}()
	}
	time.Sleep(2 * time.Second)
	// 启动接收方
	go func() {
		for v := range ch {
			fmt.Println(v)
		}
	}()

	time.Sleep(2 * time.Second)
	//close(ch)
	//fmt.Println("111")

}