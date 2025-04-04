#!/bin/bash

kubectl delete -f dist_sched/deploy/central-sched.yaml
kubectl delete -f dist_sched/deploy/remote-sched.yaml
kubectl delete -f dist_sched/deploy/pi-job.yaml
