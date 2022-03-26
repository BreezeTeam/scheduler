package worker

import (
	"context"
	"github.com/BreezeTeam/scheduler/common"
	"github.com/coreos/etcd/clientv3"
)

type HandlerFunc func() error
type LockError string

func (err LockError) Error() string {
	return string(err)
}

type EtcdLock struct {
	client         *clientv3.Client
	lease          clientv3.Lease
	kv             clientv3.KV
	leaseGrantResp *clientv3.LeaseGrantResponse
	cancelFunc     context.CancelFunc
	Locked         bool
}

func InitEtcdLock(client *clientv3.Client) *EtcdLock {
	return &EtcdLock{
		client: client,
		lease:  clientv3.NewLease(client),
		kv:     clientv3.NewKV(client),
		Locked: false,
	}
}

func (lock *EtcdLock) Lock(lockKey string, locValue string) (err error) {
	var (
		leaseGrantResp    *clientv3.LeaseGrantResponse
		keepAliveRespChan <-chan *clientv3.LeaseKeepAliveResponse
		txnResp           *clientv3.TxnResponse
	)

	//1.设置续约并自动续约
	//1.1 设置租期
	if leaseGrantResp, err = lock.lease.Grant(context.TODO(), 5); err != nil {
		return err
	}

	//1.2 准备一个用于取消租约的上下文
	cancelCtx, cancelFunc := context.WithCancel(context.TODO())

	//1.3 在函数退出之前持续续租
	if keepAliveRespChan, err = lock.lease.KeepAlive(cancelCtx, leaseGrantResp.ID); err != nil {
		return err
	}

	//1.4 启动处理续约的协程
	go func() {
		var (
			keepAliveResp *clientv3.LeaseKeepAliveResponse
		)
		for {

			select {
			case keepAliveResp = <-keepAliveRespChan:
				if keepAliveResp == nil {
					goto KeepAliveListenEnd
				}
			}
		}
	KeepAliveListenEnd:
	}()

	//2.通过txn事务抢锁
	//2.1 创建事务
	txn := lock.kv.Txn(context.TODO())
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)). //如果 key 不存在
										Then(clientv3.OpPut(lockKey, locValue, clientv3.WithLease(leaseGrantResp.ID))). //上锁
										Else(clientv3.OpGet(lockKey))                                                   //否则抢锁失败
	//2.2 提交事务，并判断是否成功
	if txnResp, err = txn.Commit(); err != nil {
		goto TxnFail
	}
	//2.3 如果失败，则锁被占用
	if !txnResp.Succeeded {
		err = common.ERR_LOCK_ALREADY_REQUIRED
		goto TxnFail
	}

	//3.上锁成功
	lock.cancelFunc = cancelFunc
	lock.leaseGrantResp = leaseGrantResp
	lock.Locked = true
	return
TxnFail:
	cancelFunc()                                         //取消续约
	lock.lease.Revoke(context.TODO(), leaseGrantResp.ID) //直接结束租约
	return err
}

//释放锁
func (lock *EtcdLock) UnLock() {
	if lock.Locked {
		lock.lease.Revoke(context.TODO(), lock.leaseGrantResp.ID) //通过租约id直接结束租约
		lock.cancelFunc()                                         //关闭租约，此时，会导致租约监听协程退出
	}
}
