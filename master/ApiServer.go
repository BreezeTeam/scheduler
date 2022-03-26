package master

import (
	"encoding/json"
	"github.com/BreezeTeam/scheduler/common"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

type ApiServer struct {
	server *http.Server
}

var (
	G_apiServer *ApiServer
)

//初始化服务器，并且配置上handler
func InitApiServer() (err error) {
	//初始化apiServer,并且设置单例
	G_apiServer = &ApiServer{}

	//设置路由 绑定到 defaultServerMux
	http.HandleFunc("/job/save", handleJobSave)
	http.HandleFunc("/job/delete", handleJobDelete)
	http.HandleFunc("/job/kill", handleJobKill)

	http.HandleFunc("/job/list", handleJobList)
	http.HandleFunc("/job/log", handleJobLog)
	http.HandleFunc("/worker/list", handleWorkerList)
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(G_config.WebRoot))))

	log.Printf("InitApiServer success")
	return
}

func Start() (err error) {
	var (
		listener net.Listener
		addr     string
	)
	addr = ":" + strconv.Itoa(G_config.ApiPort)
	//启动监听
	if listener, err = net.Listen("tcp", addr); err != nil {
		return
	}
	//创建服务
	G_apiServer.server = &http.Server{
		ReadTimeout:  time.Duration(G_config.ApiReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(G_config.ApiWriteTimeout) * time.Millisecond,
		Handler:      http.DefaultServeMux,
	}
	//启动web协程
	log.Printf("ApiServer Running in %s", addr)
	return G_apiServer.server.Serve(listener)
}
func Stop() {
	G_apiServer.server.Close()
}

/**
curl --location --request GET 'http://127.0.0.1:8070/worker/list'
*/
func handleWorkerList(writer http.ResponseWriter, request *http.Request) {
	var (
		err      error
		workList []string
	)
	//1.workManager获取Work list
	if workList, err = G_workManager.ListWorkers(); err != nil {
		goto ERR
	}

	//2.返回响应
	common.BuildResponse(writer, 200, "success", workList)

	return
ERR:
	common.BuildResponse(writer, 500, err.Error(), nil)
}

/**
curl --location --request GET 'http://127.0.0.1:8070/job/log?job=1&skip=0&limit=20'
*/
func handleJobLog(writer http.ResponseWriter, request *http.Request) {
	var (
		err     error
		logList []*common.JobLog
		skip    int
		limit   int
	)
	//1.获取请求参数
	job := request.URL.Query().Get("job")
	skipParam := request.URL.Query().Get("skip")
	limitParam := request.URL.Query().Get("limit")

	if skip, err = strconv.Atoi(skipParam); err != nil {
		skip = 0
	}
	if limit, err = strconv.Atoi(limitParam); err != nil {
		limit = 20
	}

	if logList, err = G_logManager.LogList(job, skip, limit); err != nil {
		goto ERR
	}

	//2.logManager获取日志
	//3.返回响应

	common.BuildResponse(writer, 200, "success", logList)

	return
ERR:
	common.BuildResponse(writer, 500, err.Error(), nil)
}

/**
curl --location --request GET 'http://127.0.0.1:8070/job/list'
*/
func handleJobList(writer http.ResponseWriter, request *http.Request) {
	var (
		err     error
		jobList []*common.Job
	)
	//1.jobManager获取任务列表
	if jobList, err = G_jobManager.ListJob(common.JOB_SAVE_DIR); err != nil {
		goto ERR
	}

	//2.返回应答
	common.BuildResponse(writer, 200, "success", jobList)

	return
ERR:
	common.BuildResponse(writer, 500, err.Error(), nil)
}

/**
curl --location --request POST 'http://127.0.0.1:8070/job/kill?&job=job1'
*/
func handleJobKill(writer http.ResponseWriter, request *http.Request) {
	var (
		err     error
		leaseId int64
	)
	//1.解析请求参数
	name := request.URL.Query().Get("job")

	//2.jobManager杀死任务
	if leaseId, err = G_jobManager.KillJob(name); err != nil {
		goto ERR
	}

	//3.返回应答
	common.BuildResponse(writer, 200, "success", leaseId)

	return
ERR:
	common.BuildResponse(writer, 500, err.Error(), nil)
}

/**
curl --location --request POST 'http://127.0.0.1:8070/job/delete?&job=job1'
*/
func handleJobDelete(writer http.ResponseWriter, request *http.Request) {

	var (
		err    error
		oldJob *common.Job
	)
	//1.解析请求参数
	name := request.URL.Query().Get("job")

	//2.jobManager删除任务
	if oldJob, err = G_jobManager.DeleteJob(name); err != nil {
		goto ERR
	}

	//3.返回应答
	common.BuildResponse(writer, 200, "success", oldJob)

	return
ERR:
	common.BuildResponse(writer, 500, err.Error(), nil)
}

/**
curl --location --request POST 'http://127.0.0.1:8070/job/save' \
--header 'Content-Type: text/plain' \
--data-raw '{"name": "job1", "command": "echo hello", "cronExpr": "* * * * *"}'
*/
func handleJobSave(writer http.ResponseWriter, request *http.Request) {
	var (
		job    common.Job
		oldJob *common.Job
		err    error
	)
	//1.解析请求参数 并反序列化数据结构, 获取表单数据
	if err = json.NewDecoder(request.Body).Decode(&job); err != nil {
		goto ERR
	}
	//2.jobManager保存数据
	if oldJob, err = G_jobManager.SaveJob(&job); err != nil {
		goto ERR
	}

	//3.返回应答
	common.BuildResponse(writer, 200, "success", oldJob)

	return
ERR:
	common.BuildResponse(writer, 500, err.Error(), nil)
}
