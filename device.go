package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	// "strconv"
	"strings"

	"io/ioutil"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/astaxie/beego"
	"github.com/mqttprocess/configure"
	"github.com/mqttprocess/decode"
	"github.com/mqttprocess/eureka"
	"github.com/mqttprocess/kafka"
	"github.com/mqttprocess/mqttclient"
	"github.com/mqttprocess/msginfo"

	"github.com/mqttprocess/controllers"
	"github.com/mqttprocess/redis"
	log "github.com/sirupsen/logrus"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)
}

var (
	brokerList    = flag.String("kafka", os.Getenv("KAFKA_PEERS"), "The comma separated list of brokers in the Kafka cluster")
	ctopic        = flag.String("ctopic", "", "REQUIRED: the topic to consume")
	cpartitions   = flag.String("partitions", "all", "The partitions to consume, can be 'all' or comma-separated numbers")
	coffset       = flag.String("offset", "newest", "The offset to start with. Can be `oldest`, `newest`")
	verbose       = flag.Bool("verbose", false, "Whether to turn on sarama logging")
	tlsEnabled    = flag.Bool("tls-enabled", false, "Whether to enable TLS")
	tlsSkipVerify = flag.Bool("tls-skip-verify", false, "Whether skip TLS server cert verification")
	tlsClientCert = flag.String("tls-client-cert", "", "Client cert for client authentication (use with -tls-enabled and -tls-client-key)")
	tlsClientKey  = flag.String("tls-client-key", "", "Client key for client authentication (use with tls-enabled and -tls-client-cert)")

	cbufferSize = flag.Int("buffer-size", 256, "The buffer size of the message channel.")

	ptopic       = flag.String("ptopic", "", "REQUIRED: the topic to produce to")
	pkey         = flag.String("key", "", "The key of the message to produce. Can be empty.")
	pvalue       = flag.String("value", "", "REQUIRED: the value of the message to produce. You can also provide the value on stdin.")
	ppartitioner = flag.String("partitioner", "", "The partitioning scheme to use. Can be `hash`, `manual`, or `random`")
	ppartition   = flag.Int("partition", -1, "The partition to produce to.")
	pshowMetrics = flag.Bool("metrics", false, "Output metrics on successful publish to stderr")
	psilent      = flag.Bool("silent", false, "Turn off printing the message's topic, partition, and offset to stdout")

	mqhost = flag.String("mqtt", "192.168.101.189:1883", "hostname of broker")

	mqid   = flag.String("id", "", "client id")
	mquser = flag.String("user", "", "username")
	mqpass = flag.String("pass", "", "password")
	mqdump = flag.Bool("dump", false, "dump messages?")

	mqretain = flag.Bool("retain", false, "retain message?")
	mqwait   = flag.Bool("wait", false, "stay connected after publishing?")

	eurekaurls  = flag.String("eureka", "192.168.101.189:9000", "eureka server ip")
	redisurls   = flag.String("redis", "192.168.101.189:6379", "eureka server ip")
	mqtt_msg    = "mqtt_msg"
	mqtt_rsp    = "mqtt_rsp"
	hostname, _ = os.Hostname()
)

type CfgPara struct {
	Kafka    string `json:"kafka"`
	Reqtopic string `json:"reqtopic"`
	Rsptopic string `json:"rsptopic"`
	Mqtt     string `json:"mqtt"`
	Eureka   string `json:"eureka"`
	Redis    string `json:"redis"`
}

func main() {
	fmt.Println("welcome mqttprocess")
	flag.Parse()

	initConfig()
	startEureka()

	go func() {
		for {
			mqttclient.MqttSub(configure.GetMqttCfg())
		}
	}()

	go func() {
		for {
			kafka.ConsumeMsg(configure.GetConsumerCfg())
		}
	}()

	go func() {
		for {
			select {
			case v := <-msginfo.GetDataChan():
				if v.PubType == "mqtt" {
					mqttMsgHandle(v)
				} else if v.PubType == "kafka" {
					kafkaMsgHandle(v)
				}
			}
		}

	}()

	beegoRun()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()
	<-done
}

func mqttMsgHandle(v msginfo.MsgData) {
	method, id, taskid, err := msgjson.MqttMsgDecode(v.Data)
	if err != nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "device.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("MqttMsgDecode")
		return
	}
	pc, _, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "device.go",
		"line":     line,
		"func":     f.Name(),
		"method":   method,
		"devid":    id,
		"taskid":   taskid,
	}).Info("mqttMsgHandle")

	if taskid == "reqfromrestapi" {
		saveHttpRsp(v, method, id)
	} else {
		msgid := getPlatMsgId(method)
		msgbyte := msgjson.MqttMsgEncode(v.Data, method, msgid)
		kafka.ProduceMsg(configure.GetProducerCfg(), string(msgbyte))
	}
	//saveHttpRsp(v, method, id)

	// msgid := getPlatMsgId(method)
	// msgbyte := msgjson.MqttMsgEncode(v.Data, method, msgid)
	// kafka.ProduceMsg(configure.GetProducerCfg(), string(msgbyte))

}

func getPlatMsgId(method string) string {

	if method == "GetAlarm_Rsp" || method == "GetConfig_Rsp" || method == "SetConfig_Rsp" || method == "GetSysInfo_Rsp" {
		methodtmp := strings.TrimRight(method, "_Rsp")
		//hn, _ := os.Hostname()
		key := mqtt_msg + "_" + hostname + "_" + methodtmp
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "device.go",
			"line":     line,
			"func":     f.Name(),
			"msgidkey": key,
		}).Info("getPlatMsgId")
		msgid, err := redisclient.ReadRedisString(key)
		if err != nil {
			pc, _, line, _ := runtime.Caller(0)
			f := runtime.FuncForPC(pc)
			log.WithFields(log.Fields{
				"filename": "device.go",
				"line":     line,
				"func":     f.Name(),
			}).Error("getPlatMsgId redis get failed: ", err)
		}
		redisclient.DelRedis(key)
		return msgid
	}
	return "1234"
}
func saveHttpRsp(v msginfo.MsgData, method, id string) {
	if method == "GetData_Rsp" || method == "CheckTime_Rsp" || method == "Rest_Rsp" || method == "Ctrl_Rsp" {
		key := mqtt_rsp + "_" + hostname + "_" + method + "_" + id
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "device.go",
			"line":     line,
			"func":     f.Name(),
			"rspkey":   key,
		}).Info("saveHttpRsp")
		redisclient.SaveRedisByte(key, v.Data)
		return
	}

	if method == "GetConfig_Rsp" || method == "SetConfig_Rsp" || method == "GetAlarm_Rsp" {
		key := mqtt_rsp + "_" + hostname + "_" + method + "_" + id
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "device.go",
			"line":     line,
			"func":     f.Name(),
			"rspkey":   key,
		}).Info("saveHttpRsp")
		redisclient.SaveRedisByte(key, v.Data)
	}

}

func kafkaMsgHandle(v msginfo.MsgData) {
	msgdata, method, msgid, err := msgjson.KafkaMsgDecode(v.Data)
	if err != nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "device.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("kafkaMsgHandle KafkaMsgDecode Error: ", err)
		return
	}

	devid, err := msgjson.PlatMsgDecode(msgdata)
	if err != nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "device.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("kafkaMsgHandle PlatMsgDecode Error: ", err)
		return
	}

	pc, _, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "device.go",
		"line":     line,
		"func":     f.Name(),
		"msgid":    msgid,
		"method":   method,
		"devid":    devid,
	}).Info("kafkaMsgHandle")
	setPlatMsgId(method, msgid)
	mqttclient.MqttPub(configure.GetMqttCfg(), devid, msgdata)

}

func setPlatMsgId(method, msgid string) {
	if strings.Contains(method, "_Rsp") {
		return
	}

	key := mqtt_msg + "_" + hostname + "_" + method
	pc, _, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "device.go",
		"line":     line,
		"func":     f.Name(),
		"msgidkey": key,
	}).Info("setPlatMsgId")
	redisclient.SaveRedisString(key, msgid)

}
func beegoRun() {
	beego.Router("/healthcheck", &controllers.HealthController{})
	beego.Router("/debug", &controllers.DebugController{})
	beego.Router("/loglevelset", &controllers.LogsetController{})

	beego.Router("/login", &controllers.LoginController{})
	beego.Router("/heartbeat", &controllers.HbController{})
	beego.Router("/checktime", &controllers.CkController{})
	beego.Router("/getValue", &controllers.GetvalueController{})
	beego.Router("/setValue", &controllers.SetvalueController{})
	beego.Router("/reset", &controllers.ResetController{})

	beego.Router("/getCfg", &controllers.GetcfgController{})
	beego.Router("/setCfg", &controllers.SetcfgController{})
	beego.Router("/getAlarm", &controllers.GetalarmController{})

	beego.Run()
}

func initConfig() {
	cfg := &CfgPara{}
	data, err := ioutil.ReadFile("conf.json")
	if err != nil {
		log.WithFields(log.Fields{
			"filename": "device.go",
		}).Error("ReadFile Error: ", err)
		return
	}

	err = json.Unmarshal(data, cfg)
	if err != nil {
		log.WithFields(log.Fields{
			"filename": "device.go",
		}).Error("Unmarshal Error: ", err)
		return
	}

	configure.GetLocalIp()

	cp := configure.ProducerCfg{
		Topic:         cfg.Reqtopic,
		BrokerList:    cfg.Kafka,
		Key:           *pkey,
		Value:         "test",
		Partitioner:   *ppartitioner,
		Partition:     *ppartition,
		Verbose:       *verbose,
		ShowMetrics:   *pshowMetrics,
		Silent:        *psilent,
		TlsEnabled:    *tlsEnabled,
		TlsSkipVerify: *tlsSkipVerify,
		TlsClientCert: *tlsClientCert,
		TlsClientKey:  *tlsClientKey,
	}
	configure.SetProducerCfg(cp)

	cc := configure.ConsumerCfg{
		Topic:         cfg.Rsptopic,
		BrokerList:    cfg.Kafka,
		Partitions:    *cpartitions,
		Offset:        *coffset,
		Verbose:       *verbose,
		TlsEnabled:    *tlsEnabled,
		TlsSkipVerify: *tlsSkipVerify,
		TlsClientCert: *tlsClientCert,
		TlsClientKey:  *tlsClientKey,
		BufferSize:    *cbufferSize,
	}
	configure.SetConsumerCfg(cc)

	mq := configure.MqttCfg{
		Host:   cfg.Mqtt,
		User:   *mquser,
		Pass:   *mqpass,
		Dump:   *mqdump,
		Retain: *mqretain,
		Wait:   *mqwait,
		Id:     *mqid,
	}
	configure.SetMqttCfg(mq)

	eu := configure.EurekaServer{
		Url: cfg.Eureka,
	}

	ru := configure.RedisServer{
		Url: cfg.Redis,
	}

	configure.SetEurekaIp(eu)
	configure.SetRedisIp(ru)

	configure.PrintCfg()

}

func startEureka() {
	eureka.InitEurekaClient()
	eureka.RegEureka()
	eureka.SyncEureka()
	eureka.HeartBeatEureka()
}
