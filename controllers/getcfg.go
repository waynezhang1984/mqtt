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

type GetcfgController struct {
	beego.Controller
}

func (this *GetcfgController) Post() {
	devid := this.Input().Get("devid")
	pc, _, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "getcfg.go",
		"line":     line,
		"func":     f.Name(),
		"devid":    devid,
	}).Info("GetcfgController Post")

	reqmsg := msginfo.RestReqMsg{
		Method: "GetConfig",
		Devid:  devid,
	}
	msginfo.SetRestDeviceBuf(devid)
	msginfo.SetRestReqBuf(reqmsg)

	ob := &types.RestGetcfg{}

	err := json.Unmarshal(this.Ctx.Input.RequestBody, ob)
	if err != nil {
		//this.Ctx
		this.Ctx.WriteString(err.Error())
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "getcfg.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("GetcfgController Get Unmarshal error: ", err)
		rsp := types.RestRspNoData{
			Code: 100001,
		}
		this.Data["json"] = rsp
		return
	}

	pc, _, line, _ = runtime.Caller(0)
	f = runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "getcfg.go",
		"line":     line,
		"func":     f.Name(),
		"ip":       ob.Data.Ip,
		"port":     ob.Data.Port,
		"user":     ob.Data.Username,
		"passwd":   ob.Data.Password,
		"id":       ob.Data.Id,
		"attrid":   ob.Data.AttrId,
	}).Info("GetcfgController Post")

	dd := types.DeviceData{
		Id:     ob.Data.Id,
		AttrId: ob.Data.AttrId,
	}

	ddr := types.CfgQueryReq{
		Method: "GetConfig",
		Taskid: "reqfromrestapi",
		DevId:  devid,
		Data:   dd,
	}

	v, err := json.Marshal(&ddr)
	if err != nil {
		this.Ctx.WriteString(err.Error())
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "getcfg.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("GetcfgController Post Marshal error: ", err)
	}
	mqttclient.MqttPub(configure.GetMqttCfg(), devid, v)

	//rediskey := "GetData_Rsp" + "_" + devid
	rediskey := mqtt_rsp + "_" + hostname + "_" + "GetConfig_Rsp" + "_" + devid
	pc, _, line, _ = runtime.Caller(0)
	f = runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "getcfg.go",
		"line":     line,
		"func":     f.Name(),
		"rediskey": rediskey,
	}).Info("GetcfgController")

	redisv, err := redisclient.ReadRedisByte(rediskey)
	if err == nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "getcfg.go",
			"line":     line,
			"func":     f.Name(),
			"redisv":   redisv,
		}).Info("GetcfgController")
		drsp := types.CfgQueryRsp{}
		err = json.Unmarshal(redisv, &drsp)
		if err != nil {

			pc, _, line, _ := runtime.Caller(0)
			f := runtime.FuncForPC(pc)
			log.WithFields(log.Fields{
				"filename": "getcfg.go",
				"line":     line,
				"func":     f.Name(),
			}).Error("GetcfgController Post Marshal error: ", err)

			rsp := types.RestRspNoData{
				Code: 1,
			}
			this.Data["json"] = &rsp
			this.ServeJSON()
			return
		}
		redisclient.DelRedis(rediskey)
		rsp := types.RestRspGetCfg{
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

				drsp := types.CfgQueryRsp{}
				err = json.Unmarshal(redisv, &drsp)
				if err != nil {

					pc, _, line, _ := runtime.Caller(0)
					f := runtime.FuncForPC(pc)
					log.WithFields(log.Fields{
						"filename": "getcfg.go",
						"line":     line,
						"func":     f.Name(),
					}).Error("GetcfgController Post Marshal error: ", err)

					rsp := types.RestRspNoData{
						Code: 1,
					}
					this.Data["json"] = &rsp
					this.ServeJSON()
					return
				}
				redisclient.DelRedis(rediskey)
				rsp := types.RestRspGetCfg{
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
					"filename": "getcfg.go",
					"line":     line,
					"func":     f.Name(),
				}).Error("GetcfgController wait for rsp timeout!")

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
