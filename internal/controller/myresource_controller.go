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
	"errors"
	"fmt"
	"strconv"
	"time"

	"helm.sh/helm/v3/pkg/chartutil"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	samplev1 "k8s-controller.ad/api/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// MyResourceReconciler reconciles a MyResource object
type MyResourceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=sample.k8s-controller.ad,resources=myresources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=sample.k8s-controller.ad,resources=myresources/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=sample.k8s-controller.ad,resources=myresources/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MyResource object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.0/pkg/reconcile
func (r *MyResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Reconciling MyResource", "namespace", req.Namespace, "name", req.Name)

	// Call the reconcileConfigMapSSA function
	if err := r.reconcileConfigMapSSA(ctx); err != nil {
		errors.Join(err, errors.New("failed to reconcile ConfigMap by SSA"))
		return ctrl.Result{}, err
	}
	// Call the reconcileConfigMapWithUpdate function
	if err := r.reconcileConfigMapWithUpdate(ctx); err != nil {
		errors.Join(err, errors.New("failed to reconcile ConfigMap by update"))
		return ctrl.Result{}, err
	}
	// Call the reconcileConfigMapWithPatch function
	if err := r.reconcileConfigMapWithPatch(ctx); err != nil {
		errors.Join(err, errors.New("failed to reconcile ConfigMap by patch"))
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&samplev1.MyResource{}).
		Named("myresource").
		Complete(r)
}

// reconcileConfigMapSSA contains SSA logic for ConfigMap
func (r *MyResourceReconciler) reconcileConfigMapSSA(ctx context.Context) error {
	name := "example-ssa-resource"
	desired := getMyChildResource(name)

	if err := r.Client.Get(
		ctx, client.ObjectKey{Namespace: desired.GetNamespace(), Name: desired.GetName()}, desired,
	); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		r.Client.Create(ctx, desired)
	}

	if desired.Labels["test-mode"] == "origin" {
		desired.SetLabels(LabelsModified)
		desired.Spec = SpecModified
		desired.Status.State = "done"
	} else {
		desired.SetLabels(LabelsOrigin)
		desired.Spec = SpecOrigin
	}
	patchOpts := []client.PatchOption{
		client.ForceOwnership,
		client.FieldOwner("my-ssa-controller"),
	}

	desired.SetGroupVersionKind(samplev1.GroupVersion.WithKind("MyChildResource"))
	desired.ManagedFields = nil
	return r.Client.Patch(ctx, desired, client.Apply, patchOpts...)
}

func (r *MyResourceReconciler) reconcileConfigMapWithUpdate(ctx context.Context) error {
	name := "example-update-resource"
	desired := getMyChildResource(name)

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, desired, func() error {
		target := desired.DeepCopy()

		if strintToInt(desired.Labels["counter"]) < 0 {
			target.Labels["counter"] = fmt.Sprint(strintToInt(desired.Labels["counter"]) + 1)
		} else {
			if desired.Labels["test-mode"] == "origin" {
				target.SetLabels(LabelsModified)
				target.Spec = SpecModified
			} else {
				target.SetLabels(LabelsOrigin)
				target.Spec = SpecOrigin
			}
		}

		d, err := runtime.DefaultUnstructuredConverter.ToUnstructured(desired)
		if err != nil {
			return err
		}

		t, err := runtime.DefaultUnstructuredConverter.ToUnstructured(target)
		if err != nil {
			return err
		}

		result := chartutil.CoalesceTables(t, d)
		runtime.DefaultUnstructuredConverter.FromUnstructured(result, desired)
		return nil
	})

	return err
}

func (r *MyResourceReconciler) reconcileConfigMapWithPatch(ctx context.Context) error {
	name := "example-patch-resource"
	desired := getMyChildResource(name)

	_, err := controllerutil.CreateOrPatch(ctx, r.Client, desired, func() error {
		target := desired.DeepCopy()

		if strintToInt(desired.Labels["counter"]) < 0 {
			target.Labels["counter"] = fmt.Sprint(strintToInt(desired.Labels["counter"]) + 1)
		} else {
			if desired.Labels["test-mode"] == "origin" {
				target.SetLabels(LabelsModified)
				target.Spec = SpecModified
			} else {
				target.SetLabels(LabelsOrigin)
				target.Spec = SpecOrigin
			}
		}

		d, err := runtime.DefaultUnstructuredConverter.ToUnstructured(desired)
		if err != nil {
			return err
		}

		t, err := runtime.DefaultUnstructuredConverter.ToUnstructured(target)
		if err != nil {
			return err
		}

		result := chartutil.CoalesceTables(t, d)
		runtime.DefaultUnstructuredConverter.FromUnstructured(result, desired)
		return nil
	})

	return err
}

var (
	LabelsOrigin = map[string]string{
		"test-mode": "origin",
		"counter":   "0",
		"imOrigin":  "yes",
	}
	LabelsModified = map[string]string{
		"test-mode": "modified",
		"counter":   "0",
	}

	SpecOrigin = samplev1.MyChildResourceSpec{
		Foo: "foo",
		FooMap: map[string]string{
			"key1": "value1",
			"key2": "value1-2",
		},
		FooList: []string{"1", "2", "3"},
	}
	SpecModified = samplev1.MyChildResourceSpec{
		Foo: "foo",
		FooMap: map[string]string{
			"key1": "value1",
		},
		FooList: []string{"1", "2"},
	}
)

func strintToInt(s string) int {
	out, _ := strconv.Atoi(s)
	return out
}

func getMyChildResource(name string) *samplev1.MyChildResource {
	return &samplev1.MyChildResource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			Labels:    LabelsOrigin,
		},
	}

}
