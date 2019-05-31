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

type ResetController struct {
	beego.Controller
}

func (this *ResetController) Post() {
	devid := this.Input().Get("devid")
	pc, _, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "reset.go",
		"line":     line,
		"func":     f.Name(),
		"devid":    devid,
	}).Info("ResetController Post")

	reqmsg := msginfo.RestReqMsg{
		Method: "Rest",
		Devid:  devid,
	}
	msginfo.SetRestDeviceBuf(devid)
	msginfo.SetRestReqBuf(reqmsg)

	ob := &types.RestReset{}
	err := json.Unmarshal(this.Ctx.Input.RequestBody, ob)
	if err != nil {
		//this.Ctx
		this.Ctx.WriteString(err.Error())
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "reset.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("ResetController Post Unmarshal error: ", err)
		rsp := types.RestRspNoData{
			Code: 100001,
		}
		this.Data["json"] = rsp
		return
	}

	pc, _, line, _ = runtime.Caller(0)
	f = runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "reset.go",
		"line":     line,
		"func":     f.Name(),
		"ip":       ob.Data.Ip,
		"port":     ob.Data.Port,
		"user":     ob.Data.Username,
		"passwd":   ob.Data.Password,
		"type":     ob.Data.Type,
		"delay":    ob.Data.Delay,
	}).Info("ResetController Post")

	ri := types.ResetInfo{
		Type:  ob.Data.Type,
		Delay: ob.Data.Delay,
	}

	rr := types.ResetReq{
		Method: "Rest",
		Taskid: "reqfromrestapi",
		DevId:  devid,
		Data:   ri,
	}

	v, err := json.Marshal(&rr)
	if err != nil {
		this.Ctx.WriteString(err.Error())
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "reset.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("ResetController Post Marshal error: ", err)
	}
	mqttclient.MqttPub(configure.GetMqttCfg(), devid, v)

	//rediskey := "Rest_Rsp" + "_" + devid
	rediskey := mqtt_rsp + "_" + hostname + "_" + "Rest_Rsp" + "_" + devid
	pc, _, line, _ = runtime.Caller(0)
	f = runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "reset.go",
		"line":     line,
		"func":     f.Name(),
		"rediskey": rediskey,
	}).Info("ResetController")
	redisv, err := redisclient.ReadRedisByte(rediskey)
	if err == nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "reset.go",
			"line":     line,
			"func":     f.Name(),
			"redisv":   redisv,
		}).Info("ResetController")
		md, did, cd, _ := msgjson.RestDeviceRsp(redisv)
		if md == "Rest_Rsp" || did == devid {
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
				if md == "Rest_Rsp" || did == devid {
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
					"filename": "reset.go",
					"line":     line,
					"func":     f.Name(),
				}).Error("ResetController wait for rsp timeout!")

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
