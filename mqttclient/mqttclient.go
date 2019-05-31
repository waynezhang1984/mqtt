package mqttclient

import (
	"bytes"
	//	"fmt"
	"net"
	"os"

	//	"time"
	"runtime"

	"github.com/mqttprocess/configure"

	//"github.com/devicemanager/kafka"
	proto "github.com/huin/mqtt"
	"github.com/jeffallen/mqtt"
	"github.com/mqttprocess/msginfo"
	"github.com/mqttprocess/types"
	log "github.com/sirupsen/logrus"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)
}

func MqttPub(mq configure.MqttCfg, topic string, value []byte) {

	conn, err := net.Dial("tcp", mq.Host)
	if err != nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "mqttclient.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("dial: ", err)
		return
	}
	cc := mqtt.NewClientConn(conn)
	cc.Dump = mq.Dump

	if err := cc.Connect(mq.User, mq.Pass); err != nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "mqttclient.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("connect: %v\n", err)
		os.Exit(1)
		return
	}

	cc.Publish(&proto.Publish{
		Header:    proto.Header{Retain: mq.Retain},
		TopicName: topic,
		Payload:   proto.BytesPayload(value),
	})

	if mq.Wait {
		<-make(chan bool)
	}

	cc.Disconnect()
}

func MqttSub(mq configure.MqttCfg) {

	conn, err := net.Dial("tcp", mq.Host)
	if err != nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "mqttclient.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("dial: ", err)
		return
	}
	cc := mqtt.NewClientConn(conn)
	cc.Dump = mq.Dump
	cc.ClientId = mq.Id

	tq := []proto.TopicQos{
		{types.DsTopic, proto.QosAtLeastOnce},
	}

	if err := cc.Connect(mq.User, mq.Pass); err != nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "mqttclient.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("connect: %v\n", err)
		os.Exit(1)
		return
	}

	cc.Subscribe(tq)

	for m := range cc.Incoming {

		buf := &bytes.Buffer{}
		m.Payload.WritePayload(buf)

		msg := msginfo.MsgData{
			PubType: "mqtt",
			Data:    buf.Bytes(),
		}
		msginfo.GetDataChan() <- msg

	}
}
