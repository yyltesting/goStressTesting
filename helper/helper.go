// Package helper 帮助函数，时间、数组的通用处理
package helper

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"
)

// DiffNano 时间差，纳秒
func DiffNano(startTime time.Time) (diff int64) {
	diff = int64(time.Since(startTime)) //time.Since 计算time.now到since的时间差
	return
}

// InArrayStr 判断字符串是否在数组内
func InArrayStr(str string, arr []string) (inArray bool) {
	for _, s := range arr {
		if s == str {
			inArray = true
			break
		}
	}
	return
}
//生成随机数
func RandNums(max int,count int)(randomNumbers []int){
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano())
	// 定义生成随机数的范围
	//min := 1

	// 用于存储随机数的切片
	randomNumbers = make([]int, 0,count)
	used := make(map[int]bool) // 用于记录已经生成过的随机数

	// 循环生成指定数量的不重复随机数
	for len(randomNumbers) < count {
		// 生成随机数
		num := rand.Intn(max) + 1

		// 如果生成的随机数已经存在于 map 中（即重复），则跳过本次循环
		if used[num] {
			continue
		}

		// 将随机数添加到切片中
		randomNumbers = append(randomNumbers, num)
		used[num] = true // 记录已经使用过的随机数
	}

	return randomNumbers
}
//读取CSV文件信息
func OpenCsv(path string)(value [][]string){
	filebyte ,err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
	}
	filestring := string(filebyte)
	split := strings.Split(filestring, "\r\n")
	var data [][]string

	for i:=0 ;i< len(split);i++{
		d := split[i]
		var a []string
		a = append(a,d)
		data = append(data, a)
	}
	return data
}
//倒叙读取CSV文件信息
func DescOpenCsv(path string)(value [][]string){
	filebyte ,err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
	}
	filestring := string(filebyte)
	split := strings.Split(filestring, "\r\n")
	var data [][]string

	for i := len(split)-1 ;i > 0;i--{
		d := split[i]
		var a []string
		a = append(a,d)
		data = append(data, a)
	}
	return data
}
