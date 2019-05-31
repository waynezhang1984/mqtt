package msgjson

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"runtime"

	//"strings"

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

type DeviceRspMessage struct {
	Data   string
	Method string `json:"method"`
	MsgId  string `json:"msgid"`
}

func KafkaMsgDecode(buf []byte) ([]byte, string, string, error) {

	pc, _, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "decode.go",
		"line":     line,
		"func":     f.Name(),
	}).Info("KafkaMsgDecode: ", string(buf))

	info := DeviceRspMessage{}
	err := json.Unmarshal(buf, &info)
	if err != nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "decode.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("kafkaMsgDecode error: ", err)
		return []byte{}, "", "", err
	}

	bbuf, err := base64.StdEncoding.DecodeString(info.Data)
	if err != nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "decode.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("base64 decode failure, error=[%v]\n", err)
		return []byte{}, "", "", err
	}
	pc, _, line, _ = runtime.Caller(0)
	f = runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "decode.go",
		"line":     line,
		"func":     f.Name(),
		"method":   info.Method,
		"msgid":    info.MsgId,
		"Data":     string(bbuf),
	}).Info("KafkaMsgDecode")

	return bbuf, info.Method, info.MsgId, nil
}

func PlatReqDecode(buf []byte) (string, error) {
	info := types.CommonReq{}
	err := json.Unmarshal(buf, &info)
	if err != nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "decode.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("PlatReqDecode msgDecode error: ", err)
		return "", err
	}

	pc, _, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "decode.go",
		"line":     line,
		"func":     f.Name(),
		"method":   info.Method,
		"devid":    info.DevId,
	}).Info("PlatReqDecode")
	return info.DevId, nil
}

func PlatRsp(buf []byte) (string, error) {

	pc, _, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "decode.go",
		"line":     line,
		"func":     f.Name(),
	}).Info("PlatRsp: ", string(buf))

	rsp := types.CommonRsp{}

	err := json.Unmarshal(buf, &rsp)
	if err != nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "decode.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("msgDecode error: ", err)
		return "", err
	}

	pc, _, line, _ = runtime.Caller(0)
	f = runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "decode.go",
		"line":     line,
		"func":     f.Name(),
		"method":   rsp.Method,
		"devid":    rsp.DevId,
		"taskid":   rsp.Taskid,
		"code":     rsp.Code,
	}).Info("PlatRsp")
	return rsp.DevId, nil
}

func PlatMsgDecode(buf []byte) (string, error) {

	id, err := PlatRsp(buf)
	if err == nil {
		return id, err
	}

	id, err = PlatReqDecode(buf)
	if err == nil {
		return id, err
	}

	pc, _, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "decode.go",
		"line":     line,
		"func":     f.Name(),
	}).Info("PlatMsgDecode error: ", err)
	return id, err
}

func DeviceReqDecode(buf []byte) (string, string, string, error) {
	info := types.CommonReq{}
	err := json.Unmarshal(buf, &info)
	if err != nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "decode.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("DeviceReqDecode msgDecode error: ", err)
		return "", "", "", err
	}

	pc, _, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "decode.go",
		"line":     line,
		"func":     f.Name(),
		"method":   info.Method,
		"devid":    info.DevId,
		"taskid":   info.Taskid,
	}).Info("DeviceReqDecode")
	return info.Method, info.DevId, info.Taskid, nil
}

func DeviceRsp(buf []byte) (string, string, string, error) {
	rsp := types.CommonRsp{}

	err := json.Unmarshal(buf, &rsp)
	if err != nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "decode.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("DeviceRsp msgDecode error: ", err)
		return "", "", "", err
	}

	pc, _, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "decode.go",
		"line":     line,
		"func":     f.Name(),
		"method":   rsp.Method,
		"devid":    rsp.DevId,
		"error":    rsp.Code,
		"taskid":   rsp.Taskid,
	}).Info("DeviceRsp")
	return rsp.Method, rsp.DevId, rsp.Taskid, nil
}

func RestDeviceRsp(buf []byte) (string, string, int, error) {
	rsp := types.CommonRsp{}

	err := json.Unmarshal(buf, &rsp)
	if err != nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "decode.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("RestDeviceRsp msgDecode error: ", err)
		return "", "", 1, err
	}

	pc, _, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "decode.go",
		"line":     line,
		"func":     f.Name(),
		"method":   rsp.Method,
		"devid":    rsp.DevId,
		"error":    rsp.Code,
	}).Info("RestDeviceRsp")
	return rsp.Method, rsp.DevId, rsp.Code, nil
}

func MqttMsgDecode(buf []byte) (string, string, string, error) {

	pc, _, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "decode.go",
		"line":     line,
		"func":     f.Name(),
	}).Info("MqttMsgDecode: ", string(buf))

	method, id, tid, err := DeviceReqDecode(buf)
	if err == nil {
		return method, id, tid, err
	}

	method, id, tid, err = DeviceRsp(buf)
	if err == nil {
		return method, id, tid, err
	}

	pc, _, line, _ = runtime.Caller(0)
	f = runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "decode.go",
		"line":     line,
		"func":     f.Name(),
	}).Info("MqttMsgDecode error: ", err)
	return method, id, tid, err
}

func DeviceUpdateReqDecode(buf []byte) (string, string, error) {
	info := types.DeviceUpdateReq{}
	err := json.Unmarshal(buf, &info)
	if err != nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "decode.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("msgDecode error: ", err)
		return "", "", err
	}

	if info.Method != "Upgrade" {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "decode.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("msgDecode error: ", errors.New("method mismatch!"))
		return "", "", errors.New("method mismatch!")
	}

	pc, _, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "decode.go",
		"line":     line,
		"func":     f.Name(),
	}).Info("method: ", info.Method)
	return info.Method, info.DevId, nil
}

func DeviceUpdNotifyReqDecode(buf []byte) (string, string, error) {
	info := types.DeviceUpdNotifyReq{}
	err := json.Unmarshal(buf, &info)
	if err != nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "decode.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("msgDecode error: ", err)
		return "", "", err
	}

	if info.Method != "UpgradeNotify" {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "decode.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("msgDecode error: ", errors.New("method mismatch!"))
		return "", "", errors.New("method mismatch!")
	}

	pc, _, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "decode.go",
		"line":     line,
		"func":     f.Name(),
	}).Info("method: ", info.Method)
	return info.Method, info.DevId, nil
}

func CommonReqDecode(buf []byte) (string, string, error) {
	info := types.CommonReq{}
	err := json.Unmarshal(buf, &info)
	if err != nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "decode.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("msgDecode error: ", err)
		return "", "", err
	}

	pc, _, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "decode.go",
		"line":     line,
		"func":     f.Name(),
		"method":   info.Method,
		"devid":    info.DevId,
	}).Info("CommonReqDecode:")
	return info.Method, info.DevId, nil
}

func DeviceUpdateRsp(buf []byte) (string, error) {
	rsp := types.DeviceUpdateRsp{}

	err := json.Unmarshal(buf, &rsp)
	if err != nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "decode.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("msgDecode error: ", err)
		return "", err
	}

	if rsp.Method != "Upgrade_Rsp" {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "decode.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("msgDecode error: ", errors.New("method mismatch!"))
		return "", errors.New("method mismatch!")
	}

	pc, _, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "decode.go",
		"line":     line,
		"func":     f.Name(),
	}).Info("method: ", rsp.Method)
	return rsp.DevId, nil
}

func DeviceUpdaNotifyRsp(buf []byte) (string, error) {
	rsp := types.DeviceUpdaNotifyRsp{}

	err := json.Unmarshal(buf, &rsp)
	if err != nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "decode.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("msgDecode error: ", err)
		return "", err
	}

	if rsp.Method != "UpgradeNotify_Rsp" {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "decode.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("msgDecode error: ", errors.New("method mismatch!"))
		return "", errors.New("method mismatch!")
	}

	pc, _, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "decode.go",
		"line":     line,
		"func":     f.Name(),
	}).Info("method: ", rsp.Method)
	return rsp.DevId, nil
}

func CommonRsp(buf []byte) (string, error) {
	rsp := types.CommonRsp{}

	err := json.Unmarshal(buf, &rsp)
	if err != nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "decode.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("msgDecode error: ", err)
		return "", err
	}

	switch rsp.Method {
	case "Login_Rsp":
	case "GetData_Rsp":
	case "GetAlarm_Rsp":
	case "GetConfig_Rsp":
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "decode.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("msgDecode error: ", errors.New("method mismatch!"))
		return "", errors.New("method mismatch!")
	default:
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "decode.go",
			"line":     line,
			"func":     f.Name(),
		}).Info("method: ", rsp.Method)
		return rsp.DevId, nil
	}

	pc, _, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "decode.go",
		"line":     line,
		"func":     f.Name(),
	}).Info("method: ", rsp.Method)
	return rsp.DevId, nil
}
