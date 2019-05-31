package controllers

import (
	"encoding/json"
	"os"
	"runtime"
	"time"

	"github.com/astaxie/beego"
	"github.com/mqttprocess/configure"
	"github.com/mqttprocess/decode"
	"github.com/mqttprocess/mqttclient"
	"github.com/mqttprocess/msginfo"
	"github.com/mqttprocess/redis"
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

var (
	mqtt_rsp    = "mqtt_rsp"
	hostname, _ = os.Hostname()
)

type CkController struct {
	beego.Controller
}

func (this *CkController) Post() {
	devid := this.Input().Get("devid")
	pc, _, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "checktime.go",
		"line":     line,
		"func":     f.Name(),
		"devid":    devid,
	}).Info("CkController Post")

	reqmsg := msginfo.RestReqMsg{
		Method: "CheckTime",
		Devid:  devid,
	}
	msginfo.SetRestDeviceBuf(devid)
	msginfo.SetRestReqBuf(reqmsg)

	ob := &types.RestChecktime{}
	err := json.Unmarshal(this.Ctx.Input.RequestBody, ob)
	if err != nil {
		//this.Ctx
		this.Ctx.WriteString(err.Error())
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "checktime.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("CkController Post Unmarshal error: ", err)
		rsp := types.RestRspNoData{
			Code: 100001,
		}
		this.Data["json"] = rsp
		return
	}

	pc, _, line, _ = runtime.Caller(0)
	f = runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "checktime.go",
		"line":     line,
		"func":     f.Name(),
		"ip":       ob.Data.Ip,
		"port":     ob.Data.Port,
		"user":     ob.Data.Username,
		"passwd":   ob.Data.Password,
		"time":     ob.Data.Time,
	}).Info("CkController Post")

	ck := types.CheckTimeReq{
		Method: "CheckTime",
		Taskid: "reqfromrestapi",
		DevId:  devid,
		Data:   ob.Data.Time,
	}

	v, err := json.Marshal(&ck)
	if err != nil {
		this.Ctx.WriteString(err.Error())
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "checktime.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("CkController Post Marshal error: ", err)
	}
	mqttclient.MqttPub(configure.GetMqttCfg(), devid, v)

	//rediskey := "CheckTime_Rsp" + "_" + devid
	rediskey := mqtt_rsp + "_" + hostname + "_" + "CheckTime_Rsp" + "_" + devid
	pc, _, line, _ = runtime.Caller(0)
	f = runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "checktime.go",
		"line":     line,
		"func":     f.Name(),
		"rediskey": rediskey,
	}).Info("CkController")
	redisv, err := redisclient.ReadRedisByte(rediskey)
	if err == nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "checktime.go",
			"line":     line,
			"func":     f.Name(),
			"redisv":   redisv,
		}).Info("CkController")
		md, did, cd, _ := msgjson.RestDeviceRsp(redisv)
		if md == "CheckTime_Rsp" || did == devid {
			redisclient.DelRedis(rediskey)
			rsp := types.RestRspNoData{
				Code: cd,
			}
			this.Data["json"] = &rsp
			this.ServeJSON()
			return
		}
	}

	repcnt := 0

	for {
		select {
		case <-time.After(200 * time.Millisecond):

			redisv, err := redisclient.ReadRedisByte(rediskey)
			if err == nil {
				md, did, cd, _ := msgjson.RestDeviceRsp(redisv)
				if md == "CheckTime_Rsp" || did == devid {
					redisclient.DelRedis(rediskey)
					rsp := types.RestRspNoData{
						Code: cd,
					}
					this.Data["json"] = &rsp
					this.ServeJSON()
					return
				}
			}

			if repcnt >= 5 {
				pc, _, line, _ := runtime.Caller(0)
				f := runtime.FuncForPC(pc)
				log.WithFields(log.Fields{
					"filename": "checktime.go",
					"line":     line,
					"func":     f.Name(),
				}).Error("CkController wait for rsp timeout!")

				redisclient.DelRedis(rediskey)
				rsp := types.RestRspNoData{
					Code: 100001,
				}
				this.Data["json"] = &rsp
				this.ServeJSON()
				return
			}
			repcnt++
		}
	}

}
