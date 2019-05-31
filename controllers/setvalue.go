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

type SetvalueController struct {
	beego.Controller
}

func (this *SetvalueController) Post() {
	devid := this.Input().Get("devid")
	pc, _, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "setvalue.go",
		"line":     line,
		"func":     f.Name(),
		"devid":    devid,
	}).Info("SetvalueController Post")

	reqmsg := msginfo.RestReqMsg{
		Method: "Ctrl",
		Devid:  devid,
	}
	msginfo.SetRestDeviceBuf(devid)
	msginfo.SetRestReqBuf(reqmsg)

	ob := &types.RestSetvalue{}
	err := json.Unmarshal(this.Ctx.Input.RequestBody, ob)
	if err != nil {
		//this.Ctx
		this.Ctx.WriteString(err.Error())
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "setvalue.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("SetvalueController Post Unmarshal error: ", err)
		rsp := types.RestRspNoData{
			Code: 100001,
		}
		this.Data["json"] = rsp
		return
	}

	pc, _, line, _ = runtime.Caller(0)
	f = runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "setvalue.go",
		"line":     line,
		"func":     f.Name(),
		"ip":       ob.Data.Ip,
		"port":     ob.Data.Port,
		"user":     ob.Data.Username,
		"passwd":   ob.Data.Password,
		"id":       ob.Data.Id,
		"attrid":   ob.Data.AttrId,
		"value":    ob.Data.Value,
	}).Info("SetvalueController Post")

	cd := types.ControlData{
		Id:     ob.Data.Id,
		AttrId: ob.Data.AttrId,
		Value:  ob.Data.Value,
	}

	cir := types.ControlInfoReq{
		Method: "Ctrl",
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
			"filename": "setvalue.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("SetvalueController Post Marshal error: ", err)
	}
	mqttclient.MqttPub(configure.GetMqttCfg(), devid, v)

	//rediskey := "Ctrl_Rsp" + "_" + devid
	rediskey := mqtt_rsp + "_" + hostname + "_" + "Ctrl_Rsp" + "_" + devid
	pc, _, line, _ = runtime.Caller(0)
	f = runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "setvalue.go",
		"line":     line,
		"func":     f.Name(),
		"rediskey": rediskey,
	}).Info("SetvalueController")
	redisv, err := redisclient.ReadRedisByte(rediskey)
	if err == nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "setvalue.go",
			"line":     line,
			"func":     f.Name(),
			"redisv":   redisv,
		}).Info("SetvalueController")
		md, did, cd, _ := msgjson.RestDeviceRsp(redisv)
		if md == "Ctrl_Rsp" || did == devid {
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
				if md == "Ctrl_Rsp" || did == devid {
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
					"filename": "setvalue.go",
					"line":     line,
					"func":     f.Name(),
				}).Error("SetvalueController wait for rsp timeout!")

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
