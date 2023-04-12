package k8s

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/scale"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"sync"
)

var (
	GlobalClient *Client
	once         sync.Once
)

type Client struct {
	Config    *rest.Config
	ClientSet *kubernetes.Clientset
}

func NewClient() error {
	var err error
	once.Do(func() {
		err = newClient()
	})
	return err
}
func newClient() error {
	conf, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		log.Println(err)
		return err
	}
	clientSet, err := kubernetes.NewForConfig(conf)
	if err != nil {
		log.Println(err)
		return err
	}
	GlobalClient = &Client{
		Config:    conf,
		ClientSet: clientSet,
	}
	return nil
}
func (c *Client) NewScaleClient() (scale.ScalesGetter, error) {
	apiGroupResources, err := restmapper.GetAPIGroupResources(c.ClientSet)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	mapper := restmapper.NewDiscoveryRESTMapper(apiGroupResources)
	resolver := scale.NewDiscoveryScaleKindResolver(c.ClientSet)

	scalesGetter, err := scale.NewForConfig(c.Config, mapper, dynamic.LegacyAPIPathResolverFunc, resolver)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return scalesGetter, nil

}
