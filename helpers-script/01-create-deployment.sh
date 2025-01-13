#!/bin/bash

docker build ./logger -t tcp-proxy && \
       	minikube kubectl create -- -f redis-deployments.yaml 

