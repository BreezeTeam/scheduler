package common

import (
	"context"
	"time"
)

type JobExecuteStatus struct {
	Job        *Job               // 任务信息
	PlanTime   time.Time          // 理论上的调度时间
	RealTime   time.Time          // 实际的调度时间
	CancelCtx  context.Context    // 任务command的context  context.WithCancel(context)输出
	CancelFunc context.CancelFunc // 用于取消command执行的cancel函数 context.WithCancel(context)输出
}

type JobExecuteResult struct {
	ExecuteInfo *JobExecuteStatus // 执行状态
	Output      []byte            // 脚本输出
	Err         error             // 脚本错误原因
	StartTime   time.Time         // 启动时间
	EndTime     time.Time         // 结束时间
}

func BuildJobExecuteStatus(jobSchedulePlan *JobSchedulePlan) (jobExecuteInfo *JobExecuteStatus) {
	jobExecuteInfo = &JobExecuteStatus{
		Job:      jobSchedulePlan.Job,
		PlanTime: jobSchedulePlan.NextTime, // 计算调度时间
		RealTime: time.Now(),               // 真实调度时间
	}
	jobExecuteInfo.CancelCtx, jobExecuteInfo.CancelFunc = context.WithCancel(context.TODO())
	return
}
