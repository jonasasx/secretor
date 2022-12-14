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
	"bytes"
	"context"
	"fmt"
	"github.com/thanhpk/randstr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	jonasasxiov1alpha1 "jonasasx.io/secretor/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

// SecretorReconciler reconciles a Secretor object
type SecretorReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=jonasasx.io,resources=secretors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=jonasasx.io,resources=secretors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=jonasasx.io,resources=secretors/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Secretor object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *SecretorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.FromContext(ctx)

	// Fetch the ScaledObject instance
	secretor := &jonasasxiov1alpha1.Secretor{}
	err := r.Client.Get(ctx, req.NamespacedName, secretor)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		reqLogger.Error(err, "Failed to get Secretor")
		return ctrl.Result{}, err
	}

	reqLogger.Info("Reconciling Secretor")
	reqLogger.V(10).Info(fmt.Sprintf("Reconciling: %v", secretor))

	value := ""

	if secretor.Spec.Type == "generate" {
		if secretor.Spec.Generating == nil {
			err = fmt.Errorf("spec.generating is undefined")
			reqLogger.Error(err, err.Error())
			return ctrl.Result{}, err
		}
		for _, injectTo := range secretor.Spec.InjectTo {
			secretRef := injectTo.SecretRef

			secret := &v1.Secret{}
			namespace := ""
			if secretRef.Namespace == nil {
				namespace = secretor.Namespace
			} else {
				namespace = *secretRef.Namespace
			}
			err = r.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: secretRef.Name}, secret)
			if err != nil {
				continue
			}
			if secret.Data != nil {
				if val, ok := secret.Data[secretRef.Field]; ok {
					value = string(val)
					break
				}
			}
		}
		if value == "" {
			generating := secretor.Spec.Generating
			value = randstr.String(generating.Length)
		}
	} else if secretor.Spec.Type == "constant" {
		value = *secretor.Spec.Value
	} else {
		err = fmt.Errorf(fmt.Sprintf("Unknown secretor type \"%s\"", secretor.Spec.Type))
		reqLogger.Error(err, err.Error())
		return ctrl.Result{}, err
	}

	for _, injectTo := range secretor.Spec.InjectTo {
		secretRef := injectTo.SecretRef

		exist := true
		secret := &v1.Secret{}
		namespace := ""
		if secretRef.Namespace == nil {
			namespace = secretor.Namespace
		} else {
			namespace = *secretRef.Namespace
		}
		err = r.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: secretRef.Name}, secret)
		if err != nil {
			if !errors.IsNotFound(err) {
				reqLogger.Error(err, fmt.Sprintf("Failed to get Secret %s/%s", namespace, secret.Name))
				return ctrl.Result{}, err
			}
			exist = false
		}
		if exist {
			origin := secret.DeepCopy()
			byteValue := []byte(value)
			if bytes.Compare(byteValue, secret.Data[secretRef.Field]) != 0 {
				secret.Data[secretRef.Field] = byteValue
				err = r.Client.Patch(ctx, secret, client.MergeFrom(origin))
				if err != nil {
					reqLogger.Error(err, fmt.Sprintf("Failed to patch Secret %s/%s", namespace, secretRef.Name))
					return ctrl.Result{}, err
				}
			}
		} else {
			labels := map[string]string{}
			for k, v := range secretor.Labels {
				labels[k] = v
			}
			secret.TypeMeta = metav1.TypeMeta{
				APIVersion: "v1",
			}
			secret.ObjectMeta = metav1.ObjectMeta{
				Name:      secretRef.Name,
				Namespace: namespace,
				Labels:    labels,
			}
			secret.OwnerReferences = []metav1.OwnerReference{
				{
					Name:       secretor.Name,
					APIVersion: secretor.APIVersion,
					Kind:       secretor.Kind,
					UID:        secretor.UID,
				},
			}
			secret.StringData = map[string]string{
				secretRef.Field: value,
			}
			err = r.Client.Create(ctx, secret)
			if err != nil {
				reqLogger.Error(err, fmt.Sprintf("Failed to create Secret %s/%s", namespace, secretRef.Name))
				return ctrl.Result{}, err
			}
		}
		r.Recorder.Event(secretor, v1.EventTypeNormal, "SecretorSynchronized", "Secretor is synchronized")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&jonasasxiov1alpha1.Secretor{}).
		Complete(r)
}
