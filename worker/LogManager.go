package worker

import (
	"context"
	"github.com/BreezeTeam/scheduler/common"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"time"
)

type LogManager struct {
	client     *mongo.Client         //mongo 客户端
	collection *mongo.Collection     //table
	logChan    chan *common.JobLog   //传输common.JobLog的管道
	commitChan chan *common.LogBatch //提交时是一批数据，一个batch 一个batch 的传输
}

var (
	G_logManager *LogManager
)

func InitLogManager() (err error) {
	var (
		client *mongo.Client
	)

	ctx, _ := context.WithTimeout(context.Background(), time.Duration(G_config.MongoDBConnectionTimeout)*time.Millisecond)

	credential := options.Credential{
		Username:      G_config.MongoDBUsername,
		Password:      G_config.MongoDBPassword,
		AuthSource:    G_config.MongoDBAuthSource,
		AuthMechanism: "SCRAM-SHA-256",
	}

	clientOptions := options.Client().
		ApplyURI(G_config.MongoDBUri).
		SetMaxPoolSize(uint64(100)).
		SetAuth(credential)

	if client, err = mongo.Connect(ctx, clientOptions); err != nil {
		return
	}

	if err = client.Ping(context.Background(), readpref.Primary()); err != nil {
		return
	}

	// 赋值单例
	G_logManager = &LogManager{
		client: client,
		collection: client.
			Database(G_config.MongoDBDataBase).
			Collection("cron"),
		logChan:    make(chan *common.JobLog, 1000),
		commitChan: make(chan *common.LogBatch, 1000),
	}
	log.Printf("InitLogManager success")
	return
}

func LoggerLoop() {
	var (
		logBatch    *common.LogBatch
		commitTimer *time.Timer
	)
	for {
		select {
		case log := <-G_logManager.logChan: //从 日志管道中获取数据
			if logBatch == nil { //没有历史batch
				logBatch = &common.LogBatch{}
				commitTimer = time.AfterFunc( //规定时间后自动提交
					time.Duration(G_config.LogCommitTimeout)*time.Millisecond,
					func(batch *common.LogBatch) func() { //闭包
						return func() {
							G_logManager.commitChan <- batch //定时器到时限时，发送到提交队列
						}
					}(logBatch),
				)
			}

			//将日志进行追加
			logBatch.Logs = append(logBatch.Logs, log)
			//batch 满的时候，立刻commit
			if len(logBatch.Logs) >= G_config.LogBatchSize {
				G_logManager.saveLogs(logBatch) //保存到mongo
				logBatch = nil                  // 清空logBatch
				commitTimer.Stop()
			}
		case batch := <-G_logManager.commitChan:
			G_logManager.saveLogs(batch) //保存到mongo
			logBatch = nil               // 清空
		}

	}
}

func (logManager *LogManager) saveLogs(batch *common.LogBatch) {
	logManager.collection.InsertMany(context.TODO(), batch.Logs)
	log.Printf("LogManager::saveLogs success,batchs %d", len(batch.Logs))
}

func (logManager *LogManager) SendLog(jobLog *common.JobLog) {
	select {
	case logManager.logChan <- jobLog:
	default:
		//队列满则丢弃，不阻塞
		log.Printf("LogManager::SendLog queue full,job: %s ,startTime: %d", jobLog.JobName, jobLog.StartTime)
	}

}
