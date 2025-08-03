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
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	jcrsv1 "test.jcrs.dev/jobrunner/api/v1"
)

// LeviathanBuildReconciler reconciles a LeviathanBuild object
type LeviathanBuildReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *LeviathanBuildReconciler) jobSpecsEqual(existing *batchv1.Job, desired *batchv1.JobSpec) bool {
	// Compare the specs using Kubernetes semantic equality
	return equality.Semantic.DeepEqual(existing.Spec, *desired)
}

// +kubebuilder:docs-gen:collapse=jobSpecsEqual

// +kubebuilder:rbac:groups=jcrs.jcrs.dev,resources=leviathanbuilds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=jcrs.jcrs.dev,resources=leviathanbuilds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=jcrs.jcrs.dev,resources=leviathanbuilds/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LeviathanBuild object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *LeviathanBuildReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	/*
		### 1: Load the LeviathanBuild by name

		We'll fetch the LeviathanBuild using our client. All client methods take a
		context (to allow for cancellation) as their first argument, and the object
		in question as their last. Get is a bit special, in that it takes a
		[`NamespacedName`](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client?tab=doc#ObjectKey)
		as the middle argument (most don't have a middle argument, as we'll see
		below).

		Many client methods also take variadic options at the end.
	*/
	var lvBuild jcrsv1.LeviathanBuild
	if err := r.Get(ctx, req.NamespacedName, &lvBuild); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("LeviathanBuild resource not found. Ignoring since it must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Unable to fetch LeviathanBuild")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	/*
		We need to construct a job based on our LeviathanBuild's template. We'll copy over the spec
		from the template and copy some basic object meta.

		Then, we'll set the "job time" annotation so that we can reconstitute our
		`LastJobTime` field each reconcile.

		Finally, we'll need to set an owner reference. This allows the Kubernetes garbage collector
		to clean up jobs when we delete the LeviathanBuild, and allows controller-runtime to figure out
		which leviathanBuild needs to be reconciled when a given job changes (is added, deleted, completes, etc).
	*/
	constructJobForLeviathanBuild := func(lvBuild *jcrsv1.LeviathanBuild) (*batchv1.Job, error) {
		// We want job names for a given nominal start time to have a deterministic name to avoid the same job being created twice
		name := fmt.Sprintf("%s-%d", lvBuild.Name, time.Time{}.Unix())

		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Labels:      make(map[string]string),
				Annotations: make(map[string]string),
				Name:        name,
				Namespace:   lvBuild.Namespace,
			},
			Spec: *lvBuild.Spec.JobTemplate.Spec.DeepCopy(),
		}
		for k, v := range lvBuild.Spec.JobTemplate.Annotations {
			job.Annotations[k] = v
		}
		for k, v := range lvBuild.Spec.JobTemplate.Labels {
			job.Labels[k] = v
		}
		if err := ctrl.SetControllerReference(lvBuild, job, r.Scheme); err != nil {
			return nil, err
		}

		return job, nil
	}
	// +kubebuilder:docs-gen:collapse=constructJobForLeviathanBuild

	/*
		The reconciler finds the job owned by the leviathanBuild for the status.

		Status should be able to be reconstituted from the state of the world,
		so it's generally not a good idea to read from the status of the root object.
		Instead, you should reconstruct it every run.

		We can check if a job is "finished" and whether it succeeded or failed using status
		conditions. We'll put that logic in a helper to make our code cleaner.
	*/

	// Check if the Job already exists, if not create a new one
	existingJob := &batchv1.Job{}
	err := r.Get(ctx, req.NamespacedName, existingJob)
	if err != nil && apierrors.IsNotFound(err) {
		// Define a new Job
		job, err := constructJobForLeviathanBuild(&lvBuild)
		if err != nil {
			log.Error(err, "unable to construct job from template")
			// don't bother requeuing until we get a change to the spec
			return ctrl.Result{}, nil
		}
		log.Info("Creating a new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
		if err := r.Create(ctx, job); err != nil {
			log.Error(err, "Failed to create new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
			return ctrl.Result{}, err
		}
		// Requeue the request to ensure the Job is created
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Job")
		return ctrl.Result{}, err
	}

	// Ensure the Job spec matches the desired state
	if !r.jobSpecsEqual(existingJob, &lvBuild.Spec.JobTemplate.Spec) {
		log.Info("Job Spec doesn't match desired state. Deleting existing job.", "Job.Namespace", existingJob.Namespace, "Job.Name", existingJob.Name)
		// Specs don't match, need to recreate
		if err := r.Delete(ctx, existingJob); err != nil {
			return ctrl.Result{}, err
		}
		job, err := constructJobForLeviathanBuild(&lvBuild)
		if err != nil {
			log.Error(err, "unable to construct job from template")
			// don't bother requeuing until we get a change to the spec
			return ctrl.Result{}, nil
		}
		log.Info("Creating a new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
		if err := r.Create(ctx, job); err != nil {
			log.Error(err, "Failed to create new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
			return ctrl.Result{}, err
		}
		// Requeue the request to ensure the Job is created
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	/*
		Using the data we've gathered, we'll update the status of our CRD.
		The status subresource ignores changes to spec, so it's less likely to conflict
		with any other updates, and can have separate permissions.
	*/
	if err := r.Status().Update(ctx, &lvBuild); err != nil {
		log.Error(err, "unable to update LeviathanBuild status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

/*
### Setup

In order to allow our reconciler to quickly look up Jobs by their owner, we'll need an index.
We declare an index key that we can later use with the client as a pseudo-field name,
and then describe how to extract the indexed value from the Job object.
The indexer will automatically take care of namespaces for us,
so we just have to extract the owner name if the Job has a LeviathanBuild owner.

Additionally, we'll inform the manager that this controller owns some Jobs, so that it
will automatically call Reconcile on the underlying LeviathanBuild when a Job changes, is
deleted, etc.
*/
var (
	jobOwnerKey = ".metadata.controller"
	apiGVStr    = jcrsv1.GroupVersion.String()
)

// SetupWithManager sets up the controller with the Manager.
func (r *LeviathanBuildReconciler) SetupWithManager(mgr ctrl.Manager) error {

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &batchv1.Job{}, jobOwnerKey, func(rawObj client.Object) []string {
		// grab the job object, extract the owner...
		job := rawObj.(*batchv1.Job)
		owner := metav1.GetControllerOf(job)
		if owner == nil {
			return nil
		}
		// ...make sure it's a LeviathanBuild...
		if owner.APIVersion != apiGVStr || owner.Kind != "LeviathanBuild" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&jcrsv1.LeviathanBuild{}).
		Owns(&batchv1.Job{}).
		Named("leviathanbuild").
		Complete(r)
}
