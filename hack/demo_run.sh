#!/bin/bash

for _ in {1..20}; do
  kubectl create -f ./config/samples/pipeline_run.yaml
done

watch -n 1 kubectl get pipelineruns
