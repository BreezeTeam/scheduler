package master

import (
	"context"
	"github.com/BreezeTeam/scheduler/common"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"time"
)

type LogManager struct {
	client *mongo.Client //mongo 客户端
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
	}
	log.Printf("InitLogManager success")
	return
}

func (logManager *LogManager) LogList(job string, skip int, limit int) (logList []*common.JobLog, err error) {

	var (
		cursor *mongo.Cursor
	)
	logList = make([]*common.JobLog, 0)

	filter := &common.JobLogFilter{
		JobName: job,
	}
	logSort := &common.SortLogByStartTime{
		SortOrder: -1,
	}
	if cursor, err = logManager.client.
		Database(G_config.MongoDBDataBase).
		Collection("cron").
		Find(context.Background(),
			filter,
			options.Find().SetSort(logSort),
			options.Find().SetSkip(int64(skip)),
			options.Find().SetLimit(int64(limit)),
		); err != nil {
		return
	}

	defer cursor.Close(context.TODO())

	for cursor.Next(context.Background()) {
		jobLog := &common.JobLog{}
		cursor.Decode(jobLog)
		logList = append(logList, jobLog)
	}
	log.Printf("logManager: list Log from db, jobKey: %s, skip: %d, limit: %d, logList: %d", job, skip, limit, len(logList))

	return
}

func (logManager *LogManager) InsertLog(jobName string, command string, error string, output string,
	planTime int64, scheduleTime int64, startTime int64, endTime int64) {
	var (
		res *mongo.InsertOneResult
	)
	data := &common.JobLog{
		JobName:      jobName,
		Command:      command,
		Err:          error,
		Output:       output,
		PlanTime:     planTime,
		ScheduleTime: scheduleTime,
		StartTime:    startTime,
		EndTime:      endTime,
	}

	res, _ = logManager.client.Database(G_config.MongoDBDataBase).Collection("cron").InsertOne(context.Background(), data)

	log.Printf("%v", res.InsertedID.(primitive.ObjectID).Hex())
}
