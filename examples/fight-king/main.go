package main

import (
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/imroc/req/v3"
)

var (
	roomServerAddr, namespace, podName string
	alive                              map[int]bool = make(map[int]bool)
	mux                                sync.Mutex
)

func setAlive(id int, a bool) {
	mux.Lock()
	defer mux.Unlock()
	alive[id] = a
}

func main() {
	roomServerAddr = os.Getenv("ROOM_SERVER_ADDR")
	if roomServerAddr == "" {
		panic("ROOM_SERVER_ADDR is not set")
	}
	namespace = os.Getenv("POD_NAMESPACE")
	podName = os.Getenv("POD_NAME")
	if namespace == "" || podName == "" {
		panic("POD_NAMESPACE or POD_NAME is not set")
	}
	for i := 0; i < 4; i++ {
		// 启动4个游戏房间
		go startFightRoom(i)
	}
	http.HandleFunc("GET /idle/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("id is empty"))
			return
		}
		intId, err := strconv.Atoi(id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("id is not a number"))
			return
		}
		idle := r.URL.Query().Get("idle")
		if idle == "false" {
			setIdleState(intId, false)
		} else {
			setIdleState(intId, true)
		}
	})

	http.HandleFunc("GET /stop/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("id is empty"))
			return
		}
		intId, err := strconv.Atoi(id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("id is not a number"))
			return
		}
		stopFightRoom(intId)
	})

	http.HandleFunc("GET /stopbeat/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("id is empty"))
			return
		}
		intId, err := strconv.Atoi(id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("id is not a number"))
			return
		}
		slog.Info("stop heartbeat", "id", id)
		setAlive(intId, false)
	})

	http.HandleFunc("GET /start/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("id is empty"))
			return
		}
		intId, err := strconv.Atoi(id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("id is not a number"))
			return
		}
		startFightRoom(intId)
	})

	slog.Info("server started at :80")
	if err := http.ListenAndServe(":80", nil); err != nil {
		panic(err)
	}
}

// 启动指定id的游戏房间
func startFightRoom(id int) error {
	slog.Info("start room", "id", id)
	// 注册房间
	if err := registerRoom(id); err != nil {
		return err
	}
	setAlive(id, true)
	time.Sleep(2 * time.Second)
	// 上报心跳
	go heartbeat(id)
	// 房间准备就绪，设置为空闲
	setIdleState(id, true)
	return nil
}

// 停止指定id的游戏房间
func stopFightRoom(id int) {
	slog.Info("stop room", "id", id)
	setAlive(id, false)
	deleteApiAddr := fmt.Sprintf("%s/api/room/%s/%s/%d", roomServerAddr, namespace, podName, id)
	_, err := req.R().Delete(deleteApiAddr)
	if err != nil {
		slog.Error("faled to delete room", "id", id, "error", err.Error())
	}
}

func fakeIpPort() string {
	ip := []string{}
	for i := 0; i < 4; i++ {
		n := rand.Intn(256)
		ip = append(ip, strconv.Itoa(n))
	}
	ret := strings.Join(ip, ".")
	port := rand.Intn(65535) + 1
	ret += ":" + strconv.Itoa(port)
	return ret
}

func registerRoom(id int) error {
	slog.Info("register room", "id", id)
	registerApiAddr := fmt.Sprintf("%s/api/room/%s/%s/%d", roomServerAddr, namespace, podName, id)
	var body struct {
		ExternalAddress string `json:"externalAddress"`
		Type            string `json:"type"`
	}
	body.ExternalAddress = fakeIpPort()
	body.Type = "fight"
	_, err := req.R().SetBodyJsonMarshal(&body).Post(registerApiAddr)
	if err != nil {
		slog.Error("failed to register room", "error", err.Error(), "id", id, "body", body)
		return err
	}
	return nil
}

func setIdleState(id int, idle bool) {
	slog.Info("set room idle state", "id", id, "idle", idle)
	statusApiAddr := fmt.Sprintf("%s/api/room/%s/%s/%d/status", roomServerAddr, namespace, podName, id)
	var status struct {
		Idle bool `json:"idle"`
	}
	status.Idle = idle
	resp, err := req.R().SetBodyJsonMarshal(&status).Put(statusApiAddr)
	if err != nil {
		slog.Error("failed to set idle status", "id", id, "idle", idle, "error", err.Error())
		return
	}
	if resp.StatusCode != 200 {
		slog.Error("failed to set idle status", "id", id, "idle", idle, "error", resp.String(), "status", resp.StatusCode)
	}
}

func heartbeat(id int) {
	heartbeatApiAddr := fmt.Sprintf("%s/api/room/%s/%s/%d/heartbeat", roomServerAddr, namespace, podName, id)
	for {
		if _, ok := alive[id]; !ok {
			break
		}
		slog.Info("heartbeat", "id", id)
		_, err := req.Put(heartbeatApiAddr)
		if err != nil {
			slog.Error("failed to heartbeat", "id", id, "error", err.Error())
		}
		time.Sleep(7 * time.Second)
	}
}
