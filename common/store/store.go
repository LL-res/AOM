package store

import (
	"github.com/LL-res/AOM/common/aomtype"
	"k8s.io/apimachinery/pkg/types"
)

var globalStore aomtype.AOMStore

func GetHide(name types.NamespacedName) *aomtype.Hide {
	if nil == globalStore {
		globalStore = make(map[types.NamespacedName]*aomtype.Hide)
	}
	if nil == globalStore[name] {
		h := new(aomtype.Hide)
		h.Init()
		globalStore[name] = h
	}
	return globalStore[name]
}
