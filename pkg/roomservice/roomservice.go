package roomservice

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	gamev1alpha1 "github.com/imroc/tke-room-manager/api/v1alpha1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var log = ctrl.Log.WithName("roomservice")

type RoomService struct {
	cluster.Cluster
	client.Client
	*runtime.Scheme
}

func New(cls cluster.Cluster, scheme *runtime.Scheme) (*RoomService, error) {
	if cls == nil {
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		c, err := cluster.New(config)
		if err != nil {
			return nil, err
		}
		cls = c
	}
	return &RoomService{
		Cluster: cls,
		Client:  cls.GetClient(),
		Scheme:  scheme,
	}, nil
}

func getRoomParamFromRequest(r *http.Request) (namespace, pod, id string, err error) {
	namespace = r.PathValue("namespace")
	pod = r.PathValue("pod")
	id = r.PathValue("id")
	if namespace == "" || pod == "" || id == "" {
		err = errors.New("namespace, pod or id is empty")
		return
	}
	return
}

func (rs *RoomService) getRoomFromRequest(r *http.Request, fromClient bool) (*gamev1alpha1.Room, error) {
	namespace, pod, id, err := getRoomParamFromRequest(r)
	if err != nil {
		return nil, errors.New("namespace, pod or id is empty")
	}
	room := &gamev1alpha1.Room{}
	name := getRoomName(pod, id)
	if fromClient {
		if err := rs.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, room); err != nil {
			return nil, err
		}
	} else {
		room.Name = name
		room.Namespace = namespace
		room.Spec.PodName = pod
	}
	return room, nil
}

func (rs *RoomService) GetIdleRoomsExternalAddress(namespace, tp string, num int) (addr []string, err error) {
	list := &gamev1alpha1.RoomList{}
	err = rs.List(
		context.Background(), list,
		client.InNamespace(namespace),
		client.MatchingFields{"spec.type": tp},
		client.MatchingFields{"status.idle": "true"},
		client.MatchingFields{"status.ready": "true"},
	)
	if err != nil {
		return
	}
	needUpdate := []*gamev1alpha1.Room{}
	for _, room := range list.Items {
		addr = append(addr, room.Spec.ExternalAddress)
		num--
		idle := false
		room.Status.Idle = &idle
		needUpdate = append(needUpdate, &room)
		if num == 0 {
			break
		}
	}
	if len(needUpdate) > 0 { // TODO: 考虑失败的情况
		go func() {
			for _, room := range needUpdate {
				if err := rs.Status().Update(context.Background(), room); err != nil {
					log.Error(err, "failed to update room status", "room", room.Name, "namespace", room.Namespace, "status", room.Status)
				}
			}
		}()
	}
	return
}

func (rs *RoomService) AddHttpRoute(mux *http.ServeMux) error {
	// 注册房间信息，上报外部地址信息
	mux.HandleFunc("POST /api/room/{namespace}/{pod}/{id}", func(w http.ResponseWriter, r *http.Request) {
		namespace, podName, id, err := getRoomParamFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		pod := &corev1.Pod{}
		if err := rs.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: podName}, pod); err != nil {
			if apierrors.IsNotFound(err) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("pod not found"))
				return
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			}
			return
		}
		room := &gamev1alpha1.Room{}
		name := getRoomName(podName, id)
		room.Name = name
		room.Namespace = namespace
		room.Spec.PodName = podName
		var body struct {
			ExternalAddress string `json:"externalAddress"`
			Type            string `json:"type"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		room.Spec.ExternalAddress = body.ExternalAddress
		room.Spec.Type = body.Type
		if err := controllerutil.SetOwnerReference(pod, room, rs.Scheme); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if err := rs.Create(context.Background(), room); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	})
	// 更新房间状态（是否空闲、是否ready）
	mux.HandleFunc("PUT /api/room/{namespace}/{pod}/{id}/status", func(w http.ResponseWriter, r *http.Request) {
		room, err := rs.getRoomFromRequest(r, true)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		var status gamev1alpha1.RoomStatus
		if err := json.Unmarshal(body, &status); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		needUpdate := false
		if status.Idle != nil {
			needUpdate = true
			room.Status.Idle = status.Idle
		}
		if status.Ready != nil {
			needUpdate = true
			room.Status.Ready = status.Ready
		}
		if needUpdate {
			rs.Status().Update(context.Background(), room)
		}
	})
	// 注销房间
	mux.HandleFunc("DELETE /api/room/{namespace}/{pod}/{id}", func(w http.ResponseWriter, r *http.Request) {
		room, err := rs.getRoomFromRequest(r, false)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if err := rs.Delete(context.Background(), room); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	})
	// 心跳上报
	mux.HandleFunc("PUT /api/room/{namespace}/{pod}/{id}/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		room, err := rs.getRoomFromRequest(r, false)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		room.Status.LastHeartbeatTime = &metav1.Time{Time: time.Now()}
		if err := rs.Status().Update(context.Background(), room); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	})
	// 获取空闲房间
	mux.HandleFunc("GET /api/room/idle/{namespace}/{type}/{num}", func(w http.ResponseWriter, r *http.Request) {
		namespace := r.PathValue("namespace")
		if namespace == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("namespace is empty"))
			return
		}
		roomType := r.PathValue("type")
		if roomType == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("type is empty"))
			return
		}
		num, err := strconv.Atoi(r.PathValue("num"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("num should be integer"))
			return
		}
		addrs, err := rs.GetIdleRoomsExternalAddress(namespace, roomType, num)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if err := json.NewEncoder(w).Encode(addrs); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	})
	return nil
}

func getRoomName(pod, id string) string {
	return pod + "-" + id
}
