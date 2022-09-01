/*
Copyright 2022.

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

package controllers

import (
	"context"
	"fmt"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"

	installsv1alpha1 "github.com/robwittman/remote-installer/api/v1alpha1"
)

// RemoteInstallationReconciler reconciles a RemoteInstallation object
type RemoteInstallationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type PendingInstallationRequest struct {
	Object *unstructured.Unstructured
	Gvk    *schema.GroupVersionKind
}

//+kubebuilder:rbac:groups=installs.remote-installer.io,resources=remoteinstallations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=installs.remote-installer.io,resources=remoteinstallations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=installs.remote-installer.io,resources=remoteinstallations/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RemoteInstallation object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *RemoteInstallationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	// https://ymmt2005.hatenablog.com/entry/2020/04/14/An_example_of_using_dynamic_client_of_k8s.io/client-go
	cfg, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return ctrl.Result{}, err
	}

	remoteInstallation := &installsv1alpha1.RemoteInstallation{}
	err = r.Get(ctx, req.NamespacedName, remoteInstallation)
	if err != nil {
		if errors.IsNotFound(err) {
			l.Info("Didnt find RemoteInstallation CR, ignoring..")
			return ctrl.Result{}, nil
		}
		l.Error(err, "Failed getting RemoteInstallation CR")
		return ctrl.Result{}, err
	}

	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	requests, err := getRemoteResources(remoteInstallation.Spec.Url, decoder)

	for _, request := range requests {
		mapping, err := findGVR(request.Gvk, cfg)
		if err != nil {
			// log error and continue
			continue
		}
		var dr dynamic.ResourceInterface
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			// namespaced resources should specify the namespace
			dr = dyn.Resource(mapping.Resource).Namespace(request.Object.GetNamespace())
		} else {
			// for cluster-wide resources
			dr = dyn.Resource(mapping.Resource)
		}
		fmt.Printf("Installing %s %s in namespace %s\n", request.Object.GetObjectKind(), request.Object.GetName(), request.Object.GetNamespace())

		data, err := json.Marshal(request.Object)
		if err != nil {
			// log error and continue
			continue
		}

		// 7. Create or Update the object with SSA
		//     types.ApplyPatchType indicates SSA.
		//     FieldManager specifies the field owner ID.
		_, err = dr.Patch(ctx, request.Object.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
			FieldManager: "remote-installer",
		})

	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RemoteInstallationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&installsv1alpha1.RemoteInstallation{}).
		Complete(r)
}

func getRemoteResources(url string, decoder runtime.Decoder) ([]*PendingInstallationRequest, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	sepYamlFiles := strings.Split(string(b), "---")
	retVal := make([]*PendingInstallationRequest, 0, len(sepYamlFiles))
	for _, f := range sepYamlFiles {
		if f == "\n" || f == "" {
			// ignore empty cases
			continue
		}

		obj := &unstructured.Unstructured{}
		dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		_, gvk, err := dec.Decode([]byte(f), nil, obj)
		if err != nil {
			// TODO: We should somehow escalate this error. For now, just swallow and move on
			continue
		}
		retVal = append(retVal, &PendingInstallationRequest{
			Object: obj,
			Gvk:    gvk,
		})
	}
	return retVal, nil
}

// find the corresponding GVR (available in *meta.RESTMapping) for gvk
func findGVR(gvk *schema.GroupVersionKind, cfg *rest.Config) (*meta.RESTMapping, error) {

	// DiscoveryClient queries API server about the resources
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	return mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
}
