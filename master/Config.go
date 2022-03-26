package master

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Config struct {
	ApiPort                  int      `json:"apiPort"`                  //api 端口
	ApiReadTimeout           int      `json:"apiReadTimeout"`           //api 读取请求超时时间
	ApiWriteTimeout          int      `json:"apiWriteTimeout"`          //api 发送响应超时时间
	WebRoot                  string   `json:"webRoot"`                  //web 静态页面路由
	EtcdEndpoints            []string `json:"etcdEndpoints"`            //etcd master集群端口列表
	EtcdDialTimeout          int      `json:"etcdDialTimeout"`          //etcd 连接超时时间
	MongoDBUri               string   `json:"mongoDBUri"`               //DB uri
	MongoDBConnectionTimeout int      `json:"mongoDBConnectionTimeout"` //DB 连接超时时间
	MongoDBUsername          string   `json:"mongoDBUsername"`          //DB 用户名
	MongoDBPassword          string   `json:"mongoDBPassword"`          //DB 密码
	MongoDBAuthSource        string   `json:"mongoDBAuthSource"`        //DB 权限库源
	MongoDBDataBase          string   `json:"mongoDBDataBase"`          //日志存储库名
}

var (
	//单例对象
	G_config *Config
)

func InitConfig(filename string) (err error) {
	var (
		content []byte
		conf    Config
	)
	//1.读取配置文件
	if content, err = ioutil.ReadFile(filename); err != nil {
		return
	}
	//2.反序列化json
	if err = json.Unmarshal(content, &conf); err != nil {
		return
	}

	//3.赋值单例
	G_config = &conf

	log.Printf("InitConfig success")
	return
}
