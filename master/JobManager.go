package master

import (
	"context"
	"encoding/json"
	"github.com/BreezeTeam/scheduler/common"
	"github.com/coreos/etcd/clientv3"
	"log"
	"time"
)

type JobManager struct {
	client *clientv3.Client //etcd 客户端
	kv     clientv3.KV      // etcd kv
	lease  clientv3.Lease   //租约
}

var (
	G_jobManager *JobManager
)

func InitJobManager() (err error) {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
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

	// 赋值单例
	G_jobManager = &JobManager{
		client: client,
		kv:     kv,
		lease:  lease,
	}
	log.Printf("InitJobManager success")
	return
}

func (jobManager *JobManager) SaveJob(job *common.Job) (oldJob *common.Job, err error) {
	var (
		jobKey  string
		jobInfo []byte
		putResp *clientv3.PutResponse
	)
	jobKey = common.JOB_SAVE_DIR + job.Name
	//序列化job信息
	if jobInfo, err = json.Marshal(job); err != nil {
		return
	}

	//保存到etcd
	if putResp, err = jobManager.kv.Put(context.TODO(), jobKey, string(jobInfo), clientv3.WithPrevKV()); err != nil {
		return
	}
	// 如果有旧值，返回旧值
	if putResp.PrevKv != nil {
		json.Unmarshal(putResp.PrevKv.Value, &oldJob)
	}
	log.Printf("JobManager: put in etcd, jobKey: %s, jobInfo: %s,oldValue: %v", jobKey, string(jobInfo), putResp.PrevKv)
	return

}

func (jobManager *JobManager) DeleteJob(name string) (oldJob *common.Job, err error) {
	var (
		delResp *clientv3.DeleteResponse
	)
	jobKey := common.JOB_SAVE_DIR + name
	if delResp, err = jobManager.kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
		return
	}
	if len(delResp.PrevKvs) != 0 {
		json.Unmarshal(delResp.PrevKvs[0].Value, &oldJob)

		log.Printf("JobManager: del in etcd, jobKey: %s, delJob: %s", jobKey, string(delResp.PrevKvs[0].Value))
	}

	log.Printf("JobManager: del in etcd, jobKey: %s, delJob: %s", jobKey, "")
	return
}

func (jobManager *JobManager) ListJob(dir string) (jobList []*common.Job, err error) {
	var (
		getResp *clientv3.GetResponse
	)
	jobList = make([]*common.Job, 0)
	if getResp, err = jobManager.kv.Get(context.TODO(), dir, clientv3.WithPrefix()); err != nil {
		return
	}
	for _, kvPair := range getResp.Kvs {
		job := common.Job{}
		json.Unmarshal(kvPair.Value, &job)
		jobList = append(jobList, &job)
	}

	log.Printf("JobManager: list job in etcd, jobPrefix: %s, jobList: %v", dir, jobList)
	return
}

func (jobManager *JobManager) KillJob(name string) (leaseId int64, err error) {
	var (
		leaseGrantResp *clientv3.LeaseGrantResponse
	)
	jobKillKey := common.JOB_KILLER_DIR + name
	if leaseGrantResp, err = jobManager.lease.Grant(context.TODO(), 1); err != nil {
		return
	}
	leaseRespId := leaseGrantResp.ID
	_, err = jobManager.kv.Put(context.TODO(), jobKillKey, "", clientv3.WithLease(leaseRespId))
	leaseId = int64(leaseRespId)
	log.Printf("JobManager: kill in etcd, jobKey: %s, leaseId: %d", jobKillKey, leaseId)
	return
}
