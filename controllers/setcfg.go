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

	log.SetFormatter(&log.JSONFormatter{})

	log.SetOutput(os.Stdout)
}

type SetcfgController struct {
	beego.Controller
}

func (this *SetcfgController) Post() {
	devid := this.Input().Get("devid")
	pc, _, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "setcfg.go",
		"line":     line,
		"func":     f.Name(),
		"devid":    devid,
	}).Info("SetcfgController Post")

	reqmsg := msginfo.RestReqMsg{
		Method: "SetConfig",
		Devid:  devid,
	}
	msginfo.SetRestDeviceBuf(devid)
	msginfo.SetRestReqBuf(reqmsg)

	ob := &types.RestSetcfg{}
	err := json.Unmarshal(this.Ctx.Input.RequestBody, ob)
	if err != nil {
		//this.Ctx
		this.Ctx.WriteString(err.Error())
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "setcfg.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("SetcfgController Post Unmarshal error: ", err)
		rsp := types.RestRspNoData{
			Code: 100001,
		}
		this.Data["json"] = rsp
		return
	}

	pc, _, line, _ = runtime.Caller(0)
	f = runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename":    "setcfg.go",
		"line":        line,
		"func":        f.Name(),
		"ip":          ob.Data.Ip,
		"port":        ob.Data.Port,
		"user":        ob.Data.Username,
		"passwd":      ob.Data.Password,
		"id":          ob.Data.Id,
		"attrid":      ob.Data.AttrId,
		"min":         ob.Data.Min,
		"max":         ob.Data.Max,
		"normalvalue": ob.Data.NormalValue,
		"kv":          ob.Data.Kv,
		"bv":          ob.Data.Bv,
	}).Info("SetcfgController Post")

	cd := types.CfgSetInfo{
		Id:          ob.Data.Id,
		AttrId:      ob.Data.AttrId,
		Min:         ob.Data.Min,
		Max:         ob.Data.Max,
		NormalValue: ob.Data.NormalValue,
		Kv:          ob.Data.Kv,
		Bv:          ob.Data.Bv,
	}

	cir := types.CfgSetReq{
		Method: "SetConfig",
		Taskid: "reqfromrestapi",
		DevId:  devid,
		Data:   cd,
	}

	v, err := json.Marshal(&cir)
	if err != nil {
		this.Ctx.WriteString(err.Error())
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "setcfg.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("SetcfgController Post Marshal error: ", err)
	}
	mqttclient.MqttPub(configure.GetMqttCfg(), devid, v)

	//rediskey := "Ctrl_Rsp" + "_" + devid
	rediskey := mqtt_rsp + "_" + hostname + "_" + "SetConfig_Rsp" + "_" + devid
	pc, _, line, _ = runtime.Caller(0)
	f = runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "setcfg.go",
		"line":     line,
		"func":     f.Name(),
		"rediskey": rediskey,
	}).Info("SetcfgController")
	redisv, err := redisclient.ReadRedisByte(rediskey)
	if err == nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "setcfg.go",
			"line":     line,
			"func":     f.Name(),
			"redisv":   redisv,
		}).Info("SetcfgController")
		md, did, cd, err := msgjson.RestDeviceRsp(redisv)
		if err == nil {
			pc, _, line, _ := runtime.Caller(0)
			f := runtime.FuncForPC(pc)
			log.WithFields(log.Fields{
				"filename": "setcfg.go",
				"line":     line,
				"func":     f.Name(),
				"method":   md,
				"devid":    did,
				"code":     cd,
			}).Info("SetcfgController")
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
				md, did, cd, err := msgjson.RestDeviceRsp(redisv)
				if err == nil {
					pc, _, line, _ := runtime.Caller(0)
					f := runtime.FuncForPC(pc)
					log.WithFields(log.Fields{
						"filename": "setcfg.go",
						"line":     line,
						"func":     f.Name(),
						"method":   md,
						"devid":    did,
						"code":     cd,
					}).Info("SetcfgController")
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
					"filename": "setcfg.go",
					"line":     line,
					"func":     f.Name(),
				}).Error("SetcfgController wait for rsp timeout!")

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
