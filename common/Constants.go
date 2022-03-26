package common

const (
	JOB_SAVE_DIR = "/cron/jobs/" // 任务保存目录

	JOB_KILLER_DIR = "/cron/killer/" // 任务强杀目录

	JOB_LOCK_DIR = "/cron/lock/" // 任务锁目录

	JOB_WORKER_DIR = "/cron/workers/" // 服务注册目录

	JOB_EVENT_SAVE = 1 // 保存任务事件

	JOB_EVENT_DELETE = 2 // 删除任务事件

	JOB_EVENT_KILL = 3 // 强杀任务事件
)
