/**
监听任务的增删改查，强杀等
*/
package worker

import (
	"context"
	"github.com/BreezeTeam/scheduler/common"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"log"
	"time"
)

type JobManager struct {
	client  *clientv3.Client //etcd 客户端
	kv      clientv3.KV      //etcd kv
	lease   clientv3.Lease   //租约
	watcher clientv3.Watcher //监视器
}

var (
	G_jobManager *JobManager
)

func InitJobManager() (err error) {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		watcher clientv3.Watcher
	)

	// 初始化配置
	config = clientv3.Config{
		Endpoints:   G_config.EtcdEndpoints,                                     // 集群地址
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond, // 连接超时时间
	}

	// 建立连接
	if client, err = clientv3.New(config); err != nil {
		return
	}

	// 得到KV和Lease的API子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	watcher = clientv3.NewWatcher(client)

	// 赋值单例
	G_jobManager = &JobManager{
		client:  client,
		kv:      kv,
		lease:   lease,
		watcher: watcher,
	}
	log.Printf("InitJobManager success")
	return
}

func WatchJobs() {
	var (
		err       error
		getResp   *clientv3.GetResponse
		watchChan clientv3.WatchChan
		watchResp clientv3.WatchResponse
		job       *common.Job
		jobName   string
	)
	//1.获取当前的所有任务(with prefix)，以及reversion
	if getResp, err = G_jobManager.kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix()); err != nil {
		return
	}
	//2.将扫描到的所有任务发送给schedule
	for _, kvPair := range getResp.Kvs {
		log.Printf("jobManager::WatchJobs,Scan: %s %s", kvPair.Key, kvPair.Value)
		if job, err = common.UnpackJob(kvPair.Value); err == nil {
			G_scheduler.PushJobEvent(common.BuildJobEvent(common.JOB_EVENT_SAVE, job))
		}
	}

	//获取开始监听版本
	reversion := getResp.Header.Revision + 1
	//开始监听
	watchChan = G_jobManager.watcher.Watch(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithRev(reversion), clientv3.WithPrefix())

	log.Printf("jobManager: WatchJobs running, watched: %s", common.JOB_SAVE_DIR)
	for watchResp = range watchChan {
		for _, event := range watchResp.Events {
			switch event.Type {
			case mvccpb.PUT: //任务更新
				if job, err = common.UnpackJob(event.Kv.Value); err == nil {
					G_scheduler.PushJobEvent(common.BuildJobEvent(common.JOB_EVENT_SAVE, job))
				}

				log.Printf("JobManager::WatchJobs,event JOB_EVENT_SAVE ,%s %s", event.Kv.Key, event.Kv.Value)
			case mvccpb.DELETE: //任务删除
				//没有value，因此通过key获取jobName
				jobName = common.ExtractJobName(string(event.Kv.Key))
				G_scheduler.PushJobEvent(common.BuildJobEvent(common.JOB_EVENT_DELETE, &common.Job{Name: jobName}))

				log.Printf("JobManager::WatchJobs,event JOB_EVENT_DELETE,%s %s", event.Kv.Key, event.Kv.Value)
			}
		}
	}
}

//监听强杀任务
func WatchKill() {
	var (
		watchResp clientv3.WatchResponse
		jobName   string
	)

	//开启监视器，监听强杀目录的前缀
	watchChan := G_jobManager.watcher.Watch(context.TODO(), common.JOB_KILLER_DIR, clientv3.WithPrefix())

	log.Printf("jobManager: WatchKill running, watched: %s", common.JOB_KILLER_DIR)
	for watchResp = range watchChan {
		for _, event := range watchResp.Events {
			switch event.Type {
			case mvccpb.PUT: //写入了强杀任务
				jobName = common.ExtractKillerName(string(event.Kv.Key))
				G_scheduler.PushJobEvent(common.BuildJobEvent(common.JOB_EVENT_KILL, &common.Job{Name: jobName}))
				log.Printf("JobManager::WatchKill,event JOB_EVENT_KILL,%s %s", event.Kv.Key, jobName)
			}
		}
	}
}

//创建一个锁
func CreateLock() *EtcdLock {
	return InitEtcdLock(G_jobManager.client)
}
