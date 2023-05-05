package k8s

import (
	"context"
	"fmt"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"log"
	"testing"
)

func TestClient_GetReplica(t *testing.T) {
	deploy, err := GlobalClient.ClientSet.AppsV1().Deployments("default").Get(
		context.Background(),
		"my-app-deployment",
		v1.GetOptions{},
	)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(*deploy.Spec.Replicas)
	n, err := GlobalClient.GetReplica("default", autoscalingv2.CrossVersionObjectReference{
		Kind:       "Deployment",
		Name:       "my-app-deployment",
		APIVersion: "apps",
	})
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(n)

}
func TestClient_NewScaleClient(t *testing.T) {
	scale, err := GlobalClient.NewScaleClient()
	if err != nil {
		t.Error(err)
		return
	}
	obj, err := scale.Scales("default").Get(context.Background(), schema.GroupResource{
		Group:    "apps",
		Resource: "Deployment",
	}, "my-app-deployment", v1.GetOptions{})
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(obj.Spec.Replicas)
}
func TestMain(m *testing.M) {
	err := NewClient()
	if err != nil {
		log.Panic(err)
		return
	}

	m.Run()
}
