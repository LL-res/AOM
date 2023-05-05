import os
import signal
import sys
import json
import socket

import data_preparation
import net
import param


class Request(object):
    def __init__(self, d):
        d = json.loads(d)
        self.__dict__ = d
        

class Response(object):
    def __init__(self,trained=None,key=None,prediction=None,loss=None,error=""):
        self.trained = trained
        self.key = key
        self.prediction = prediction
        self.loss = loss
        self.error = error
        
        
def check_status(request):
    status = 0
    if hasattr(request,'train_history') and request.train_history is not None and len(request.train_history):
        status |= param.STATUS_TRAIN
    if hasattr(request,'predict_history') and request.predict_history is not None and len(request.predict_history):
        status |= param.STATUS_PREDICT
    return status

def handle(request):
    rsp = Response(False,None,None)
    # 进行预测，接收到数据处理之后，需要传回响应
    if hasattr(request,'epochs'):
        param.epochs = request.epochs
    if hasattr(request,'n_layers'):
        param.n_layers = request.n_layers
    if check_status(request) & param.STATUS_PREDICT:
        metrics = []
        for index, val in enumerate(request.predict_history):
            metrics.append(val)
        try:
            out = net.predict(metrics,request.key)
        except Exception as e:
            rsp.error = e.__dict__
            return rsp
        rsp.trained = True
        rsp.key = request.key
        rsp.prediction = out.tolist()[0]
        return rsp
    # 训练，接收到数据后进行处理，传回响应的地址并不是数据传递过来时的地址
    if check_status(request) & param.STATUS_TRAIN:
        metrics = []
        for index, val in enumerate(request.train_history):
            metrics.append(val)
        train_loader = data_preparation.train_data_prepare(metrics)
        client = SocketClient(request.resp_recv_address)
        try:
            loss = net.train(train_loader, request.key)
        except Exception as e:
            rsp.error = e.__dict__
            client.send(json.dumps(rsp.__dict__))
            return
        rsp.trained = True
        rsp.key = request.key
        rsp.loss = loss
        data = json.dumps(rsp.__dict__)
        client.send(data)




class SocketClient:
    def __init__(self,address):
        self.address = address
        self.client_socket = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
    def send(self,data):
        self.client_socket.connect(self.address)
        self.client_socket.send(data.encode())
        self.client_socket.close()

class SocketServer:
    def __init__(self):
        # unix domain sockets
        socket_family = socket.AF_UNIX
        socket_type = socket.SOCK_STREAM
        if os.path.exists(param.socket_address):
            os.remove(param.socket_address)
        self.sock = socket.socket(socket_family, socket_type)
        self.sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        self.sock.bind(param.socket_address)
        self.sock.listen(1)
        print(f"listening on '{param.socket_address}'.")

        # register signal handler
        signal.signal(signal.SIGINT, self._signal_handler)
        signal.signal(signal.SIGTERM, self._signal_handler)

    def wait_and_deal_client_connect(self):
        while True:
            connection, client_address = self.sock.accept()
            data = b''
            while True:
                chunk = connection.recv(1024)
                if not chunk:
                    break
                data += chunk
                if data.endswith(b"\n"):
                    break
            print(f"recv data from client '{client_address}': {data.decode()}")
            resp = handle(Request(data.decode().removesuffix("\n")))
            if resp:
                connection.sendall(json.dumps(resp.__dict__).encode())
            connection.shutdown(socket.SHUT_WR)

    def _signal_handler(self, signum, frame):
        print(f"\nreceived signal {signum}, exiting...")
        self.__del__()
        sys.exit(0)

    def __del__(self):

        self.sock.close()
        os.system('rm -rf {}'.format(param.socket_address))

if __name__ == "__main__":
    socket_server_obj = SocketServer()
    socket_server_obj.wait_and_deal_client_connect()
