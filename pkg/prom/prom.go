package prom

import (
	"fmt"
	"net/http"
	"sync"

	gamev1alpha1 "github.com/imroc/tke-room-manager/api/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var roomNum = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "tke",
	Subsystem: "room",
	Name:      "num",
}, []string{
	"idle",
	"type",
})

var mux sync.Mutex

var (
	idleMap = make(map[string]*RoomInfo)
	busyMap = make(map[string]*RoomInfo)
)

type RoomInfo struct {
	Type string
}

func getRoomKey(room *gamev1alpha1.Room) string {
	return fmt.Sprintf("%s/%s", room.Namespace, room.Name)
}

func Delete(namespace, name string) {
	roomKey := fmt.Sprintf("%s/%s", namespace, name)
	mux.Lock()
	defer mux.Unlock()
	if ri, ok := idleMap[roomKey]; ok {
		delete(idleMap, roomKey)
		roomNum.WithLabelValues("true", ri.Type).Dec()
	}
	if ri, ok := busyMap[roomKey]; ok {
		delete(busyMap, roomKey)
		roomNum.WithLabelValues("false", ri.Type).Dec()
	}
}

func Count(room *gamev1alpha1.Room) {
	mux.Lock()
	defer mux.Unlock()
	idle := room.Status.Idle
	if idle && !room.Status.Ready {
		idle = false
	}
	roomKey := getRoomKey(room)
	if idle {
		if _, ok := idleMap[roomKey]; !ok { // 房间跳变为空闲状态
			idleMap[roomKey] = &RoomInfo{Type: room.Spec.Type}
			roomNum.WithLabelValues("true", room.Spec.Type).Inc()
			if _, ok := busyMap[roomKey]; ok {
				delete(busyMap, roomKey)
				roomNum.WithLabelValues("false", room.Spec.Type).Dec()
			}
		}
	} else {
		if _, ok := busyMap[roomKey]; !ok { // 房间跳变为忙碌状态
			busyMap[roomKey] = &RoomInfo{Type: room.Spec.Type}
			roomNum.WithLabelValues("false", room.Spec.Type).Inc()
			if _, ok := idleMap[roomKey]; ok {
				delete(idleMap, roomKey)
				roomNum.WithLabelValues("true", room.Spec.Type).Dec()
			}
		}
	}
}

func StartServer(addr string) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(addr, mux)
}
