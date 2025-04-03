#!/bin/bash

kubectl apply -f dist_sched/deploy/central-sched.yaml
kubectl apply -f dist_sched/deploy/remote-sched.yaml
kubectl apply -f dist_sched/deploy/pi-job.yaml
