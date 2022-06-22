## 背景介绍

⽤ Golang 编写⼀个应⽤，调度和执⾏任务。

## 需求分析

1. 支持任务动态创建；
2. 支持分布式部署；
3. 支持任务结果查询；
4. 任务支持重试；

## 实现概述

- ### 节点分布式架构图

![image](https://github.com/weixiaodong/tanzhi/blob/master/imag/节点分布式架构图.png)

1. 采用master-worker架构，master负责http服务、解析crontab表达式，将到期任务发送到redis队列；多worker通过redis订阅竞争执行到期任务；

2. master和worker节点之间不直接交互，由redis间接通信，例如任务的创建，只需master写入redis，worker通过watch机制捕获更改；

3. 任务执行记录由worker写入mysql，再由master读取

- ### 任务job执行流程

![image](https://github.com/weixiaodong/tanzhi/blob/master/imag/job执行流程.png)

1. crontab 模块，从配置文件、db、http获取crontab配置任务；配置文件、db启动时加载，http创建通过internal/notifych中JobCreateCh通道获取；

2. internal/job/queue 模块，本地队列存储到期任务，可实现为优先级队列，分布式可用redis替换；

3. internal/job/scheduler 模块，任务调度器，轮询获取本地队列中任务，并调用执行器执行，失败后重试，重新将任务放入队列中；分布式部署时要使用分布式锁代替本地锁；

4. internal/job/executor 模块，任务执行器，一种任务类型对应一个执行器执行，扩展任务类型，考虑是否需要增加执行器来执行；

5. internal/notifych 模块，通知模块，http任务创建时写入（非阻塞，channel满记录日志），crontab从通道中读取新建任务。

## 具体实现

- ### 服务内部架构层次依赖

自研框架，简化协议部分，更关注业务逻辑实现。新增服务接口只需要在internal/service中增加handler，然后在api层完成路由注册即可，框架层会自动处理中间件及请求解析。

![image](https://github.com/weixiaodong/tanzhi/blob/master/imag/服务内部架构层次依赖图.png)

1. transport 层负责对外传输协议（http、grpc）;
2. internal/service 层负责服务内部逻辑;
3. internal/job 层负责任务调度执行；
4. api 层负责对外http接口注册；
5. crontab 定时任务层，负责加载定时任务；（master节点）

- ### 代码目录

```sh

├── api                      // api注册
│   └── api.go
├── cmd
│   └── main                 // 服务启动入口
│       └── main.go
├── component                // 公共组件
│   ├── httpclient
│   │   └── client.go
│   └── log
│       ├── ctxutil.go
│       └── log.go
├── config                   // 配置目录
│   ├── configer.go
│   └── config.toml
├── crontab                  // 定时任务启动
│   ├── crontab.go
│   └── handler
│       └── cornjob.go
├── ecode                    // 错误码目录
│   └── ecode.go
├── internal
│   ├── job                  // 任务调度执行目录
│   │   ├── handler.go
│   │   ├── queue
│   │   ├── scheduler
│   │   └── executor
│   ├── notifych             // 任务通知
│   │   └── notifych.go
│   ├── service              // 内部服务逻辑
│   │   └── service.go
│   └── store                // 数据存储
│       └── db
└── transport                // 传输协议目录
    └── http
        ├── endpoint
        ├── innerhandler
        ├── middleware
        ├── server.go
        └── session

```

- ### 实体抽象

1. 任务配置实体 （配置文件、db、http）

    ```go
        type Job struct {
            Name    string        `mapstructure:"name"`
            Expr    string        `mapstructure:"expr"`
            Command CommandConfig `mapstructure:"command"`
        }

        type CommandConfig struct {
            Type   string `mapstructure:"type", json:"type"`
            Method string `mapstructure:"method", json:"method"`
            Target string `mapstructure:"target", json:"target"`
        }
    ```

2. 执行器实体

    ```go
    type Executor interface {
        // 获取任务名称
        GetJobName() string
        // 获取任务类型
        GetJobType() string
        String() string
        // 执行任务
        Execed(context.Context) error
        // 获取失败次数
        GetFailedCnt() uint32
        // 获取重试次数
        GetRetry() uint32
        // 设置任务失败处理
        Failed()
        // 判断任务是否需要重试
        IsRetry() bool
    }
    ```

3. db job配置表实体

    ```go
    type Job struct {
        Id         int       `db:"id" json:"id"`
        Name       string    `db:"name" json:"name"`
        Expr       string    `db:"expr" json:"expr"`
        Command    string    `db:"command" json:"command"`
        CreateTime time.Time `db:"create_time" json:"create_time"`
    }
    ```

4. db job任务记录表实体

    ```go
    type JobRecord struct {
        Id           int       `db:"id" json:"id"`
        Name         string    `db:"name" json:"name"`
        Type         string    `db:"type" json:"type"`
        Command      string    `db:"command" json:"command"`
        Result       string    `db:"result" json:"result"`               // 记录任务结果
        FailedCnt    uint32    `db:"failed_cnt" json:"failed_cnt"`       // 保存任务失败次数
        CreateTime   time.Time `db:"create_time" json:"create_time"`     // 任务创建时间
        StartedTime  time.Time `db:"started_time" json:"started_time"`   // 任务开始执行时间
        FinishedTime time.Time `db:"finished_time" json:"finished_time"` // 任务结束时间
    }
    ```

## 功能完成及测试

任务类型： http / shell命令

- ### 服务启动命令

    ```bash
    go run cmd/main/main.go   # (启动时可用通过添加--race检查竞态)
    ```

    ![image](https://github.com/weixiaodong/tanzhi/blob/master/imag/服务启动.png)

    1. 添加config配置文件中job， 添加db配置表中job
    2. 启动监听服务端口

- ### http 创建定时任务

    ```bash
    curl -vd '{"name":"create-http-job-1", "expr": "0/5 * * * * *","command":{"type":"shell", "target":"echo hello"}}' 'http://127.0.0.1:9000/createJob'
    ```

    ![image](https://github.com/weixiaodong/tanzhi/blob/master/imag/创建http任务执行.png)

    1. log记录http请求参数和响应
    2. crontab从notifych中获取任务（add_httpcreate_job_successed）
    3. 任务到期执行（arrived_job）
    4. 执行器执行（execed_job并保存数据）

- ### http 查询任务执行结果

    1. 获取任务执行列表接口，支持分页；
    2. 根据执行记录id获取执行结果；

    ```bash
    curl -vd '{"page": 0,"size": 1}' 'http://127.0.0.1:9000/listJobRecord'

    curl -d '{"recordId": 44}' 'http://127.0.0.1:9000/getJobRecordResult'
    ```

    ![image](https://github.com/weixiaodong/tanzhi/blob/master/imag/获取任务执行结果.png)

    1. 获取shell执行输出
    2. 获取http响应内容和状态

- ### 重试机制

    调度器会根据任务重试次数配置以及执行结果来判断是否需要重试，如果需要重试则将任务重新放入队列

    shell执行器默认设置重试3次，将config配置中的echo命令改成echo1，测试任务失败重试

    ![image](https://github.com/weixiaodong/tanzhi/blob/master/imag/任务重试.png)

    1. 根据err, failedCnt和retry来控制重试次数（可能需要考虑重试时间间隔）

- ### 分布式任务调度

    master节点将到期任务发送到redis中，worker节点订阅获取任务，然后将执行结果保存到db中

    本项目是将到期任务执行器发送到本地队列中，调度器直接轮询队列获取执行器执行，效率更高。（分布式中将任务信息发送到redis中，调度器根据任务类型调度执行器执行）

- ### 系统可观测

    1. 服务中使用context贯穿上下文，log中根据trace_id贯穿请求整个生命周期。
    2. 可使用 metrics 统计任务失败次数及处理时长

- ### 架构合理，代码简洁

    1. 代码层次划分明确，新增api服务逻辑异常简单，像rpc一样，请求中间件会详细记录请求参数和响应数据；
    2. 新增任务类型可扩展，简单增加任务执行器即可；
    3. 未使用orm，框架自研，真实环境运行过，承担过日活千万，15万qps，单节点过万，未曾出现问题。
