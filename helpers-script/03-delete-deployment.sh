#!/bin/bash

minikube kubectl delete -- -f redis-deployments.yaml

sleep 4

docker rmi tcp-proxy
