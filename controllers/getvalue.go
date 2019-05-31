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
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)
}

type GetvalueController struct {
	beego.Controller
}

func (this *GetvalueController) Post() {
	devid := this.Input().Get("devid")
	pc, _, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "getvalue.go",
		"line":     line,
		"func":     f.Name(),
		"devid":    devid,
	}).Info("GetvalueController Post")

	reqmsg := msginfo.RestReqMsg{
		Method: "GetData",
		Devid:  devid,
	}
	msginfo.SetRestDeviceBuf(devid)
	msginfo.SetRestReqBuf(reqmsg)

	ob := &types.RestGetvalue{}

	err := json.Unmarshal(this.Ctx.Input.RequestBody, ob)
	if err != nil {
		//this.Ctx
		this.Ctx.WriteString(err.Error())
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "getvalue.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("GetvalueController Get Unmarshal error: ", err)
		rsp := types.RestRspNoData{
			Code: 100001,
		}
		this.Data["json"] = rsp
		return
	}

	pc, _, line, _ = runtime.Caller(0)
	f = runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "getvalue.go",
		"line":     line,
		"func":     f.Name(),
		"ip":       ob.Data.Ip,
		"port":     ob.Data.Port,
		"user":     ob.Data.Username,
		"passwd":   ob.Data.Password,
		"id":       ob.Data.Id,
		"attrid":   ob.Data.AttrId,
	}).Info("GetvalueController Post")

	dd := types.DeviceData{
		Id:     ob.Data.Id,
		AttrId: ob.Data.AttrId,
	}

	var dda []types.DeviceData
	dda = append(dda, dd)
	ddr := types.DeviceDatasReq{
		Method: "GetData",
		Taskid: "reqfromrestapi",
		DevId:  devid,
		Data:   dda,
	}

	v, err := json.Marshal(&ddr)
	if err != nil {
		this.Ctx.WriteString(err.Error())
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "getvalue.go",
			"line":     line,
			"func":     f.Name(),
		}).Error("GetvalueController Post Marshal error: ", err)
	}
	mqttclient.MqttPub(configure.GetMqttCfg(), devid, v)

	//rediskey := "GetData_Rsp" + "_" + devid
	rediskey := mqtt_rsp + "_" + hostname + "_" + "GetData_Rsp" + "_" + devid
	pc, _, line, _ = runtime.Caller(0)
	f = runtime.FuncForPC(pc)
	log.WithFields(log.Fields{
		"filename": "getvalue.go",
		"line":     line,
		"func":     f.Name(),
		"rediskey": rediskey,
	}).Info("GetvalueController")

	redisv, err := redisclient.ReadRedisByte(rediskey)
	if err == nil {
		pc, _, line, _ := runtime.Caller(0)
		f := runtime.FuncForPC(pc)
		log.WithFields(log.Fields{
			"filename": "getvalue.go",
			"line":     line,
			"func":     f.Name(),
			"redisv":   redisv,
		}).Info("GetvalueController")
		drsp := types.DeviceDatasRsp{}
		err = json.Unmarshal(redisv, &drsp)
		if err != nil {

			pc, _, line, _ := runtime.Caller(0)
			f := runtime.FuncForPC(pc)
			log.WithFields(log.Fields{
				"filename": "getvalue.go",
				"line":     line,
				"func":     f.Name(),
			}).Error("GetvalueController Post Marshal error: ", err)

			rsp := types.RestRspNoData{
				Code: 1,
			}
			this.Data["json"] = &rsp
			this.ServeJSON()
			return
		}
		redisclient.DelRedis(rediskey)
		rsp := types.RestRspGetValue{
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

				drsp := types.DeviceDatasRsp{}
				err = json.Unmarshal(redisv, &drsp)
				if err != nil {

					pc, _, line, _ := runtime.Caller(0)
					f := runtime.FuncForPC(pc)
					log.WithFields(log.Fields{
						"filename": "getvalue.go",
						"line":     line,
						"func":     f.Name(),
					}).Error("GetvalueController Post Marshal error: ", err)

					rsp := types.RestRspNoData{
						Code: 1,
					}
					this.Data["json"] = &rsp
					this.ServeJSON()
					return
				}
				redisclient.DelRedis(rediskey)
				rsp := types.RestRspGetValue{
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
					"filename": "getvalue.go",
					"line":     line,
					"func":     f.Name(),
				}).Error("GetvalueController wait for rsp timeout!")

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
