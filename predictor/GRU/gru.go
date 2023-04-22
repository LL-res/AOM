package GRU

import (
	"context"
	"encoding/json"
	"errors"
	automationv1 "github.com/LL-res/AOM/api/v1"
	"github.com/LL-res/AOM/collector"
	"github.com/LL-res/AOM/predictor"
	"go.uber.org/atomic"
	"io"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"net"
	"os"

	"log"
)

const (
	pySocket       = "/tmp/gru.socket"
	RespRecvAdress = "/tmp/rra.socket"
	Epochs         = 100
	Nlayers        = 2
)

// 在controller里控制要不要进行predict或者train，status里记录了模型的状态
// 在这里的预测服务只要专心进行预测即可

func New(collectorWorker collector.MetricCollector, model automationv1.Model, ScaleTargetRef autoscalingv2.CrossVersionObjectReference, withModelKey string) (*GRU, error) {
	return &GRU{
		Base: predictor.Base{
			MetricHistory: collectorWorker.Send(),
		},
		withModelKey:    withModelKey,
		model:           model,
		collectorWorker: collectorWorker,
		readyToPredict:  atomic.NewBool(false),
		address:         model.GRU.Address,
		ScaleTargetRef:  ScaleTargetRef,
	}, nil

}

func (g *GRU) Predict(ctx context.Context, aom *automationv1.AOM) (predictor.PredictResult, error) {
	if !g.readyToPredict.Load() {
		return predictor.PredictResult{}, errors.New("the model is not ready to predict")
	}
	// 如果worker中的数据量不足，直接返回
	if g.collectorWorker.DataCap() < g.model.GRU.LookForward {
		return predictor.PredictResult{}, errors.New("no sufficient data to predict")
	}
	tempData := g.collectorWorker.Send()
	// with timestamp
	predictData := tempData[len(tempData)-g.model.GRU.LookBack:]
	//no timestamp
	predictHistory := make([]float64, 0, len(predictData))
	for _, v := range predictData {
		predictHistory = append(predictHistory, v.Value)
	}
	req := Request{
		PredictHistory: predictHistory,
		LookBack:       g.model.GRU.LookBack,
		LookForward:    g.model.GRU.LookForward,
	}
	reqJson, err := json.Marshal(req)
	if err != nil {
		return predictor.PredictResult{}, err
	}
	conn, err := net.Dial("unix", g.address)
	if err != nil {
		return predictor.PredictResult{}, err
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
		return predictor.PredictResult{}, err
	}
	bufSize := 1024
	buf := make([]byte, bufSize)
	n, err := conn.Read(buf)
	if err != nil {
		return predictor.PredictResult{}, err
	}
	responseJson := string(buf[:n])

	for n == bufSize {
		n, err = conn.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return predictor.PredictResult{}, err
		}
		responseJson += string(buf[:n])
	}
	response := Response{}
	err = json.Unmarshal([]byte(responseJson), &response)
	if err != nil {
		return predictor.PredictResult{}, err
	}
	if !response.Trained {
		return predictor.PredictResult{}, errors.New("the model is not ready to predict")
	}
	result := predictor.PredictResult{}

	return response.Prediction, nil
}

func (g *GRU) GetType() string {
	return automationv1.TypeGRU
}

func (g *GRU) Train(ctx context.Context) error {
	if len(g.MetricHistory) < g.model.GRU.TrainSize {
		return errors.New("no sufficient data to train")
	}
	tempData := g.collectorWorker.Send()
	//with timestamp
	TrainData := tempData[len(tempData)-g.model.GRU.TrainSize:]
	//no timestamp
	TrainHistory := make([]float64, 0, len(TrainData))
	for _, v := range TrainData {
		TrainHistory = append(TrainHistory, v.Value)
	}
	req := Request{
		TrainHistory: TrainHistory,
		LookBack:     g.model.GRU.LookBack,
		LookForward:  g.model.GRU.LookForward,
		BatchSize:    g.model.GRU.TrainSize / 10,
		Epochs:       Epochs,
		NLayers:      Nlayers,
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
		_ = g.WaitAndUpdate(ctx)
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
	g.readyToPredict.Store(resp.Trained)
	return nil
}

func (g *GRU) Key() string {
	return g.withModelKey
}
