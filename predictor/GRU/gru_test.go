package GRU

import (
	"context"
	"fmt"
	"github.com/LL-res/AOM/clients/k8s"
	"github.com/LL-res/AOM/common/basetype"
	"github.com/LL-res/AOM/common/consts"
	"github.com/LL-res/AOM/fake"
	"math"
	"os"
	"os/exec"
	"testing"
	"time"
)

var gru *GRU

func TestGRU_Predict(t *testing.T) {
	ctx := context.WithValue(context.Background(), consts.NAMESPACE, "default")
	result, err := gru.Predict(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(result)
}
func TestGRU_Train(t *testing.T) {
	err := gru.Train(context.Background())
	if err != nil {
		t.Error(err)
	}
	select {}
}
func TestMain(m *testing.M) {
	if err := k8s.NewClient(); err != nil {
		panic(err)
	}
	collectWorker := fake.CollectorWorker{
		N: 1000,
		Function: func(i int) float64 {
			return math.Sin(float64(i) * 0.01)
		},
		Start:    time.Now(),
		Interval: 5 * time.Second,
	}
	gruAttr := basetype.GRU{
		Address:        "/tmp/uds_socket",
		RespRecvAdress: RespRecvAdress,
		LookBack:       100,
		LookForward:    60,
		TrainSize:      1000,
		Epochs:         1,
		NLayers:        2,
	}
	var err error
	gru, err = New(&collectWorker, gruAttr, collectWorker.NoModelKey()+"$gru")
	if err != nil {
		panic(err)
	}
	m.Run()
}
func TestPython(t *testing.T) {
	fmt.Println(os.Getwd())
	cmd := exec.Command("python3", pythonDir)
	if err := cmd.Run(); err != nil {
		t.Error(err)
	}
}
