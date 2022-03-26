package main

import (
	"flag"
	"log"
	"prepare/worker"
	"runtime"
)

var (
	configFile string // 配置文件路径
)

func main() {
	var (
		err error
	)
	//解析命令行参数
	//worker -config ./worker.json -xxx 123 -yyy ddd
	initArgs()

	//配置golang环境
	initEnv()

	//配置解析
	if err = worker.InitConfig(configFile); err != nil {
		goto ERR
	}
	//初始化Registrant
	if err = worker.InitRegistrant(); err != nil {
		goto ERR
	}
	//启动一个协程进行注册
	go worker.Register()

	//初始化LogManager
	if err = worker.InitLogManager(); err != nil {
		goto ERR
	}
	// 启动一个协程来处理日志
	go worker.LoggerLoop()

	//初始化Executor
	if err = worker.InitExecutor(); err != nil {
		goto ERR
	}

	//初始化JobManager
	if err = worker.InitJobManager(); err != nil {
		goto ERR
	}
	//启动协程进行任务和增删监控
	go worker.WatchKill()
	go worker.WatchJobs()

	//初始化Scheduler
	if err = worker.InitScheduler(); err != nil {
		goto ERR
	}
	//启动任务调度循环
	worker.JobSchedulerLoop()
	return
ERR:
	log.Printf("init error : %v", err.Error())
}

//设置和解析命令行参数
func initArgs() {
	flag.StringVar(&configFile, "config", "./worker.json", "配置文件")
	flag.Parse()
}

//初始化golang环境
func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
