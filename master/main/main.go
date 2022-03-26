package main

import (
	"flag"
	"log"
	"prepare/master"
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
	//master -config ./master.json -xxx 123 -yyy ddd
	initArgs()

	//配置golang环境
	initEnv()

	//配置解析
	if err = master.InitConfig(configFile); err != nil {
		goto ERR
	}
	//初始化JobManager
	if err = master.InitJobManager(); err != nil {
		goto ERR
	}
	//初始化WorkManage
	if err = master.InitWorkManager(); err != nil {
		goto ERR
	}
	//初始化LogManager
	if err = master.InitLogManager(); err != nil {
		goto ERR
	}
	//初始化ApiServer
	if err = master.InitApiServer(); err != nil {
		goto ERR
	}

	//启动 ApiServer
	if err = master.Start(); err != nil {
		goto ERR
	}
	return
ERR:
	log.Printf("init error : %v", err.Error())
}

//设置和解析命令行参数
func initArgs() {
	flag.StringVar(&configFile, "config", "./master.json", "配置文件")
	flag.Parse()
}

//初始化golang环境
func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
