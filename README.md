# go实现的压测工具

- 单台机器对 HTTP 短连接 QPS 1W+ 的压测实战
- 单台机器 100W 长连接的压测实战
- 对 grpc 接口进行压测
- 支持http1.1和2.0长连接
- 支持websocket消息统计
> 简单扩展即可支持 私有协议

## 目录
- [1、项目说明](#1项目说明)
    - [1.1 go-stress-testing](#11-go-stress-testing)
    - [1.2 项目体验](#12-项目体验)
- [2、压测](#2压测)
    - [2.1 压测是什么](#21-压测是什么)
    - [2.2 为什么要压测](#22-为什么要压测)
    - [2.3 压测名词解释](#23-压测名词解释)
        - [2.3.1 压测类型解释](#231-压测类型解释)
        - [2.3.2 压测名词解释](#232-压测名词解释)
        - [2.3.3 机器性能指标解释](#233-机器性能指标解释)
        - [2.3.4 访问指标解释](#234-访问指标解释)
    - [3.4 如何计算压测指标](#24-如何计算压测指标)
- [3、go-stress-testing go语言实现的压测工具](#4go-stress-testing-go语言实现的压测工具)
    - [4.1 介绍](#41-介绍)
    - [4.2 用法](#42-用法)
    - [4.3 实现](#43-实现)
    - [4.4 go-stress-testing 对 Golang web 压测](#44-go-stress-testing-对-golang-web-压测)
    - [4.5 grpc压测](#45-grpc压测)
- [4、单台机器100w连接压测实战](#6单台机器100w连接压测实战)
    - [6.1 说明](#61-说明)
    - [6.2 内核优化](#62-内核优化)
    - [6.3 客户端配置](#63-客户端配置)
    - [6.4 准备](#64-准备)
    - [6.5 压测数据](#65-压测数据)
- [5、常见问题](#7常见问题)



## 1、项目说明
### 1.1 go-stress-testing

go 实现的压测工具，每个用户用一个协程的方式模拟，最大限度的利用 CPU 资源

### 1.2 项目体验

- 可以在 mac/linux/windows 不同平台下执行的命令

- [go-stress-testing](https://github.com/link1st/go-stress-testing/releases) 源码压测工具下载地址

若在linux服务器上进行压测，可进行以下步骤：
#### 1、配置编译环境变量 go env 查看环境变量，配置goos为Linux set GOOS=linux
#### 2、打包项目文件为二进制文件 go build
#### 3、在linux服务器上导入打包好的文件 rz
#### 4、修改运行文件权限 chmod 777 main
#### 5、命令运行main文件 ./main -c 10000 -n 1 -sender 1 -clients 0 -v json -u wss://dev-gateway-coppy.wss1.cn?t=2UubEiP4RiqY3GFqSWwT4vlD1lX

参数说明:

`-c` 表示并发数

`-n` 每个并发执行请求的次数，总请求的次数 = 并发数 `*` 每个并发执行请求的次数

`-u` 需要压测的地址

```shell

# 运行 以mac为示例
./go-stress-testing-mac -c 1 -n 100 -u https://www.baidu.com/

```

- 压测结果展示

执行以后，终端每秒钟都会输出一次结果，压测完成以后输出执行的压测结果

压测结果展示:

```

─────┬───────┬───────┬───────┬────────┬────────┬────────┬────────┬────────
 耗时│ 并发数 │ 成功数│ 失败数 │   qps  │最长耗时 │最短耗时│平均耗时 │ 错误码
─────┼───────┼───────┼───────┼────────┼────────┼────────┼────────┼────────
   1s│      1│      8│      0│    8.09│  133.16│  110.98│  123.56│200:8
   2s│      1│     15│      0│    8.02│  138.74│  110.98│  124.61│200:15
   3s│      1│     23│      0│    7.80│  220.43│  110.98│  128.18│200:23
   4s│      1│     31│      0│    7.83│  220.43│  110.23│  127.67│200:31
   5s│      1│     39│      0│    7.81│  220.43│  110.23│  128.03│200:39
   6s│      1│     46│      0│    7.72│  220.43│  110.23│  129.59│200:46
   7s│      1│     54│      0│    7.79│  220.43│  110.23│  128.42│200:54
   8s│      1│     62│      0│    7.81│  220.43│  110.23│  128.09│200:62
   9s│      1│     70│      0│    7.79│  220.43│  110.23│  128.33│200:70
  10s│      1│     78│      0│    7.82│  220.43│  106.47│  127.85│200:78
  11s│      1│     84│      0│    7.64│  371.02│  106.47│  130.96│200:84
  12s│      1│     91│      0│    7.63│  371.02│  106.47│  131.02│200:91
  13s│      1│     99│      0│    7.66│  371.02│  106.47│  130.54│200:99
  13s│      1│    100│      0│    7.66│  371.02│  106.47│  130.52│200:100


*************************  结果 stat  ****************************
处理协程数量: 1
请求总数: 100 总请求时间: 13.055 秒 successNum: 100 failureNum: 0
*************************  结果 end   ****************************

```

参数解释:

**耗时**: 程序运行耗时。程序每秒钟输出一次压测结果

**并发数**: 并发数，启动的协程数

**成功数**: 压测中，请求成功的数量

**失败数**: 压测中，请求失败的数量

**qps**: 当前压测的QPS(每秒钟处理请求数量)

**最长耗时**: 压测中，单个请求最长的响应时长

**最短耗时**: 压测中，单个请求最短的响应时长

**平均耗时**: 压测中，单个请求平均的响应时长

**错误码**: 压测中，接口返回的 code码:返回次数的集合

## 2、压测
### 2.1 压测是什么

压测，即压力测试，是确立系统稳定性的一种测试方法，通常在系统正常运作范围之外进行，以考察其功能极限和隐患。

主要检测服务器的承受能力，包括用户承受能力（多少用户同时玩基本不影响质量）、流量承受等。

### 2.2 为什么要压测

- 压测的目的就是通过压测(模拟真实用户的行为)，测算出机器的性能(单台机器的QPS)，从而推算出系统在承受指定用户数(100W)时，需要多少机器能支撑得住
- 压测是在上线前为了应对未来可能达到的用户数量的一次预估(提前演练)，压测以后通过优化程序的性能或准备充足的机器，来保证用户的体验。

### 2.3 压测名词解释
#### 2.3.1 压测类型解释

| 压测类型 |   解释  |
| :----   | :---- |
| 压力测试(Stress Testing)          |  也称之为强度测试，测试一个系统的最大抗压能力，在强负载(大数据、高并发)的情况下，测试系统所能承受的最大压力，预估系统的瓶颈    |
| 并发测试(Concurrency Testing)     |  通过模拟很多用户同一时刻访问系统或对系统某一个功能进行操作，来测试系统的性能，从中发现问题(并发读写、线程控制、资源争抢)      |
| 耐久性测试(Configuration Testing) |  通过对系统在大负荷的条件下长时间运行，测试系统、机器的长时间运行下的状况,从中发现问题(内存泄漏、数据库连接池不释放、资源不回收)     |


#### 2.3.2 压测名词解释

| 压测名词 |   解释  |
| :----   | :---- |
| 并发(Concurrency)     |  指一个处理器同时处理多个任务的能力(逻辑上处理的能力)     |
| 并行(Parallel)        |  多个处理器或者是多核的处理器同时处理多个不同的任务(物理上同时执行)     |
| QPS(每秒钟查询数量 Query Per Second) | 服务器每秒钟处理请求数量  （并发量/平均响应时间）  |
| 事务(Transactions) | 是用户一次或者是几次请求的集合    |
| TPS(每秒钟处理事务数量 Transaction Per Second) | 服务器每秒钟处理事务数量(一个事务可能包括多个请求) (req/sec  请求数/秒  一段时间内总请求数/总请求时间)    |
| 请求成功数(Request Success Number) | 在一次压测中，请求成功的数量    |
| 请求失败数(Request Failures Number) | 在一次压测中，请求失败的数量    |
| 错误率(Error Rate) | 在压测中，请求成功的数量与请求失败数量的比率  |
| 最大响应时间(Max Response Time) | 在一次压测中，从发出请求或指令系统做出的反映(响应)的最大时间  |
| 最少响应时间(Mininum Response Time) | 在一次压测中，从发出请求或指令系统做出的反映(响应)的最少时间  |
| 平均响应时间(Average Response Time) | 在一次压测中，从发出请求或指令系统做出的反映(响应)的平均时间  |

#### 2.3.3 机器性能指标解释

| 机器性能 |   解释  |
| :----   | :---- |
| CUP利用率(CPU Usage)       |  CUP 利用率分用户态、系统态和空闲态，CPU利用率是指:CPU执行非系统空闲进程的时间与CPU总执行时间的比率      |
| 内存使用率(Memory usage)    |  内存使用率指的是此进程所开销的内存。      |
| IO(Disk input/ output)    |  磁盘的读写包速率       |
| 网卡负载(Network Load)      |  网卡的进出带宽,包量       |

#### 2.3.4 访问指标解释

| 访问 |   解释  |
| :----   | :---- |
| PV(页面浏览量 Page View)           |  用户每打开1个网站页面，记录1个PV。用户多次打开同一页面，PV值累计多次      |
| UV(网站独立访客 Unique Visitor)    |  通过互联网访问、流量网站的自然人。1天内相同访客多次访问网站，只计算为1个独立访客       |

### 2.4 如何计算压测指标

- 压测我们需要有目的性的压测，这次压测我们需要达到什么目标(如:单台机器的性能为 100QPS?网站能同时满足100W人同时在线)
- 可以通过以下计算方法来进行计算:
- 压测原则:每天80%的访问量集中在20%的时间里，这20%的时间就叫做峰值
- 公式: ( 总PV数`*`80% ) / ( 每天的秒数`*`20% ) = 峰值时间每秒钟请求数(QPS)
- 机器: 峰值时间每秒钟请求数(QPS) / 单台机器的QPS = 需要的机器的数量

- 假设:网站每天的用户数(100W)，每天的用户的访问量约为3000W PV，这台机器的需要多少QPS?
> ( 30000000\*0.8 ) / (86400 * 0.2) ≈ 1389 (QPS)

- 假设:单台机器的的QPS是69，需要需要多少台机器来支撑？
> 1389 / 69 ≈ 20

## 3、go-stress-testing go语言实现的压测工具

### 3.1 介绍

- go-stress-testing 是go语言实现的简单压测工具，源码开源、支持二次开发，可以压测http、webSocket请求、私有rpc调用，使用协程模拟单个用户，可以更高效的利用CPU资源。

- 项目源码地址 [https://github.com/link1st/go-stress-testing](https://github.com/link1st/go-stress-testing)

### 3.2 用法

- [go-stress-testing](https://github.com/link1st/go-stress-testing/releases) 源码下载地址
- clone 项目源码运行的时候，需要将项目 clone 到 **$GOPATH** 目录下
- 支持参数:

```
Usage of ./go-stress-testing-mac:
  -c uint
      并发数 (default 1)
  -n uint
      请求数(单个并发/协程) (default 1)
  -t uint64
      请求执行时间（循环请求） min
  -s uint
      阶梯压测每秒启动数，注意需要和并发数整除
  -i uint
      阶梯压测时间启动间隔S
  -u string
      压测地址
  -d string
      调试模式 (default "false")
  -http2 false
      是否开http2.0
  -k false
      是否开启长连接 若是长连接，请求后会等待所有协程跑完后断开
  -m int
    	单个host最大连接数 (default 1)
  -H value
      自定义头信息传递给服务器 示例:-H 'Content-Type: application/json'
  -data string
      HTTP POST方式传送数据
  -v string
      验证方法 http 支持:statusCode、json webSocket支持:json
  -p string
      curl文件路径
  -client int
      是否多个客户端,1一个、2多个、0不设置则为单接口
  -issenders fasle
      是否多个发送端
  -iscollectmsg false
      是否将消息收集起来统计
  -toclients fasle
      是否一个发送端向多个群或人（客户端）发送消息，默认只发送一次
  -apptimeout int64
      程序超时时间，默认不设置
  -timeout uint64
      接口超时时间，默认60秒
  -redirect true
      是否重定向
  -connectionmode 1
      建立连接方式 1:顺序建立长链接 2:并发建立长链接
  -variable string
      项目定制化变量，变量通过,号分割 分别为sendtoken,sendid,recvtoken,recvid
```

- `-n` 是单个用户请求的次数，请求总次数 = `-c`* `-n`， 这里考虑的是模拟用户行为，所以这个是每个用户请求的次数

- 下载以后执行下面命令即可压测

- 使用示例:

```
# 查看用法
./go-stress-testing-mac

# 使用请求百度页面
./go-stress-testing-mac -c 1 -n 100 -u https://www.baidu.com/

# 使用debug模式请求百度页面
./go-stress-testing-mac -c 1 -n 1 -d true -u https://www.baidu.com/

# 使用 curl文件(文件在curl目录下) 的方式请求
./go-stress-testing-mac -c 1 -n 1 -p curl/baidu.curl.txt

# 压测webSocket连接
./go-stress-testing-mac -c 10 -n 10 -u ws://127.0.0.1:8089/acc

一个发送端，没有客户端 进行发送压测
-c 100 -n 1 -issenders false -clients 0 -v -variable json -u wss://dev-gateway-coppy.wss1.cn/?t=2V6v2toRktC4AS0zGCARHQzAWh7

一个发送端，一个客户端接受统计 进行发送压测
-c 100 -n 1 -issenders false -clients 1 -iscollectmsg true -v json -u wss://dev-gateway-coppy.wss1.cn/?t=2V6v2toRktC4AS0zGCARHQzAWh7

多个发送端，一个客户端接受统计 进行发送压测
-c 100 -n 1 -issenders true -clients 1 -iscollectmsg true -v json -u wss://dev-gateway-coppy.wss1.cn/?t=2V6v2toRktC4AS0zGCARHQzAWh7

1个发送端，没有客户端接受统计，持续2分钟 进行发送压测
-c 100 -n 1 -issenders false -clients 0 -t 2 -v json -u wss://dev-gateway-coppy.wss1.cn/?t=2V6v2toRktC4AS0zGCARHQzAWh7

多个发送端，多个客户端接受统计，私聊一对一 进行发送压测
-c 100 -n 1 -issenders true -clients 2 -v json -u wss://dev-gateway-coppy.wss1.cn/?t=2V6v2toRktC4AS0zGCARHQzAWh7

多个发送端，多个客户端接受统计，群聊 进行发送压测
-c 100 -n 1 -issender true -clients 2 -iscollectmsg true -v json -u wss://dev-gateway-coppy.wss1.cn/?t=2V6v2toRktC4AS0zGCARHQzAWh7

多个发送端，多个客户端接受统计，且每个发送端往多个群发送消息 进行发送压测
-c 100 -n 1 -issenders true -clients 2 -iscollectmsg true -toclients true -variable sendtoken.csv,sendtoken.csv,recvtoken.csv,groupId.csv -v json -u wss://dev-gateway-coppy.wss1.cn/?t=2V6v2toRktC4AS0zGCARHQzAWh7
```

- 完整压测命令示例
```shell script
# 更多参数 支持 header、post body
go run main.go -c 1 -n 1 -d true -u 'https://page.aliyun.com/delivery/plan/list' \
  -H 'authority: page.aliyun.com' \
  -H 'accept: application/json, text/plain, */*' \
  -H 'content-type: application/x-www-form-urlencoded' \
  -H 'origin: https://cn.aliyun.com' \
  -H 'sec-fetch-site: same-site' \
  -H 'sec-fetch-mode: cors' \
  -H 'sec-fetch-dest: empty' \
  -H 'referer: https://cn.aliyun.com/' \
  -H 'accept-language: zh-CN,zh;q=0.9' \
  -H 'cookie: aliyun_choice=CN; JSESSIONID=J8866281-CKCFJ4BUZ7GDO9V89YBW1-KJ3J5V9K-GYUW7; maliyun_temporary_console0=1AbLByOMHeZe3G41KYd5WWZvrM%2BGErkaLcWfBbgveKA9ifboArprPASvFUUfhwHtt44qsDwVqMk8Wkdr1F5LccYk2mPCZJiXb0q%2Bllj5u3SQGQurtyPqnG489y%2FkoA%2FEvOwsXJTvXTFQPK%2BGJD4FJg%3D%3D; cna=L3Q5F8cHDGgCAXL3r8fEZtdU; isg=BFNThsmSCcgX-sUcc5Jo2s2T4tF9COfKYi8g9wVwr3KphHMmjdh3GrHFvPTqJD_C; l=eBaceXLnQGBjstRJBOfwPurza77OSIRAguPzaNbMiT5POw1B5WAlWZbqyNY6C3GVh6lwR37EODnaBeYBc3K-nxvOu9eFfGMmn' \
  -data 'adPlanQueryParam=%7B%22adZone%22%3A%7B%22positionList%22%3A%5B%7B%22positionId%22%3A83%7D%5D%7D%2C%22requestId%22%3A%2217958651-f205-44c7-ad5d-f8af92a6217a%22%7D'
```

- 使用 curl文件进行压测

curl是Linux在命令行下的工作的文件传输工具，是一款很强大的http命令行工具。

使用curl文件可以压测使用非GET的请求，支持设置http请求的 method、cookies、header、body等参数


**I:** chrome 浏览器生成 curl文件，打开开发者模式(快捷键F12)，如图所示，生成 curl 在终端执行命令
![chrome cURL](http://img.91vh.com/img/copy%20cURL.png)

**II:** postman 生成 curl 命令
![postman cURL](http://img.91vh.com/img/postman%20cURL.png)

生成内容粘贴到项目目录下的**curl/baidu.curl.txt**文件中，执行下面命令就可以从curl.txt文件中读取需要压测的内容进行压测了

```
# 使用 curl文件(文件在curl目录下) 的方式请求
go run main.go -c 1 -n 1 -p curl/baidu.curl.txt
```


### 3.3 实现

- 具体需求可以查看项目源码

- 项目目录结构

```
|____main.go                      // main函数，获取命令行参数
|____conf                        // 配置
| |____conf.go                 // 项目配置参数以及可变参数读取
|____server                       // 处理程序目录
| |____dispose.go                 // 压测启动，注册验证器、启动统计函数、启动协程进行压测
| |____statistics                 // 统计目录
| | |____statistics.go            // 接收压测统计结果并处理
| |____golink                     // 建立连接目录
| | |____http_link.go             // http建立连接
| | |____websocket_link.go        // webSocket建立连接
| |____client                     // 请求数据客户端目录
| | |____http_client.go           // http客户端
| | |____websocket_client.go      // webSocket客户端
| |____verify                     // 对返回数据校验目录
| | |____http_verify.go           // http返回数据校验
| | |____websokcet_verify.go      // webSocket返回数据校验
| |____receive                     // websocket客户端接收器
| | |____websocket_recv.go         // websocket客户端接收消息接收器
|____heper                        // 通用函数目录
| |____heper.go                   // 通用函数
|____model                        // 模型目录
| |____request_model.go           // 请求数据模型
| |____curl_model.go              // curl文件解析
|____curl                        // curl文件集
|____csv                        // csv参数文件集
|____vendor                       // 项目依赖目录
```


### 3.4 go-stress-testing 对 Golang web 压测


这里使用go-stress-testing对go server进行压测(部署在同一台机器上)，并统计压测结果

- 申请的服务器配置

CPU: 4核 (Intel Xeon(Cascade Lake) Platinum 8269  2.5 GHz/3.2 GHz)

内存: 16G
硬盘: 20G SSD
系统: CentOS 7.6

go version: go1.12.9 linux/amd64

![go-stress-testing01](http://img.91vh.com/img/go-stress-testing01.png)

- go server

```golangpackage main

import (
    "log"
    "net/http"
    "runtime"
)

const (
    httpPort = "8088"
)

func main() {

    runtime.GOMAXPROCS(runtime.NumCPU() - 1)

    hello := func(w http.ResponseWriter, req *http.Request) {
        data := "Hello, go-stress-testing! \n"

        w.Header().Add("Server", "golang")
        w.Write([]byte(data))

        return
    }

    http.HandleFunc("/", hello)
    err := http.ListenAndServe(":"+httpPort, nil)

    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}

```

- go_stress_testing 压测命令

```
./go-stress-testing-linux -c 100 -n 10000 -u http://127.0.0.1:8088/
```


- 压测结果
- [压测结果 示例](https://github.com/link1st/go-stress-testing/issues/32)

| 并发数  |  go_stress_testing QPS  |
| :----: |  :----:  |
|   1    | 6394.86  |
|   4    | 16909.36 |
|   10   | 18456.81 |
|   20   | 19490.50 |
|   30   | 19947.47 |
|   50   | 19922.56 |
|   80   | 19155.33 |
|   100  | 18336.46 |

从压测的结果上看：效果还不错，压测QPS有接近2W

### 3.5 grpc压测
- 介绍如何压测 grpc 接口
> [添加对 grpc 接口压测 commit](https://github.com/link1st/go-stress-testing/commit/2b4b14aaf026d08276531cf76f42de90efd3bc61)
- 1. 启动Server
```shell script
# 进入 grpc server 目录
cd tests/grpc

# 启动 grpc server
go run main.go
```

- 2. 对 grpc server 协议进行压测
```
# 回到项目根目录
go run main.go -c 300 -n 1000 -u grpc://127.0.0.1:8099 -data world

开始启动  并发数:300 请求数:1000 请求参数:
request:
 form:grpc
 url:grpc://127.0.0.1:8099
 method:POST
 headers:map[Content-Type:application/x-www-form-urlencoded; charset=utf-8]
 data:world
 verify:
 timeout:30s
 debug:false

─────┬───────┬───────┬───────┬────────┬────────┬────────┬────────┬────────┬────────┬────────
 耗时 │ 并发数 │ 成功数 │ 失败数 │   qps  │最长耗时  │最短耗时 │平均耗时  │下载字节 │字节每秒  │ 错误码
─────┼───────┼───────┼───────┼────────┼────────┼────────┼────────┼────────┼────────┼────────
   1s│    186│  14086│      0│34177.69│   22.40│    0.63│    8.78│        │        │200:14086
   2s│    265│  30408│      0│26005.09│   32.68│    0.63│   11.54│        │        │200:30408
   3s│    300│  46747│      0│21890.46│   40.84│    0.63│   13.70│        │        │200:46747
   4s│    300│  62837│      0│20057.06│   45.81│    0.63│   14.96│        │        │200:62837
   5s│    300│  79119│      0│19134.52│   45.81│    0.63│   15.68│        │        │200:79119
```

- 如何扩展其它私有协议
> 由于私有协议、grpc 协议 都涉及到代码的书写，所以需要 编写go 的代码才能完成
> 参考 [添加对 grpc 接口压测 commit](https://github.com/link1st/go-stress-testing/commit/2b4b14aaf026d08276531cf76f42de90efd3bc61)


## 4、单台机器100w连接压测实战
### 4.1 说明

之前写了一篇文章，[基于websocket单台机器支持百万连接分布式聊天(IM)系统](https://github.com/link1st/gowebsocket)(不了解这个项目可以查看上一篇或搜索一下文章)，这里我们要实现单台机器支持100W连接的压测

目标:

* 单台机器能保持100W个长连接
* 机器的CPU、内存、网络、I/O 状态都正常

说明:

gowebsocket 分布式聊天(IM)系统:

* 之前用户连接以后有个全员广播，这里需要将用户连接、退出等事件关闭


- 服务器准备:
> 由于自己手上没有自己的服务器，所以需要临时购买的云服务器

压测服务器:

16台(稍后解释为什么需要16台机器)

CPU: 2核
内存: 8G
硬盘: 20G
系统: CentOS 7.6


![webSocket压测服务器](http://img.91vh.com/img/webSocket%E5%8E%8B%E6%B5%8B%E6%9C%8D%E5%8A%A1%E5%99%A8.png)

被压测服务:

1台

CPU: 4核
内存: 32G
硬盘: 20G SSD
系统: CentOS 7.6

![webSocket被压测服务器](http://img.91vh.com/img/webSocket%E8%A2%AB%E5%8E%8B%E6%B5%8B%E6%9C%8D%E5%8A%A1%E5%99%A8.png)


### 4.2 内核优化

- 修改程序最大打开文件数

被压测服务器需要保持100W长连接，客户和服务器端是通过socket通讯的，每个连接需要建立一个socket，程序需要保持100W长连接就需要单个程序能打开100W个文件句柄


```
# 查看系统默认的值
ulimit -n
# 设置最大打开文件数
ulimit -n 1040000
```

这里设置的要超过100W，程序除了有100W连接还有其它资源连接(数据库、资源等连接)，这里设置为 104W

centOS 7.6 上述设置不生效，需要手动修改配置文件

`vim /etc/security/limits.conf`

这里需要把硬限制和软限制、root用户和所有用户都设置为 1040000

core 是限制内核文件的大小，这里设置为 unlimited

```
# 添加以下参数
root soft nofile 1040000
root hard nofile 1040000
root soft nproc 1040000
root hard nproc 1040000

* soft nofile 1040000
* hard nofile 1040000
* soft nproc 1040000
* hard nproc 1040000

root soft core unlimited
root hard core unlimited

* soft core unlimited
* hard core unlimited
```

注意:

`/proc/sys/fs/file-max` 表示系统级别的能够打开的文件句柄的数量，不能小于limits中设置的值

如果file-max的值小于limits设置的值会导致系统重启以后无法登录

```
# file-max 设置的值参考
cat /proc/sys/fs/file-max
12553500
```

修改以后重启服务器，`ulimit -n` 查看配置是否生效


### 4.3 客户端配置

由于linux端口的范围是 `0~65535(2^16-1)`这个和操作系统无关，不管linux是32位的还是64位的

这个数字是由于tcp协议决定的，tcp协议头部表示端口只有16位，所以最大值只有65535(如果每台机器多几个虚拟ip就能突破这个限制)

1024以下是系统保留端口，所以能使用的1024到65535

如果需要100W长连接，每台机器有 65535-1024 个端口， 100W / (65535-1024) ≈ 15.5，所以这里需要16台服务器

- `vim /etc/sysctl.conf` 在文件末尾添加

```
net.ipv4.ip_local_port_range = 1024 65000
net.ipv4.tcp_mem = 786432 2097152 3145728
net.ipv4.tcp_rmem = 4096 4096 16777216
net.ipv4.tcp_wmem = 4096 4096 16777216
```

`sysctl -p` 修改配置以后使得配置生效命令

配置解释:

- `ip_local_port_range` 表示TCP/UDP协议允许使用的本地端口号 范围:1024~65000
- `tcp_mem` 确定TCP栈应该如何反映内存使用，每个值的单位都是内存页（通常是4KB）。第一个值是内存使用的下限；第二个值是内存压力模式开始对缓冲区使用应用压力的上限；第三个值是内存使用的上限。在这个层次上可以将报文丢弃，从而减少对内存的使用。对于较大的BDP可以增大这些值（注意，其单位是内存页而不是字节）
- `tcp_rmem` 为自动调优定义socket使用的内存。第一个值是为socket接收缓冲区分配的最少字节数；第二个值是默认值（该值会被rmem_default覆盖），缓冲区在系统负载不重的情况下可以增长到这个值；第三个值是接收缓冲区空间的最大字节数（该值会被rmem_max覆盖）。
- `tcp_wmem` 为自动调优定义socket使用的内存。第一个值是为socket发送缓冲区分配的最少字节数；第二个值是默认值（该值会被wmem_default覆盖），缓冲区在系统负载不重的情况下可以增长到这个值；第三个值是发送缓冲区空间的最大字节数（该值会被wmem_max覆盖）。

### 4.4 准备


1. 在被压测服务器上启动Server服务(gowebsocket)

2. 查看被压测服务器的内网端口

3. 登录上16台压测服务器，这里我提前把需要优化的系统做成了镜像，申请机器的时候就可以直接使用这个镜像(参数已经调好)

![压测服务器16台准备](http://img.91vh.com/img/%E5%8E%8B%E6%B5%8B%E6%9C%8D%E5%8A%A1%E5%99%A816%E5%8F%B0%E5%87%86%E5%A4%87.png)

4. 启动压测

```
 ./go_stress_testing_linux -c 62500 -n 1  -u ws://192.168.0.74:443/acc
```

`62500*16 = 100W `正好可以达到我们的要求

建立连接以后，`-n 1`发送一个**ping**的消息给服务器，收到响应以后保持连接不中断

5. 通过 gowebsocket服务器的http接口，实时查询连接数和项目启动的协程数

6. 压测过程中查看系统状态

```
# linux 命令
ps      # 查看进程内存、cup使用情况
iostat  # 查看系统IO情况
nload   # 查看网络流量情况
/proc/pid/status # 查看进程状态
```

### 4.5 压测数据

- 压测以后，查看连接数到100W，然后保持10分钟观察系统是否正常

- 观察以后，系统运行正常、CPU、内存、I/O 都正常，打开页面都正常

- 压测完成以后的数据

查看goWebSocket连接数统计，可以看到 **clientsLen**连接数为100W，**goroutine**数量2000008个，每个连接两个goroutine加上项目启动默认的8个。这里可以看到连接数满足了100W

![查看goWebSocket连接数统计](http://img.91vh.com/img/%E6%9F%A5%E7%9C%8BgoWebSocket%E8%BF%9E%E6%8E%A5%E6%95%B0%E7%BB%9F%E8%AE%A1.png)

从压测服务上查看连接数是否达到了要求，压测完成的统计数据并发数为62500，是每个客户端连接的数量,总连接数： `62500*16=100W`，

![压测服务16台 压测完成](http://img.91vh.com/img/%E5%8E%8B%E6%B5%8B%E6%9C%8D%E5%8A%A116%E5%8F%B0%20%E5%8E%8B%E6%B5%8B%E5%AE%8C%E6%88%90.png)

- 记录内存使用情况，分别记录了1W到100W连接数内存使用情况

| 连接数      |  内存 |
| :----:     | :----:|
|   10000    | 281M  |
|   100000   | 2.7g  |
|   200000   | 5.4g  |
|   500000   | 13.1g |
|   1000000  | 25.8g |


100W连接时的查看内存详细数据:

```
cat /proc/pid/status
VmSize: 27133804 kB
```

`27133804/1000000≈27.1` 100W连接，占用了25.8g的内存，粗略计算了一下，一个连接占用了27.1Kb的内存，由于goWebSocket项目每个用户连接起了两个协程处理用户的读写事件，所以内存占用稍微多一点

如果需要如何减少内存使用可以参考 **@Roy11568780** 大佬给的解决方案
> 传统的golang中是采用的一个goroutine循环read的方法对应每一个socket。实际百万链路场景中这是巨大的资源浪费，优化的原理也不是什么新东西，golang中一样也可以使用epoll的，把fd拿到epoll中，检测到事件然后在协程池里面去读就行了，看情况读写分别10-20的协程goroutine池应该就足够了

至此，压测已经全部完成，单台机器支持100W连接已经满足~

## 5.常见问题
- **Q:** 压测过程中会出现大量 **TIME_WAIT**

 A: 参考TCP四次挥手原理，主动关闭连接的一方会出现 **TIME_WAIT** 状态，等待的时长为 2MSL(约1分钟左右)

 原因是：主动断开的一方回复 ACK 消息可能丢失，TCP 是可靠的传输协议，在没有收到 ACK 消息的另一端会重试，重新发送FIN消息，所以主动关闭的一方会等待 2MSL 时间，防止对方重试，这就出现了大量 **TIME_WAIT** 状态（参考: 四次挥手的最后两次）

TCP 握手：
<img border="0" src="http://img.91vh.com/img/TCP%E4%B8%89%E6%AC%A1%E6%8F%A1%E6%89%8B%E3%80%81%E5%9B%9B%E6%AC%A1%E6%8C%A5%E6%89%8B.png" width="830"/>

