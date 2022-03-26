package worker

import (
	"github.com/BreezeTeam/scheduler/common"
	"log"
	"time"
)

type Scheduler struct {
	jobEventChan         chan *common.JobEvent               //job 事件队列
	jobExecuteResultChan chan *common.JobExecuteResult       //job 执行结果队列
	jobPlanTable         map[string]*common.JobSchedulePlan  //任务计划表
	jobExecutingTable    map[string]*common.JobExecuteStatus //任务执行表
}

var (
	G_scheduler *Scheduler
)

//初始化
func InitScheduler() (err error) {
	G_scheduler = &Scheduler{
		jobEventChan:         make(chan *common.JobEvent, 1000),
		jobExecuteResultChan: make(chan *common.JobExecuteResult, 1000),
		jobPlanTable:         make(map[string]*common.JobSchedulePlan),
		jobExecutingTable:    make(map[string]*common.JobExecuteStatus),
	}
	log.Printf("InitScheduler success")
	return
}

//事件处理循环
func JobSchedulerLoop() {
	var (
		event  *common.JobEvent
		result *common.JobExecuteResult
	)
	//1.初始化任务表,此时plan table 为 空，睡1s
	nearSchedulerTime := G_scheduler.updateJobPlanTable()
	nearTimer := time.NewTimer(nearSchedulerTime)
	for {
		select {
		case event = <-G_scheduler.jobEventChan:
			G_scheduler.handJobEvent(event)
		case result = <-G_scheduler.jobExecuteResultChan:
			G_scheduler.handJobResult(result)
		case <-nearTimer.C:
		}

		//时间到期,更新任务计划表，并且重置时间
		nearSchedulerTime = G_scheduler.updateJobPlanTable()
		nearTimer.Reset(nearSchedulerTime)
		if G_config.Debug {
			log.Printf("Scheduler::JobSchedulerLoop SUCCESS")
		}
	}

}

//jobManager 将会调用这个函数进行event推送
func (scheduler *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	scheduler.jobEventChan <- jobEvent

}

//Exeuter将会调用这个函数进行执行结果的推送
func (scheduler *Scheduler) PushJobExecuteResult(result *common.JobExecuteResult) {
	scheduler.jobExecuteResultChan <- result
}

//job 执行结果，当executer执行完job后，会向 chan *common.JobExecuteResult 中写入jobresult
func (scheduler *Scheduler) handJobResult(result *common.JobExecuteResult) {
	//1. 从执行表中删除job
	delete(scheduler.jobExecutingTable, result.ExecuteInfo.Job.Name)
	//2. 生成日志并发送日志到日志管理器
	if result.Err != common.ERR_LOCK_ALREADY_REQUIRED {
		jobLog := &common.JobLog{
			JobName:      result.ExecuteInfo.Job.Name,
			Command:      result.ExecuteInfo.Job.Command,
			Output:       string(result.Output),
			PlanTime:     result.ExecuteInfo.PlanTime.UnixNano() / 1000 / 1000,
			ScheduleTime: result.ExecuteInfo.RealTime.UnixNano() / 1000 / 1000,
			StartTime:    result.StartTime.UnixNano() / 1000 / 1000,
			EndTime:      result.EndTime.UnixNano() / 1000 / 1000,
		}
		if result.Err != nil {
			jobLog.Err = result.Err.Error()
		} else {
			jobLog.Err = ""
		}

		G_logManager.SendLog(jobLog)
		log.Printf("Scheduler::SendLog queue success,job: %s ,startTime: %d", jobLog.JobName, jobLog.StartTime)
	}

}

//job Event 处理函数
func (scheduler *Scheduler) handJobEvent(event *common.JobEvent) {
	switch event.EventType {
	case common.JOB_EVENT_SAVE:
		//将该job添加/覆盖到任务计划表中
		if schedulePlan, err := common.BuildJobSchedulePlan(event.Job); err == nil {
			scheduler.jobPlanTable[event.Job.Name] = schedulePlan
		}
		log.Printf("Scheduler::handJobEvent,event: JOB_EVENT_SAVE ,job: %s", event.Job.Name)
	case common.JOB_EVENT_DELETE:
		//从调度表中删除该job
		if _, ok := scheduler.jobPlanTable[event.Job.Name]; ok {
			delete(scheduler.jobPlanTable, event.Job.Name)
		}
		log.Printf("Scheduler::handJobEvent,event: JOB_EVENT_DELETE ,job: %s", event.Job.Name)
	case common.JOB_EVENT_KILL:
		//如果该函数在在执行中，调用其取消函数
		if jobExecutingInfo, ok := scheduler.jobExecutingTable[event.Job.Name]; ok {
			log.Printf("Scheduler::jobExecutingInfo.CancelFunc,job: %s", event.Job.Name)
			println("jobExecutingInfo.CancelFunc", jobExecutingInfo.RealTime.UnixNano()/1000/1000)
			jobExecutingInfo.CancelFunc()
		}
		log.Printf("Scheduler::handJobEvent,event: JOB_EVENT_KILL ,job: %s", event.Job.Name)

	}

}

//更新job计划表,返回最近的调度时间
func (scheduler *Scheduler) updateJobPlanTable() time.Duration {
	var (
		nearTime *time.Time
	)
	if len(scheduler.jobPlanTable) == 0 {
		return time.Second
	}
	now := time.Now()
	for _, jobPlan := range scheduler.jobPlanTable {
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {
			//执行job
			scheduler.startJob(jobPlan)
			//更新jobPlan的下次执行时间
			jobPlan.NextTime = jobPlan.Expr.Next(now)
		}
		//寻找jobPlanTable中最近要过期的任务
		if nearTime == nil || jobPlan.NextTime.Before(*nearTime) {
			nearTime = &jobPlan.NextTime
		}
	}
	//返回下次调度时间和现在的间隔
	return (*nearTime).Sub(now)
}

func (scheduler *Scheduler) startJob(jobPlan *common.JobSchedulePlan) {
	//1. 放入执行表中,有可能一个任务会运行很久，导致并发执行
	if _, ok := scheduler.jobExecutingTable[jobPlan.Job.Name]; ok {
		log.Printf("job %s Running,Skip execute", jobPlan.Job.Name)
		return
	}
	//2. 构建执行信息
	jobExecuteInfo := common.BuildJobExecuteStatus(jobPlan)
	//3. 保存到执行信息表
	scheduler.jobExecutingTable[jobExecuteInfo.Job.Name] = jobExecuteInfo

	//4. 发送到Executor
	G_executor.ExecuteJob(jobExecuteInfo)
	log.Printf("Scheduler::startJob,job: %s", jobExecuteInfo.Job.Name)
}
