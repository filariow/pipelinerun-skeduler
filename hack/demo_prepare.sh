#!/bin/bash

set -eo pipefail

IMG=pipelinerun-skeduler:demo
CLUSTER_NAME=pipelinerun-skeduler

# build the skeduler
IMG=${IMG} make docker-build 

# Create the cluster
kind delete cluster --name "${CLUSTER_NAME}" || true 
kind create cluster --name "${CLUSTER_NAME}"

# Deploy the skeduler
kind load docker-image "${IMG}" --name "${CLUSTER_NAME}"
IMG=${IMG} make deploy

# Deploy Tekton
kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml

# Deploy Kyverno
helm repo add kyverno https://kyverno.github.io/kyverno/
helm repo update
helm install kyverno kyverno/kyverno -n kyverno --create-namespace

# Deploy Kyverno policies
timeout 10s bash -c "while ! kubectl get clusterpolicies.kyverno.io; do sleep 2; done"
kustomize build ./config/policies | kubectl apply -f -

# Wait for kyverno to rollout
kubectl rollout status -n kyverno deployment kyverno-admission-controller
