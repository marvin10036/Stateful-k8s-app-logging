#!/bin/bash

redisClient=$(minikube kubectl get pods | tail -n 2 | grep redis | cut -d " " -f 1 | head -n 1)
redisServer=$(minikube kubectl get pods | tail -n 2 | grep redis | cut -d " " -f 1 | tail -n 1)

xfce4-terminal -e "minikube kubectl exec -- -it $redisClient -- /bin/bash" 
xfce4-terminal -e "minikube kubectl exec -- -it -c logger-pod $redisServer -- /bin/bash"
xfce4-terminal -e "minikube kubectl exec -- -it -c redis-server-pod $redisServer -- /bin/bash"

