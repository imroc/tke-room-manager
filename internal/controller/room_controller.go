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
	"time"

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
// +kubebuilder:rbac:groups=game.kruise.io,resources=gameservers,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=game.kruise.io,resources=gameservers/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Room object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *RoomReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctl ctrl.Result, err error) {
	_ = log.FromContext(ctx)
	room := &gamev1alpha1.Room{}
	if err = r.Get(ctx, req.NamespacedName, room); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// 心跳上报与 ready 状态的对账
	ctl.RequeueAfter, err = r.ensureHeartbeat(ctx, room)
	if err != nil {
		return
	}
	return
}

const heartbeatTimeoutDuration = 10 * time.Second

func (r *RoomReconciler) ensureHeartbeat(ctx context.Context, room *gamev1alpha1.Room) (requeueAfter time.Duration, err error) {
	if ht := room.Status.LastHeartbeatTime; !ht.IsZero() { // 上报过心跳
		elapsed := time.Since(ht.Time)
		if elapsed > heartbeatTimeoutDuration { // 心跳超时
			if room.Status.Ready { // 且仍为 ready 状态，改成 not ready
				log.FromContext(ctx).Info("room heartbeat timeout, set to not ready")
				if room.Status.Ready {
					room.Status.Ready = false
					err = r.Status().Update(ctx, room)
					if err != nil {
						return
					}
				}
			}
		} else { // 心跳未超时
			if !room.Status.Ready { // 如果是 not ready，改成 ready
				room.Status.Ready = true
				log.FromContext(ctx).Info("set room status to ready")
				err = r.Status().Update(ctx, room)
				if err != nil {
					return
				}
			}
			requeueAfter = heartbeatTimeoutDuration - elapsed // 在超时的时间重新入队，以便心跳超时后能改成 not ready
		}
	} else if room.Status.Ready { // ready 状态但没有心跳，强行设为not ready
		room.Status.Ready = false
		log.FromContext(ctx).Info("ready status without heartbeat, set to not ready")
		err = r.Status().Update(ctx, room)
		if err != nil {
			return
		}
	}
	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *RoomReconciler) SetupWithManager(mgr ctrl.Manager) error {
	indexer := mgr.GetFieldIndexer()
	gamev1alpha1.IndexField(indexer)
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
