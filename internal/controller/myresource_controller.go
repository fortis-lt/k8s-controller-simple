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
	"encoding/json"
	"errors"
	"time"

	"helm.sh/helm/v3/pkg/chartutil"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	samplev1 "k8s-controller.ad/api/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

var (
	AnnotationOrigin = map[string]string{
		"test-mode": "origin",
	}
	AnnotationModified = map[string]string{
		"test-mode": "modified",
	}

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
		Foo: "foo-2",
		FooMap: map[string]string{
			"key1": "value1",
		},
		FooList: []string{"1", "2"},
	}
)

const (
	ManagerName   = "ssa-manager"
	AnnotationKey = "manifest_applied"
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

	// // Call the reconcileChildResourceSSA function
	// if err := r.reconcileChildResourceSSA(ctx); err != nil {
	// 	errors.Join(err, errors.New("failed to reconcile child resource by SSA"))
	// 	return ctrl.Result{}, err
	// }
	// // Call the reconcileChildResourceWithUpdate function
	// if err := r.reconcileChildResourceWithUpdateCurrent(ctx); err != nil {
	// 	errors.Join(err, errors.New("failed to reconcile child resource by update"))
	// 	return ctrl.Result{}, err
	// }
	// // Call the reconcileChildResourceWithUpdate function
	// if err := r.reconcileChildResourceWithReplace(ctx); err != nil {
	// 	errors.Join(err, errors.New("failed to reconcile child resource by update"))
	// 	return ctrl.Result{}, err
	// }
	// // Call the reconcileChildResourceWithPatch function
	// if err := r.reconcileChildResourceWithPatchCurrent(ctx); err != nil {
	// 	errors.Join(err, errors.New("failed to reconcile child resource by patch"))
	// 	return ctrl.Result{}, err
	// }
	// Call the reconcileChildResource function
	if err := r.reconcileChildResourceSuggestion(ctx); err != nil {
		errors.Join(err, errors.New("failed to reconcile child resource by patch"))
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

// reconcileChildResourceSSA contains SSA logic for child resource
func (r *MyResourceReconciler) reconcileChildResourceSSA(ctx context.Context) error {
	name := "example-resource-ssa"

	if err := CreateChildResource(ctx, r.Client, name); err != nil {
		return err
	}

	desired := getMyChildResource(name)
	current := getMyChildResource(name)
	if err := r.Client.Get(
		ctx, client.ObjectKeyFromObject(current), current,
	); err != nil {
		return err
	}

	if current.Labels["skip-change"] != "yes" {
		if current.Labels["test-mode"] == "origin" {
			desired.SetLabels(LabelsModified)
			desired.Spec = SpecModified
			desired.Status.State = "done"
		} else {
			desired.SetLabels(LabelsOrigin)
			desired.Spec = SpecOrigin
		}
	}
	patchOpts := []client.PatchOption{
		client.ForceOwnership,
		client.FieldOwner(ManagerName),
	}

	desired.SetGroupVersionKind(samplev1.GroupVersion.WithKind("MyChildResource"))
	desired.ManagedFields = nil
	return r.Client.Patch(ctx, desired, client.Apply, patchOpts...)
}

func (r *MyResourceReconciler) reconcileChildResourceWithUpdateCurrent(ctx context.Context) error {
	name := "example-resource-update-current"

	if err := CreateChildResource(ctx, r.Client, name); err != nil {
		return err
	}
	desired := getMyChildResource(name)

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, desired, func() error {
		target := desired.DeepCopy()

		if desired.Labels["skip-change"] != "yes" {
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

func (r *MyResourceReconciler) reconcileChildResourceWithReplace(ctx context.Context) error {
	name := "example-resource-update-replace"

	if err := CreateChildResource(ctx, r.Client, name); err != nil {
		return err
	}
	desired := getMyChildResource(name)

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, desired, func() error {

		if desired.Labels["skip-change"] != "yes" {
			if desired.Labels["test-mode"] == "origin" {
				desired.SetLabels(LabelsModified)
				desired.Spec = SpecModified
			} else {
				desired.SetLabels(LabelsOrigin)
				desired.Spec = SpecOrigin
			}
		}
		return nil
	})

	return err
}

func (r *MyResourceReconciler) reconcileChildResourceWithPatchCurrent(ctx context.Context) error {
	name := "example-resource-patch-current"

	if err := CreateChildResource(ctx, r.Client, name); err != nil {
		return err
	}
	desired := getMyChildResource(name)

	_, err := controllerutil.CreateOrPatch(ctx, r.Client, desired, func() error {
		target := desired.DeepCopy()

		if desired.Labels["skip-change"] != "yes" {
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

func (r *MyResourceReconciler) reconcileChildResourceSuggestion(ctx context.Context) error {
	name := "example-resource-suggested"

	// request current
	current := getMyChildResource(name)
	if err := r.Client.Get(
		ctx, client.ObjectKeyFromObject(current), current,
	); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		if err := CreateChildResource(ctx, r.Client, name); err != nil {
			return err
		}
		return nil

	}

	desired := getMyChildResource(name)
	// if current.Labels["skip-change"] != "yes" {
	// 	if current.Labels["test-mode"] == "origin" {
	// 		desired.SetAnnotations(AnnotationModified)
	// 		desired.SetLabels(LabelsModified)
	// 		desired.Spec = SpecModified
	// 	} else {
	// 		desired.SetAnnotations(AnnotationOrigin)
	// 		desired.SetLabels(LabelsOrigin)
	// 		desired.Spec = SpecOrigin
	// 	}
	// }

	patchOpts := []client.PatchOption{
		client.ForceOwnership,
		client.FieldOwner(ManagerName),
	}

	gvk, err := r.getGvk(desired)
	if err != nil {
		return err
	}
	desired.SetGroupVersionKind(gvk)

	// screen the bug with creationTimestamp https://github.com/kubernetes/kubernetes/issues/116861
	unstr, err := runtime.DefaultUnstructuredConverter.ToUnstructured(desired)
	if err != nil {
		return err
	}

	meta := unstr["metadata"].(map[string]interface{})
	delete(meta, "creationTimestamp")
	delete(unstr, "status")
	unstr["metadata"] = meta

	obj := &unstructured.Unstructured{
		Object: unstr,
	}

	return r.Client.Patch(ctx, obj, client.Apply, patchOpts...)
}

func (r *MyResourceReconciler) getGvk(obj client.Object) (schema.GroupVersionKind, error) {
	gvk, _, err := r.Client.Scheme().ObjectKinds(obj)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}
	return gvk[0], nil

}

func CreateChildResource(ctx context.Context, c client.Client, name string) error {
	resource := getMyChildResource(name)

	if err := c.Get(
		ctx, client.ObjectKeyFromObject(resource), resource,
	); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		resource.Annotations = map[string]string{
			"init-annotation": "yes",
		}
		resource.Labels = map[string]string{
			"init-label": "yes",
		}
		c.Create(ctx, resource)
	}
	return nil
}

func getMyChildResource(name string) *samplev1.MyChildResource {
	return &samplev1.MyChildResource{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   "default",
			Annotations: map[string]string{},
			Labels:      LabelsOrigin,
		},
	}

}

func ObjectToState(obj client.Object) (string, error) {
	objStr, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return "", err
	}

	delete(objStr, "status")
	json, err := json.Marshal(objStr)
	if err != nil {
		return "", err
	}
	return string(json), err
}

func StateToObject(s string, obj client.Object) error {
	if s == "" {
		return nil
	}
	if err := json.Unmarshal([]byte(s), obj); err != nil {
		return err
	}
	return nil
}
