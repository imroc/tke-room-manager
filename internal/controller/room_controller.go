/*
Copyright 2024 imroc.

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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	gamev1alpha1 "github.com/imroc/tke-room-manager/api/v1alpha1"
)

// RoomReconciler reconciles a Room object
type RoomReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=game.cloud.tencent.com,resources=rooms,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=game.cloud.tencent.com,resources=rooms/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=game.cloud.tencent.com,resources=rooms/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;update
// +kubebuilder:rbac:groups="",resources=pods/status,verbs=get
// +kubebuilder:rbac:groups=game.kruise.io,resources=gameserver,verbs=get;list;watch;update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Room object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *RoomReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RoomReconciler) SetupWithManager(mgr ctrl.Manager) error {
	indexer := mgr.GetFieldIndexer()
	indexer.IndexField(context.Background(), &gamev1alpha1.Room{}, "status.ready", func(o client.Object) []string {
		ready := o.(*gamev1alpha1.Room).Status.Ready
		if ready != nil {
			return []string{fmt.Sprint(*ready)}
		}
		return nil
	})
	indexer.IndexField(context.Background(), &gamev1alpha1.Room{}, "status.idle", func(o client.Object) []string {
		idle := o.(*gamev1alpha1.Room).Status.Idle
		if idle != nil {
			return []string{fmt.Sprint(*idle)}
		}
		return nil
	})
	indexer.IndexField(context.Background(), &gamev1alpha1.Room{}, "spec.type", func(o client.Object) []string {
		tp := o.(*gamev1alpha1.Room).Spec.Type
		return []string{tp}
	})
	return ctrl.NewControllerManagedBy(mgr).
		For(&gamev1alpha1.Room{}).
		// Watches(
		// 	&corev1.Pod{},
		// 	handler.EnqueueRequestsFromMapFunc(r.findObjectsForPod),
		// ).
		Complete(r)
}

// func (r *RoomReconciler) findObjectsForPod(ctx context.Context, pod client.Object) []reconcile.Request {
// 	list := &gamev1alpha1.RoomList{}
// 	log := log.FromContext(ctx)
// 	err := r.List(
// 		ctx,
// 		list,
// 		client.InNamespace(pod.GetNamespace()),
// 		client.MatchingFields{
// 			"spec.podName": pod.GetName(),
// 		},
// 	)
// 	if err != nil {
// 		log.Error(err, "failed to list dedicatedclblisteners", "podName", pod.GetName())
// 		return []reconcile.Request{}
// 	}
// 	if len(list.Items) == 0 {
// 		return []reconcile.Request{}
// 	}
//
// 	requests := make([]reconcile.Request, len(list.Items))
// 	for i, item := range list.Items {
// 		requests[i] = reconcile.Request{
// 			NamespacedName: types.NamespacedName{
// 				Name:      item.GetName(),
// 				Namespace: item.GetNamespace(),
// 			},
// 		}
// 	}
// 	return requests
// }
