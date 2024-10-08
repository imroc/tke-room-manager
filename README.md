# tke-room-manager

TKE 房间管理器，用于游戏战斗服、会议等房间类场景，可支持单 Pod 多房间的管理。

## 与 OpenKruiseGame 联动

对于游戏场景，可自动联动 [OpenKruiseGame](https://openkruise.io/zh/kruisegame/introduction/)，在业务发版更新或缩容时，优先删除所有房间都空闲的 Pod，避免占用中的房间被中断，实现不停服更新和丝滑的弹性伸缩。

## 根据房间占用比例自动伸缩

提供了房间信息的 Prometheus 监控指标:

```promql
# HELP tke_room_num
# TYPE tke_room_num gauge
tke_room_num{idle="false",type="fight"} 4
tke_room_num{idle="true",type="fight"} 4
```

可通过 KEDA 配置 Prometheus 触发器的 `ScaledObject` 来实现根据房间占用比例自动伸缩:

```yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: fight-king-scaledobject
  namespace: fight-king
spec:
  scaleTargetRef:
    apiVersion: game.kruise.io/v1alpha1
    kind: GameServerSet
    name: fight-king
  pollingInterval: 5
  minReplicaCount: 1
  maxReplicaCount: 100
  triggers:
    - type: prometheus
      metadata:
        serverAddress: http://vmsingle-monitoring-victoria-metrics-k8s-stack.monitoring.svc.cluster.local:8429
        query: |
          tke_room_num{type="fight", idle="false"} / sum(tke_room_num{type="fight"})
        threshold: "0.7" # 房间占用比扩缩容阈值：70%

```

## API 接入

### 通用路径参数说明

| 参数      | 说明                                           |
| --------- | ---------------------------------------------- |
| namespace | 命名空间（可通过Downward API从环境变量中获取） |
| pod       | Pod 名称（可通过Downward API从环境变量中获取） |
| id        | 房间 ID（通常为Pod中的房间序号，如0,1,2,3）    |

### 注册房间信息

```txt
POST /api/room/{namespace}/{pod}/{id}

{
  "externalAddress": "2.2.2.2:9889",
  "type": "fight"
}
```

请求体参数说明：

| 参数            | 说明                                                            |
| --------------- | --------------------------------------------------------------- |
| externalAddress | 房间对外暴露的地址（通常Pod通过Downward API获取自身的外部地址） |
| type            | 房间类型（如有多个游戏，或游戏分多种类型房间，通过此字段区分）  |


### 更新房间状态(是否空闲)

```txt
PUT /api/room/{namespace}/{pod}/{id}/status

{
  "idle": true
}
```

请求体参数说明：

| 参数 | 说明     |
| ---- | -------- |
| idle | 是否空闲 |

1. 房间就绪后需上报房间为空闲状态，以待匹配时被分配给玩家。
2. 在游戏或会议结束后，如后面还要复用该房间，需再次调用此接口上报房间状态为空闲状态。

### 上报心跳

```txt
PUT /api/room/{namespace}/{pod}/{id}/heartbeat
```

说明：
1. 房间进程需定期上报心跳，以表明房间进程存活，否则会自动标记房间为不健康，获取空闲房间时将会被排除在外。
2. 心跳超时时间默认为10秒。

### 获取空闲房间

```txt
GET /api/room/idle/{namespace}/{type}/{num}
```

路径参数说明：

| 参数 | 说明               |
| ---- | ------------------ |
| type | 房间类型           |
| num  | 需要获取的房间数量 |


说明：
1. 最终返回的数量小于等于 `num`，在空闲房间不足时会小于 `num`，配置了基于房间占用比例的伸缩策略时，会自动扩容，业务侧可不断重试以获取新扩出的空闲房间。
2. 房间被返回后会被自动标记为占用状态，业务侧需在游戏或会议结束后，如需后续继续复用房间，可调用更新房间状态的接口将房间重置为空闲状态。

### 注销房间

```txt
DELETE /api/room/{namespace}/{pod}/{id}
```

说明：
1. 通常只用于房间占用结束后，自动启动新房间进程来替代的场景，需在游戏或会议结束后，调此接口注销房间，新房间进程启动后再调用注册房间的接口注册新房间。
2. Pod 被删除会自动注销Pod所关联的所有房间，无需调用此接口来注销。
