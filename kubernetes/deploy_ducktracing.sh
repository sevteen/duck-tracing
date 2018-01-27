#!/bin/bash

kubectl create -f ducktracing.yml

echo "Auth URL: `minikube service auth --url`"
echo "Ducky URL: `minikube service ducky --url`"
echo "Jaeger URL: `minikube service jaeger-query --url`"
