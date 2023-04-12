package GRU

import (
	"context"
	"encoding/json"
	"errors"
	automationv1 "github.com/LL-res/AOM/api/v1"
	"github.com/LL-res/AOM/clients/k8s"
	"github.com/LL-res/AOM/collector"
	"github.com/LL-res/AOM/predictor"
	"io"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"net"
	"os"

	"log"
)

const (
	defaultTimeout = 30000
	RespRecvAdress = "/tmp/rra.socket"
	Epochs         = 100
	Nlayers        = 2
)

type AlgorithmRunner interface {
	RunAlgorithmWithValue(algorithmPath string, value string, timeout int) (string, error)
}

// 在controller里控制要不要进行predict或者train，status里记录了模型的状态
// 在这里的预测服务只要专心进行预测即可
// Predict provides logic for using GRU to make a prediction
type GRU struct {
	predictor.Base
	model           automationv1.Model
	collectorWorker collector.MetricCollector
	readyToPredict  bool
	address         string
	ScaleTargetRef  autoscalingv2.CrossVersionObjectReference
}

type Request struct {
	PredictHistory []float64 `json:"predict_history"`
	TrainHistory   []float64 `json:"train_history"`
	RespRecvAdress string    `json:"resp_recv_address"`
	LookBack       int       `json:"look_back"`
	LookForward    int       `json:"look_forward"`
	BatchSize      int       `json:"batch_size"`
	Epochs         int       `json:"epochs"`
	NLayers        int       `json:"n_layers"`
}
type Response struct {
	Trained    bool      `json:"trained"`
	Prediction []float64 `json:"prediction"`
}

func New(collectorWorker collector.MetricCollector, model automationv1.Model, address string, ScaleTargetRef autoscalingv2.CrossVersionObjectReference) (*GRU, error) {
	return &GRU{
		Base: predictor.Base{
			MetricHistory: collectorWorker.Send(),
		},
		model:           model,
		collectorWorker: collectorWorker,
		readyToPredict:  false,
		address:         address,
		ScaleTargetRef:  ScaleTargetRef,
	}, nil

}
func (g *GRU) Predict(ctx context.Context, aom *automationv1.AOM) ([]int32, error) {
	if !g.readyToPredict {
		return nil, errors.New("the model is not ready to predict")
	}
	// 如果worker中的数据量不足，直接返回
	if g.collectorWorker.DataCap() < g.model.GRU.LookForward {
		return nil, errors.New("no sufficient data to predict")
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
		return nil, err
	}
	conn, err := net.Dial("unix", g.address)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	bufSize := 1024
	buf := make([]byte, bufSize)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	responseJson := string(buf[:n])

	for n == bufSize {
		n, err = conn.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		responseJson += string(buf[:n])
	}
	response := Response{}
	err = json.Unmarshal([]byte(responseJson), &response)
	if err != nil {
		return nil, err
	}
	if !response.Trained {
		return nil, errors.New("the model is not ready to predict")
	}

	scaleGetter, err := k8s.GlobalClient.NewScaleClient()
	if err != nil {
		return nil, err
	}
	scaleObj, err := scaleGetter.Scales(aom.Namespace).Get(ctx, schema.GroupResource{
		Group:    aom.Spec.ScaleTargetRef.APIVersion,
		Resource: aom.Spec.ScaleTargetRef.Kind,
	}, aom.Spec.ScaleTargetRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	//通过预测指标算出需要的副本数
	currentReplicas := scaleObj.Spec.Replicas
	targetReplicasList := make([]int32, 0)

	for _, v := range response.Prediction {
		targetReplicas := int32(v/g.model.GRU.ScaleUpThreshold) * currentReplicas
		currentReplicas = targetReplicas
		targetReplicasList = append(targetReplicasList, targetReplicas)
	}

	return targetReplicasList, nil
}

func (g *GRU) GetType() string {
	return automationv1.TypeGRU
}
func (g *GRU) Train() error {
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
		_ = g.WaitAndUpdate()
	}()

	return nil

}
func (g *GRU) WaitAndUpdate() error {
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
	g.readyToPredict = resp.Trained
	return nil
}
