import json
import math
import sys

import numpy as np
import torch

from torch.utils.data import Dataset, DataLoader
from torch.utils.data import TensorDataset

import param

class GRUParameters:
    def __init__(self, d):
       self.__dict__ = d

#gru_params = json.loads(sys.stdin.read(), object_hook=GRUParameters)


def get_metrics(gru_params):
    gru_params = json.loads(sys.stdin.read(), object_hook=GRUParameters)
    if check_status(gru_params) & param.STATUS_PREDICT:
        metrics = []
        for index, val in enumerate(gru_params.predict_history):
            metrics.append(val.metric)
    if check_status(gru_params) & param.STATUS_TRAIN:
        metrics = []
        for index, val in enumerate(gru_params.train_history):
            metrics.append(val.metric)
        train_fake(metrics, gru_params.train_history[0].type)

def train_data_prepare(metrics):
    """
        :param metrics: list of float
    """
    #如修改标签维数，需再进行一次滑动窗口
    print("start preparing data for training")
    metrics = np.array(metrics)

    #去掉标签数据用不到的部分
    labels_t = metrics[param.look_back:]
    #因为标签只有一个数，所以去掉最后一个输入用不上的数
    inputs_t = metrics[:-param.look_forward]

    labels = np.array([])
    for i in range(len(labels_t)-param.look_forward+1):
        element = labels_t[i:i+param.look_forward]
        labels = np.append(labels, element)

    inputs = np.array([])
    #滑动窗口，窗口大小为look_back
    for i in range(len(inputs_t)-param.look_back+1):
        element = metrics[i:i+param.look_back]
        #这个append会把元素全部展开进行append
        inputs = np.append(inputs,element)
    #[[[1.],  [2.],  [3.]],, [[2.],  [3.],  [4.]],, [[3.],  [4.],  [5.]]]
    inputs = inputs.reshape((-1,param.look_back,1))
    #[[4,5], [5,6], [6,7]]
    labels = labels.reshape((-1,param.look_forward))

    train_data = TensorDataset(torch.from_numpy(inputs), torch.from_numpy(labels))
    train_loader = DataLoader(train_data, shuffle=True, batch_size=param.batch_size, drop_last=True)
    print('training data prepared')
    return train_loader

# metrics = [1,2,3,4,5,6,7]
# train_loader = train_data_prepare(metrics)
# for x,y in train_loader:
#     print('x=',x,'y=',y)
