package utils

import (
	"fmt"
	"testing"
)

func TestSocketSendReq(t *testing.T) {
	req := new(SocketReq)
	req.SetAddress("/tmp/uds_socket").SetBody("client").SetNetwork("unix")
	rsp, err := SocketSendReq(*req)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(rsp)
}
