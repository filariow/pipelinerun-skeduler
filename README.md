# pipelinerun-skeduler

This repo contains and MVP for a simple Tekton PipelineRun scheduler.

It is a very simple scheduler that will put in execution no more than a fixed number of PipelineRuns.
The next PipelineRun to be executed is chosen with a First-In First-Out (FIFO) policy.

To prevent the execution of a PipelineRun, we rely on its `spec.status` field.
If the PipelineRun is not started yet and its `spec.status` field is set to `PipelineRunPending`, then Tekton won't process the PipelineRun (cf. https://tekton.dev/docs/pipelines/pipelineruns/#pending-pipelineruns).

SREs can still enable high priority PipelineRuns manually, by clearing the `spec.status` field.

## How it works

Every time there is an event on a PipelineRun, the [SkedulerReconciler](./internal/controller/skeduler_controller.go) will check if there is room for scheduling a new PipelineRun.
This is done by the [canScheduleAPipelineRun function](./internal/controller/skeduler_controller.go), that counts the number of running PipelineRuns and compares them with a fixed threshold.

Once a PipelineRun is selected for execution, the scheduler looks for the next PipelineRun to execute (cf. [findNextPipelineRun fuction](./internal/controller/skeduler_controller.go)).
If it finds the next PipelineRun, the scheduler then clears the PipelineRun's `spec.status` field (cf. [scheduleNextPipelineRun function in SkedulerReconciler](./internal/controller/skeduler_controller.go)).

## Demo

### TL;DR

```bash
./hack/demo_prepare.sh # creates the cluster with dependecies installed

# TODO: make sure kyverno and tekton are running

./hack/demo_run.sh
```

### Prerequites:

* bash
* kind
* kustomize
* kubectl
* helm
* make

### Prepare the demo cluster

To prepare the demo cluster, run the `./hack/demo_prepare.sh` script from the project ROOT folder.
This script will:
1. build the scheduler
1. if exists, delete the kind cluster `pipelinerun-skeduler`
1. create the kind cluster `pipeline-skeduler`
1. load the image into the cluster
1. deploy the scheduler
1. deploy Tekton Pipelines
1. deploy Kyverno
1. deploy the [custom policy](./config/policies/mutate_pipelinerun.yaml) for enforcing the Pending state on new PipelineRuns

### Run the demo

To run the demo, execute the script `./hack/demo_run.sh`.
It will create 20 demo PipelineRuns, that will complete in ~40 seconds each.

The Scheduler will make them execute 5 at a time.

## Future developments

* [ ] the value for the maximum number PipelineRuns that can be enabled at the same time is read from dynamic configuration
    * no need to restart the pod
    * inject a config map as volume and read the value from file at each reconcile or rely on watch the file and cache the value (`inotify`)
* [ ] be fair among tenants and avoid starvation
    * avoid the scenario in which tenant A enqueues 100K PipelineRuns and tenant B needs to wait a long time
    * remove the FIFO policy for a fairer one
* [ ] size PipelineRuns to be more effective in choicing the next one
    * is the PipelineRun preventing a Security fix from being released?
    * is the PipelineRun required to complete a very long task?
    * is the PipelineRun providing high business value?
    * is the PipelineRun a very short one that can be processed very quickly? (room for an AI help?)

