#!/bin/bash

docker build ./go-logger -t tcp-proxy-go && \
       	minikube kubectl create -- -f redis-deployments.yaml 

