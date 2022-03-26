package worker

import (
	"context"
	"github.com/BreezeTeam/scheduler/common"
	"github.com/coreos/etcd/clientv3"
	"log"
	"net"
	"time"
)

//注册到etcd 服务节点上
type Registrant struct {
	client  *clientv3.Client //etcd 客户端
	kv      clientv3.KV      // etcd kv
	lease   clientv3.Lease   //租约
	localIP string           //本机ip
}

var (
	G_register *Registrant
)

func InitRegistrant() (err error) {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		localIp string
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

	//获取本地ip
	if localIp, err = getLocalIP(); err != nil {
		return
	}

	// 赋值单例
	G_register = &Registrant{
		client:  client,
		kv:      kv,
		lease:   lease,
		localIP: localIp,
	}
	log.Printf("InitRegister success")
	return
}

//在容器中时会不会有问题
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipNet, isIpNet := addr.(*net.IPNet); isIpNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}
	return "", common.ERR_NO_LOCAL_IP_FOUND
}

func Register() {
	var (
		err            error
		leaseGrantResp *clientv3.LeaseGrantResponse
		keepAliveChan  <-chan *clientv3.LeaseKeepAliveResponse
		keepAliveResp  *clientv3.LeaseKeepAliveResponse
		cancelCtx      context.Context
		cancelFunc     context.CancelFunc
	)
	for {
		cancelFunc = nil

		//注册路径
		resitKey := common.JOB_WORKER_DIR + G_register.localIP

		//申请一个10s的租约
		leaseGrantResp, err = G_register.lease.Grant(context.TODO(), 10)
		if err != nil {
			goto CONTINUE
		}
		//自动续租
		keepAliveChan, err = G_register.lease.KeepAlive(context.TODO(), leaseGrantResp.ID)
		if err != nil {
			goto CONTINUE
		}

		cancelCtx, cancelFunc = context.WithCancel(context.TODO())

		//注册到etcd,上下文选择可取消上下文，租约选择之前的可续签的租约
		_, err = G_register.kv.Put(cancelCtx, resitKey, "", clientv3.WithLease(leaseGrantResp.ID))
		if err != nil {
			goto CONTINUE
		}

		//处理租约
		for {
			select {
			case keepAliveResp = <-keepAliveChan:
				if keepAliveResp == nil { //如果返回nil，则续租失败，不然会返回一个GrantResp，可以获取到ID
					goto CONTINUE
				}
				if G_config.Debug {
					log.Printf("Register::KeepAlive SUCCESS")
				}
			}

		}
	CONTINUE:
		time.Sleep(1 * time.Second)
		if cancelFunc != nil {
			cancelFunc()
		}
	}

}
