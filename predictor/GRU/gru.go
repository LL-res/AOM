package GRU

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/LL-res/AOM/collector"
	"github.com/LL-res/AOM/common/basetype"
	"github.com/LL-res/AOM/common/consts"
	"github.com/LL-res/AOM/common/errs"
	ptype "github.com/LL-res/AOM/predictor/type"
	"github.com/LL-res/AOM/utils"
	"go.uber.org/atomic"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

const (
	pySocket       = "/tmp/gru.socket"
	RespRecvAdress = "/tmp/rra.socket"
	Epochs         = 100
	Nlayers        = 2
	pythonMain     = "../../algorithms/DL/main.py"
	pythonDir      = "../../algorithms/DL"
)

// 在controller里控制要不要进行predict或者train，status里记录了模型的状态
// 在这里的预测服务只要专心进行预测即可

func New(collectorWorker collector.MetricCollector, model map[string]string, withModelKey string) (*GRU, error) {
	cmd := exec.Command("python3", pythonMain)
	go func() {
		if err := cmd.Run(); err != nil {
			log.Println(err)
		}
	}()
	time.Sleep(time.Second)
	j, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}
	param := Param{}
	err = json.Unmarshal(j, &param)
	if err != nil {
		return nil, err
	}
	lookBack, err := strconv.Atoi(param.LookBack)
	if err != nil {
		return nil, err
	}
	lookForward, err := strconv.Atoi(param.LookForward)
	if err != nil {
		return nil, err
	}
	batchSize, err := strconv.Atoi(param.BatchSize)
	if err != nil {
		return nil, err
	}
	trainSize, err := strconv.Atoi(param.TrainSize)
	if err != nil {
		return nil, err
	}
	epochs, err := strconv.Atoi(param.Epochs)
	if err != nil {
		return nil, err
	}
	nLayers, err := strconv.Atoi(param.NLayers)
	if err != nil {
		return nil, err
	}
	return &GRU{
		Base: ptype.Base{
			MetricHistory: collectorWorker.Send(),
		},
		withModelKey: withModelKey,
		model: basetype.GRU{
			Address:        param.Address,
			RespRecvAdress: param.RespRecvAdress,
			LookBack:       lookBack,
			LookForward:    lookForward,
			BatchSize:      batchSize,
			TrainSize:      trainSize,
			Epochs:         epochs,
			NLayers:        nLayers,
		},
		collectorWorker: collectorWorker,
		readyToPredict:  atomic.NewBool(false),
		address:         param.Address,
	}, nil

}

func (g *GRU) Predict(ctx context.Context) (result ptype.PredictResult, err error) {
	if !g.readyToPredict.Load() {
		return ptype.PredictResult{}, errs.UNREADY_TO_PREDICT
	}
	// 如果worker中的数据量不足，直接返回
	if g.collectorWorker.DataCap() < g.model.LookForward {
		return ptype.PredictResult{}, errs.NO_SUFFICENT_DATA
	}
	tempData := g.collectorWorker.Send()
	// with timestamp
	predictData := tempData[len(tempData)-g.model.LookBack:]
	//no timestamp
	predictHistory := make([]float64, 0, len(predictData))
	for _, v := range predictData {
		predictHistory = append(predictHistory, v.Value)
	}
	req := Request{
		Key:            g.withModelKey,
		PredictHistory: predictHistory,
		LookBack:       g.model.LookBack,
		LookForward:    g.model.LookForward,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return ptype.PredictResult{}, err
	}
	socketReq := &utils.SocketReq{}
	socketReq.SetAddress(g.address).SetBody(string(body)).SetNetwork("unix")
	socketRsp, err := utils.SocketSendReq(*socketReq)
	if err != nil {
		return ptype.PredictResult{}, err
	}
	response := Response{}
	err = json.Unmarshal([]byte(socketRsp), &response)
	if err != nil {
		return ptype.PredictResult{}, err
	}
	if !response.Trained {
		return ptype.PredictResult{}, errs.UNREADY_TO_PREDICT
	}
	result.StartMetric = predictHistory[len(predictHistory)-1]
	result.PredictMetric = response.Prediction
	result.Loss = response.Loss
	if err != nil {
		return ptype.PredictResult{}, err
	}
	//// 更新amo的status,TODO 移至上层防止包循环引用
	//statusHistory, _ := aom.Status.PredictorHistory.Load(g.withModelKey)
	//statusHistory.AppendPredictorHistory(time.Now())

	return
}

func (g *GRU) GetType() string {
	return consts.GRU
}

func (g *GRU) Train(ctx context.Context) error {
	if len(g.MetricHistory) < g.model.TrainSize {
		return errs.NO_SUFFICENT_DATA
	}
	tempData := g.collectorWorker.Send()
	//with timestamp
	TrainData := tempData[len(tempData)-g.model.TrainSize:]
	//no timestamp
	TrainHistory := make([]float64, 0, len(TrainData))
	for _, v := range TrainData {
		TrainHistory = append(TrainHistory, v.Value)
	}
	req := Request{
		Key:            g.withModelKey,
		RespRecvAdress: g.model.RespRecvAdress,
		TrainHistory:   TrainHistory,
		LookBack:       g.model.LookBack,
		LookForward:    g.model.LookForward,
		BatchSize:      g.model.TrainSize / 10,
		Epochs:         g.model.Epochs,
		NLayers:        g.model.NLayers,
	}
	reqJson, err := json.Marshal(req)
	if err != nil {
		return err
	}
	conn, err := net.Dial("unix", g.address)
	if err != nil {
		return err
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Println(err)
		}
	}()
	// 客户端发送一次的数据接收到响应后断开连接
	_, err = conn.Write(reqJson)

	if err != nil {
		return err
	}
	// 由于训练时间较长
	// 起一个协程去等输出，此刻直接返回
	// 协程收到返回后进行状态更新
	go func() {
		if err := g.WaitAndUpdate(ctx); err != nil {
			runtime.Goexit()
		}
	}()

	return nil

}

func (g *GRU) WaitAndUpdate(ctx context.Context) error {
	//如果socket文件存在，则进行移除
	if _, err := os.Stat(RespRecvAdress); err == nil {
		err := os.Remove(RespRecvAdress)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	// 收到一次响应后直接断开
	l, err := net.Listen("unix", RespRecvAdress)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() {
		err := l.Close()
		if err != nil {
			log.Println(err)
			return
		}
	}()
	conn, err := l.Accept()
	if err != nil {
		log.Println(err)
		return err
	}
	bufSize := 1024
	buf := make([]byte, bufSize)
	var res string
	for {
		n, err := conn.Read(buf)
		if err == io.EOF {
			res += string(buf[:n])
			break
		}
		if err != nil {
			log.Println(err)
			return err
		}
		if n == bufSize {
			res += string(buf)
			continue
		}
		if n < bufSize {
			res += string(buf[:n])
			break
		}
	}
	resp := Response{}
	err = json.Unmarshal([]byte(res), &resp)
	if err != nil {
		log.Println(err)
		return err
	}
	if resp.Error != "" {
		log.Println(resp.Error)
		return fmt.Errorf("python error : %s", resp.Error)
	}
	log.Println(resp)
	g.readyToPredict.Store(resp.Trained)
	return nil
}

func (g *GRU) Key() string {
	return g.withModelKey
}
