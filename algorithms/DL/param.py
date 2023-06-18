import torch

model_type ='GRU'
look_back = 100
look_forward = 60
batch_size = look_forward/10
n_layers = 2
epochs = 100

STATUS_TRAIN =  1
STATUS_PREDICT = 2

device = torch.device("cuda") if torch.cuda.is_available() else torch.device("cpu")

socket_address = '/tmp/uds_socket'