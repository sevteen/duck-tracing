#!/bin/bash

./reset_ducktracing.sh

kubectl delete service/jaeger-agent
kubectl delete service/jaeger-query
kubectl delete service/jaeger-collector
kubectl delete service/zipkin
kubectl delete deployment/jaeger-deployment