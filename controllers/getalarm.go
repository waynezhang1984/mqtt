package controllers

import (
	"encoding/json"
	// "fmt"
	"os"
	"runtime"
	"time"

	"github.com/astaxie/beego"
	"github.com/mqttprocess/configure"
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

type GetalarmController struct {
	beego.Controller
}

func (this *GetalarmController) Post() {
	devid := this.Input().Get("devid")
	pc, _, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "getalarm.go",
		"line":     line,
		"func":     f.Name(),
		"devid":    devid,
	}).Info("GetalarmController Post")

	reqmsg := msginfo.RestReqMsg{
		Method: "GetAlarm",
		Devid:  devid,
	}
	msginfo.SetRestDeviceBuf(devid)
	msginfo.SetRestReqBuf(reqmsg)

	ob := &types.RestGetalarm{}

	err := json.Unmarshal(this.Ctx.Input.RequestBody, ob)
	if err != nil {
		//this.Ctx
		this.Ctx.WriteString(err.Error())
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "getalarm.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("GetalarmController Get Unmarshal error: ", err)
		rsp := types.RestRspNoData{
			Code: 100001,
		}
		this.Data["json"] = rsp
		return
	}

	pc, _, line, _ = runtime.Caller(0)
	f = runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "getalarm.go",
		"line":     line,
		"func":     f.Name(),
		"ip":       ob.Data.Ip,
		"port":     ob.Data.Port,
		"user":     ob.Data.Username,
		"passwd":   ob.Data.Password,
		"id":       ob.Data.Id,
		"type":     ob.Data.Type,
		"sTime":    ob.Data.StartTime,
		"eTime":    ob.Data.EndTime,
	}).Info("GetalarmController Post")

	am := types.AlertMeta{
		Type:      ob.Data.Type,
		Id:        ob.Data.Id,
		StartTime: ob.Data.StartTime,
		EndTime:   ob.Data.EndTime,
	}

	aqr := types.AlertQueryReq{
		Method: "GetAlarm",
		Taskid: "reqfromrestapi",
		DevId:  devid,
		Data:   am,
	}

	v, err := json.Marshal(&aqr)
	if err != nil {
		this.Ctx.WriteString(err.Error())
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "getalarm.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("GetalarmController Post Marshal error: ", err)
	}
	mqttclient.MqttPub(configure.GetMqttCfg(), devid, v)

	//rediskey := "GetData_Rsp" + "_" + devid
	rediskey := mqtt_rsp + "_" + hostname + "_" + "GetAlarm_Rsp" + "_" + devid
	pc, _, line, _ = runtime.Caller(0)
	f = runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "getalarm.go",
		"line":     line,
		"func":     f.Name(),
		"rediskey": rediskey,
	}).Info("GetalarmController")

	redisv, err := redisclient.ReadRedisByte(rediskey)
	if err == nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "getalarm.go",
			"line":     line,
			"func":     f.Name(),
			"redisv":   redisv,
		}).Info("GetalarmController")
		drsp := types.AlertQueryRsp{}
		err = json.Unmarshal(redisv, &drsp)
		if err != nil {

			pc, _, line, _ := runtime.Caller(0)
			f := runtime.FuncForPC(pc)
			log.WithFields(log.Fields{
				"filename": "getalarm.go",
				"line":     line,
				"func":     f.Name(),
			}).Error("GetalarmController Post Marshal error: ", err)

			rsp := types.RestRspNoData{
				Code: 1,
			}
			this.Data["json"] = &rsp
			this.ServeJSON()
			return
		}
		redisclient.DelRedis(rediskey)
		rsp := types.RestRspGetAlarm{
			Code: 0,
			Data: drsp.Data,
		}
		this.Data["json"] = &rsp
		this.ServeJSON()
		return

	}

	repcnt := 0
	for {
		select {
		case <-time.After(200 * time.Millisecond):

			redisv, err := redisclient.ReadRedisByte(rediskey)
			if err == nil {

				drsp := types.AlertQueryRsp{}
				err = json.Unmarshal(redisv, &drsp)
				if err != nil {

					pc, _, line, _ := runtime.Caller(0)
					f := runtime.FuncForPC(pc)
					log.WithFields(log.Fields{
						"filename": "getalarm.go",
						"line":     line,
						"func":     f.Name(),
					}).Error("GetalarmController Post Marshal error: ", err)

					rsp := types.RestRspNoData{
						Code: 1,
					}
					this.Data["json"] = &rsp
					this.ServeJSON()
					return
				}
				redisclient.DelRedis(rediskey)
				rsp := types.RestRspGetAlarm{
					Code: 0,
					Data: drsp.Data,
				}
				this.Data["json"] = &rsp
				this.ServeJSON()
				return

			}

			if repcnt >= 5 {
				pc, _, line, _ := runtime.Caller(0)
				f := runtime.FuncForPC(pc)
				log.WithFields(log.Fields{
					"filename": "getalarm.go",
					"line":     line,
					"func":     f.Name(),
				}).Error("GetalarmController wait for rsp timeout!")

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
