package main

import (
	"fmt"
	"github.com/link1st/go-stress-testing/helper"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestOpencsx(t *testing.T)  {
	data := helper.OpenCsv("D:/file.csv")
	want := [][] string{
		[]string{"第一个参数"},
		[]string{"第二个参数"},
	}
	if !reflect.DeepEqual(want, data) { // 因为slice不能比较直接，借助反射包中的方法比较
		t.Errorf("excepted:%v, got:%v", want, data) // 测试失败输出错误提示
	}
}
func TestOpencsxDesc(t *testing.T)  {
	data := helper.DescOpenCsv("D:/file.csv")
	want := [][] string{
		[]string{"第一个参数"},
		[]string{"第二个参数"},
	}
	if !reflect.DeepEqual(want, data) { // 因为slice不能比较直接，借助反射包中的方法比较
		t.Errorf("excepted:%v, got:%v", want, data) // 测试失败输出错误提示
	}
}
func TestRandNum(t *testing.T) {
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano())
	// 定义生成随机数的范围
	//min := 1

	// 用于存储随机数的切片
	randomNumbers := make([]int,0,20)
	used := make(map[int]bool) // 用于记录已经生成过的随机数

	// 循环生成指定数量的不重复随机数
	for len(randomNumbers) < 20 {
		// 生成随机数
		num := rand.Intn(100) + 1

		// 如果生成的随机数已经存在于 map 中（即重复），则跳过本次循环
		if used[num] {
			continue
		}

		// 将随机数添加到切片中
		randomNumbers = append(randomNumbers, num)
		used[num] = true // 记录已经使用过的随机数
	}

	// 打印生成的随机数
	fmt.Println(len(randomNumbers))
	fmt.Println(randomNumbers)

	// 将 uint 类型转换为 float64 类型进行计算
	resultFloat := float64(100) * 0.2
	// 将浮点数结果转换为整数
	resultInt := int(resultFloat)
	fmt.Println("resultInt:",resultInt)
}