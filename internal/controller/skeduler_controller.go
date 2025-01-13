/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	tektonapi "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
)

// SkedulerReconciler reconciles a Skeduler object
type SkedulerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns,verbs=get;list;watch;update

const MaxConcurrentPipelineRuns int = 5

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Skeduler object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *SkedulerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	ll := tektonapi.PipelineRunList{}
	if err := r.List(ctx, &ll); err != nil {
		return ctrl.Result{}, err
	}

	l.Info("calculate if cluster tolerates execution of another PipelineRun")
	if !r.canScheduleAPipelineRun(&ll) {
		return ctrl.Result{}, nil
	}

	l.Info("selecting next PipelineRun to schedule")
	n := r.findNextPipelineRun(&ll)
	if n == nil {
		l.Info("no new PipelinRun to run")
		return ctrl.Result{}, nil
	}

	l.Info("scheduling next PipelineRun", "next-pipelinerun", n.GetNamespacedName())
	return ctrl.Result{}, r.scheduleNextPipelineRun(ctx, n)
}

func (r *SkedulerReconciler) canScheduleAPipelineRun(ll *tektonapi.PipelineRunList) bool {
	running := 0
	for _, p := range ll.Items {
		if !p.IsPending() && !p.IsDone() {
			running++
		}
	}

	return running < MaxConcurrentPipelineRuns
}

func (r *SkedulerReconciler) findNextPipelineRun(ll *tektonapi.PipelineRunList) *tektonapi.PipelineRun {
	var next *tektonapi.PipelineRun
	for _, p := range ll.Items {
		if !p.IsPending() {
			continue
		}

		if next == nil || p.GetCreationTimestamp().Time.Before(next.GetCreationTimestamp().Time) {
			next = &p
		}
	}

	return next
}

func (r *SkedulerReconciler) scheduleNextPipelineRun(ctx context.Context, next *tektonapi.PipelineRun) error {
	// put pipelinerun into execution
	next.Spec.Status = ""
	return r.Update(ctx, next)
}

// SetupWithManager sets up the controller with the Manager.
func (r *SkedulerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tektonapi.PipelineRun{}).
		Named("skeduler").
		Complete(r)
}
