package worker

import (
	"github.com/BreezeTeam/scheduler/common"
	"math/rand"
	"os/exec"
	"runtime"
	"time"
)

type Executor struct {
}

var (
	G_executor *Executor
)

func InitExecutor() error {
	G_executor = &Executor{}
	return nil
}

func (executor *Executor) ExecuteJob(jobInfo *common.JobExecuteStatus) {
	go func() {
		result := &common.JobExecuteResult{
			ExecuteInfo: jobInfo,
		}
		sysType := runtime.GOOS
		cmd := &exec.Cmd{}

		// 上锁
		// 随机睡眠(0~1s)
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

		//1.请求分布式锁
		lock := CreateLock()
		locError := lock.Lock(common.JOB_LOCK_DIR+jobInfo.Job.Name, "")
		defer lock.UnLock()
		//2.执行
		if locError == nil {
			result.StartTime = time.Now()
			if sysType == "linux" {
				cmd = exec.CommandContext(result.ExecuteInfo.CancelCtx, "/bin/sh", "-c", result.ExecuteInfo.Job.Command)
				output, err := common.CleanCmd(result.ExecuteInfo.CancelCtx, cmd)
				result.Output = output
				result.Err = err
			} else if sysType == "windows" {
				cmd = exec.CommandContext(result.ExecuteInfo.CancelCtx, "Powershell", "-Command", result.ExecuteInfo.Job.Command)
				output, err := cmd.CombinedOutput()
				result.Output = output
				result.Err = err
			}
			result.EndTime = time.Now()
		} else {
			result.Err = locError
			result.EndTime = time.Now()

		}
		println("PushJobExecuteResult", result.ExecuteInfo.RealTime.UnixNano()/1000/1000)
		G_scheduler.PushJobExecuteResult(result)
	}()
}
