package utils

import (
	"io"
	"net"
)

type SocketReq struct {
	address    string
	network    string
	body       string
	bufferSize int
}

func (r *SocketReq) SetAddress(address string) *SocketReq {
	if r == nil {
		return nil
	}
	r.address = address
	return r
}
func (r *SocketReq) SetNetwork(network string) *SocketReq {
	if r == nil {
		return r
	}
	r.network = network
	return r
}
func (r *SocketReq) SetBody(body string) *SocketReq {
	if r == nil {
		return r
	}
	r.body = body + "\n"
	return r
}
func (r *SocketReq) SetBufferSize(n int) *SocketReq {
	if r == nil {
		return r
	}
	r.bufferSize = n
	return r
}
func fixSocketReq(req *SocketReq) {
	if req.bufferSize <= 0 {
		req.bufferSize = 1024
	}
	if req.network != "tcp" && req.network != "udp" && req.network != "unix" {
		req.network = "tcp"
	}
}
func SocketSendReq(req SocketReq) (rsp string, err error) {
	fixSocketReq(&req)
	conn, err := net.Dial(req.network, req.address)
	if err != nil {
		return "", err
	}
	defer func() {
		err = conn.Close()
	}()
	// 客户端发送一次的数据接收到响应后断开连接
	_, err = conn.Write([]byte(req.body))

	if err != nil {
		return "", err
	}
	buf := make([]byte, req.bufferSize)
	n, err := conn.Read(buf)
	if err != nil {
		return "", err
	}
	rsp = string(buf[:n])

	for n == req.bufferSize {
		n, err = conn.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		rsp += string(buf[:n])
	}
	return
}
