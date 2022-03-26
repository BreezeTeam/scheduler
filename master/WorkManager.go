package master

import (
	"context"
	"github.com/BreezeTeam/scheduler/common"
	"github.com/coreos/etcd/clientv3"
	"log"
	"time"
)

type WorkManager struct {
	client *clientv3.Client //etcd 客户端
	kv     clientv3.KV      // etcd kv
	lease  clientv3.Lease   //租约
}

var (
	G_workManager *WorkManager
)

func InitWorkManager() (err error) {
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
	G_workManager = &WorkManager{
		client: client,
		kv:     kv,
		lease:  lease,
	}
	log.Printf("InitWorkManager success")
	return
}

func (workManager WorkManager) ListWorkers() (workList []string, err error) {
	var (
		getResp *clientv3.GetResponse
	)
	workList = make([]string, 0)
	if getResp, err = workManager.kv.Get(context.TODO(), common.JOB_WORKER_DIR, clientv3.WithPrefix()); err != nil {
		return
	}
	for _, kv := range getResp.Kvs {
		workList = append(workList, common.ExtractWorkerIP(string(kv.Key)))
	}

	log.Printf("workManager: list workers in etcd, workerListPrefix: %s, workList: %v", common.JOB_WORKER_DIR, workList)
	return
}
