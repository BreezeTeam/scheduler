package common

// 变化事件
type JobEvent struct {
	EventType int //  SAVE, DELETE
	Job       *Job
}

// 任务变化事件有2种：1）更新任务 2）删除任务
func BuildJobEvent(eventType int, job *Job) (jobEvent *JobEvent) {
	return &JobEvent{
		EventType: eventType,
		Job:       job,
	}
}
