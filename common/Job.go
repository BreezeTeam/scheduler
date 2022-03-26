package common

import "encoding/json"

type Job struct {
	Name     string `json:"name"`     //  任务名
	Command  string `json:"command"`  // shell命令
	CronExpr string `json:"cronExpr"` // cron表达式
}

// 反序列化Job
func UnpackJob(value []byte) (job *Job, err error) {
	if err = json.Unmarshal(value, &job); err != nil {
		return
	}
	return
}
