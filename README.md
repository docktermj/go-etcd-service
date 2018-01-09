# go-etcd-service

## Usage

A simple program to show how to integrate services.

### Invocation

```console
go-etcd-service
```

## Demonstrate #1

In this demonstration, show how 2 separate etcd instances are "automatically" joined to a cluster.

Because `etcd` may already be running on your system as a service, stop the service.  Example:

```console
sudo systemctl stop etcd.service
```

### Terminal #1

Create an embedded etcd agent.

Initialize environment.

```console
export GOPATH="${HOME}/go"
export PATH="${PATH}:${GOPATH}/bin:/usr/local/go/bin"
export PROJECT_DIR="${GOPATH}/src/github.com/docktermj"
export REPOSITORY_DIR="${PROJECT_DIR}/go-etcd-service"
export ETCD_CLIENT_ENDPOINTS=http://localhost:2379
export ETCD_PEER_ENDPOINTS=http://localhost:2380

rm -rf ${REPOSITORY_DIR}/localhost:2380.etcd
```

If needed, clone the project.

```console
mkdir -p ${PROJECT_DIR}
cd ${PROJECT_DIR}
git clone git@github.com:docktermj/go-etcd-service.git
```

Get prerequisistes.

```console
cd ${REPOSITORY_DIR}
make dependencies
```

Fix [bug](https://github.com/coreos/etcd/issues/8715).

```console
rm  ${REPOSITORY_DIR}/vendor/github.com/coreos/etcd/client/keys.generated.go
```

Compile program.

```console
cd ${REPOSITORY_DIR}
make
```

Run command to start etcd with client port of 2379 and peer port of 2380, which are defaults.

```console
cd ${REPOSITORY_DIR}
go-etcd-service \
  --client-endpoints ${ETCD_CLIENT_ENDPOINTS} \
  --peer-endpoints ${ETCD_PEER_ENDPOINTS}
```

### Terminal #2

Initialize environment.

```console
export GOPATH="${HOME}/go"
export PATH="${PATH}:${GOPATH}/bin:/usr/local/go/bin"
export PROJECT_DIR="${GOPATH}/src/github.com/docktermj"
export REPOSITORY_DIR="${PROJECT_DIR}/go-etcd-service"
export ETCD_CLIENT_ENDPOINTS=http://localhost:17073
export ETCD_PEER_ENDPOINTS=http://localhost:17074
export ETCD_CLUSTER_CLIENT_ENDPOINTS=http://localhost:2379

rm -rf ${REPOSITORY_DIR}/localhost:17074.etcd
```

Run command to start etcd with client port of 17073 and peer port of 17074.

```console
cd ${REPOSITORY_DIR}
go-etcd-service \
  --client-endpoints ${ETCD_CLIENT_ENDPOINTS} \
  --peer-endpoints ${ETCD_PEER_ENDPOINTS} \
  --cluster-client-endpoints ${ETCD_CLUSTER_CLIENT_ENDPOINTS}
```

### Terminal #3

Initialize environment.

```console
export GOPATH="${HOME}/go"
export PATH="${PATH}:${GOPATH}/bin:/usr/local/go/bin"
export PROJECT_DIR="${GOPATH}/src/github.com/docktermj"
export REPOSITORY_DIR="${PROJECT_DIR}/go-etcd-service"
export ETCD_CLIENT_ENDPOINTS=http://localhost:17075
export ETCD_PEER_ENDPOINTS=http://localhost:17076
export ETCD_CLUSTER_CLIENT_ENDPOINTS=http://localhost:2379

rm -rf ${REPOSITORY_DIR}/localhost:17076.etcd
```

Run command to start etcd with client port of 17075 and peer port of 17076.

```console
cd ${REPOSITORY_DIR}
go-etcd-service \
  --client-endpoints ${ETCD_CLIENT_ENDPOINTS} \
  --peer-endpoints ${ETCD_PEER_ENDPOINTS} \
  --cluster-client-endpoints ${ETCD_CLUSTER_CLIENT_ENDPOINTS}
```

### Terminal #4

Initialize environment.

```console
export GOPATH="${HOME}/go"
export PATH="${PATH}:${GOPATH}/bin:/usr/local/go/bin"
export ETCDCTL_API=3
export ETCD_CLIENT_ENDPOINT_1=http://localhost:2379
export ETCD_PEER_ENDPOINT_1=http://localhost:2380
export ETCD_CLIENT_ENDPOINT_2=http://localhost:17073
export ETCD_PEER_ENDPOINT_2=http://localhost:17074
export ETCD_CLIENT_ENDPOINT_3=http://localhost:17075
export ETCD_PEER_ENDPOINT_3=http://localhost:17076
```

View member lists from all etcd nodes.

```console
$ etcdctl --endpoint ${ETCD_CLIENT_ENDPOINT_1} member list
d567bb5c5a585dd: name=localhost:17074 peerURLs=http://localhost:17074 clientURLs=http://localhost:17073
8e646f365e5c4a9c: name=localhost:17076 peerURLs=http://localhost:17076 clientURLs=http://localhost:17075
8e9e05c52164694d: name=localhost:2380 peerURLs=http://localhost:2380 clientURLs=http://localhost:2379

$ etcdctl --endpoint ${ETCD_CLIENT_ENDPOINT_2} member list
d567bb5c5a585dd: name=localhost:17074 peerURLs=http://localhost:17074 clientURLs=http://localhost:17073
8e646f365e5c4a9c: name=localhost:17076 peerURLs=http://localhost:17076 clientURLs=http://localhost:17075
8e9e05c52164694d: name=localhost:2380 peerURLs=http://localhost:2380 clientURLs=http://localhost:2379

$ etcdctl --endpoint ${ETCD_CLIENT_ENDPOINT_3} member list
d567bb5c5a585dd: name=localhost:17074 peerURLs=http://localhost:17074 clientURLs=http://localhost:17073
8e646f365e5c4a9c: name=localhost:17076 peerURLs=http://localhost:17076 clientURLs=http://localhost:17075
8e9e05c52164694d: name=localhost:2380 peerURLs=http://localhost:2380 clientURLs=http://localhost:2379
```

Set a key/value pair on one etcd node and retrieve from another etcd node.

```console
$ curl -X PUT --data value="Hello world" ${ETCD_CLIENT_ENDPOINT_1}/v2/keys/message | jq
{
  "action": "set",
  "node": {
    "key": "/message",
    "value": "Hello world",
    "modifiedIndex": 8,
    "createdIndex": 8
  }
}

$ curl -X GET ${ETCD_CLIENT_ENDPOINT_1}/v2/keys/message | jq
{
  "action": "get",
  "node": {
    "key": "/message",
    "value": "Hello world",
    "modifiedIndex": 8,
    "createdIndex": 8
  }
}

$ etcdctl --endpoint ${ETCD_CLIENT_ENDPOINT_1} get /message
Hello world

$ etcdctl --endpoint ${ETCD_CLIENT_ENDPOINT_2} get /message
Hello world

$ etcdctl --endpoint ${ETCD_CLIENT_ENDPOINT_3} get /message
Hello world
```

Get version.

```console
$ curl -X GET ${ETCD_CLIENT_ENDPOINT_1}/version | jq
{
  "etcdserver": "3.2.12",
  "etcdcluster": "3.2.0"
}
```

Get health.

```console
$ curl -X GET ${ETCD_CLIENT_ENDPOINT_1}/health | jq
{
  "health": true
}
```

Find the leader.

```console
$ curl -X GET ${ETCD_CLIENT_ENDPOINT_1}/v2/stats/leader | jq
{
  "leader": "8e9e05c52164694d",
  "followers": {
    "2ae66c8b60ad88da": {
      "latency": {
        "current": 0.002674,
        "average": 0.006149571428571429,
        "standardDeviation": 0.0036228481241437995,
        "minimum": 0.002674,
        "maximum": 0.012785
      },
      "counts": {
        "fail": 0,
        "success": 7
      }
    }
  }
}
```

Am I the leader?

```console
$ curl -X GET ${ETCD_CLIENT_ENDPOINT_1}/v2/stats/self | jq
{
  "name": "localhost:2380",
  "id": "8e9e05c52164694d",
  "state": "StateLeader",
  "startTime": "2017-12-29T14:44:06.201700523-05:00",
  "leaderInfo": {
    "leader": "8e9e05c52164694d",
    "uptime": "15m11.554077768s",
    "startTime": "2017-12-29T14:44:07.102218352-05:00"
  },
  "recvAppendRequestCnt": 0,
  "sendAppendRequestCnt": 7
}

$ curl -X GET ${ETCD_CLIENT_ENDPOINT_2}/v2/stats/self | jq
{
  "name": "localhost:17074",
  "id": "2ae66c8b60ad88da",
  "state": "StateFollower",
  "startTime": "2017-12-29T14:49:56.394369884-05:00",
  "leaderInfo": {
    "leader": "8e9e05c52164694d",
    "uptime": "9m46.2818071s",
    "startTime": "2017-12-29T14:49:57.127177766-05:00"
  },
  "recvAppendRequestCnt": 7,
  "sendAppendRequestCnt": 0
}
```

### Untested

Join clusters

```console
$ curl -X GET http://localhost:${ETCD_CLIENT_PORT}/v2/members | jq

{
  "members": [
    {
      "id": "8e9e05c52164694d",
      "name": "go-2379",
      "peerURLs": [
        "http://localhost:2380"
      ],
      "clientURLs": [
        "http://localhost:2379"
      ]
    }
  ]
}

$ export DATA='{
  "peerUrls": [
    "http://localhost:'${ETCD_PEER_PORT2}'"
  ],
  "clientUrls": [
    "http://localhost:'${ETCD_CLIENT_PORT2}'"
  ]
}'

$ curl -X POST \
    --header "Content-Type: application/json" \
    --data "${DATA}" \
    http://localhost:${ETCD_CLIENT_PORT}/v2/members | jq

{
  "id": "703691f5941046d6",
  "name": "",
  "peerURLs": [
    "http://localhost:17074"
  ],
  "clientURLs": []
}

$ curl -X GET http://localhost:${ETCD_CLIENT_PORT}/v2/members | jq

{
  "members": [
    {
      "id": "4c28a520d0e9a9d6",
      "name": "",
      "peerURLs": [
        "http://localhost:17073"
      ],
      "clientURLs": []
    },
    {
      "id": "8e9e05c52164694d",
      "name": "go-2379",
      "peerURLs": [
        "http://localhost:2380"
      ],
      "clientURLs": [
        "http://localhost:2379"
      ]
    }
  ]
}
```

## Demonstrate #2

### Terminal #1

```console
export GOPATH="${HOME}/go"
export PATH="${PATH}:${GOPATH}/bin:/usr/local/go/bin"
export PROJECT_DIR="${GOPATH}/src/github.com/docktermj"
export REPOSITORY_DIR="${PROJECT_DIR}/go-etcd-service"
export ETCD_CLIENT_ENDPOINT_1=http://localhost:2379
export ETCD_PEER_ENDPOINT_1=http://localhost:2380
rm -rf ${REPOSITORY_DIR}/localhost:2380.etcd

cd ${REPOSITORY_DIR}
go-etcd-service \
  --client-endpoints ${ETCD_CLIENT_ENDPOINT_1} \
  --peer-endpoints ${ETCD_PEER_ENDPOINT_1}
```

### Terminal #2

```console
export GOPATH="${HOME}/go"
export PATH="${PATH}:${GOPATH}/bin:/usr/local/go/bin"
export PROJECT_DIR="${GOPATH}/src/github.com/docktermj"
export REPOSITORY_DIR="${PROJECT_DIR}/go-etcd-service"
export ETCD_CLIENT_ENDPOINT_2=http://localhost:17073
export ETCD_PEER_ENDPOINT_2=http://localhost:17074
rm -rf ${REPOSITORY_DIR}/localhost:17074.etcd

cd ${REPOSITORY_DIR}
go-etcd-service \
  --client-endpoints ${ETCD_CLIENT_ENDPOINT_2} \
  --peer-endpoints ${ETCD_PEER_ENDPOINT_2}
```

### Terminal #3

```console
export GOPATH="${HOME}/go"
export PATH="${PATH}:${GOPATH}/bin:/usr/local/go/bin"
export PROJECT_DIR="${GOPATH}/src/github.com/docktermj"
export REPOSITORY_DIR="${PROJECT_DIR}/go-etcd-service"
export ETCD_CLIENT_ENDPOINT_3=http://localhost:17075
export ETCD_PEER_ENDPOINT_3=http://localhost:17076
rm -rf ${REPOSITORY_DIR}/localhost:17076.etcd

cd ${REPOSITORY_DIR}
go-etcd-service \
  --client-endpoints ${ETCD_CLIENT_ENDPOINT_3} \
  --peer-endpoints ${ETCD_PEER_ENDPOINT_3}
```

### Terminal #4

Initialize environment.

```console
export GOPATH="${HOME}/go"
export PATH="${PATH}:${GOPATH}/bin:/usr/local/go/bin"
export PROJECT_DIR="${GOPATH}/src/github.com/docktermj"
export REPOSITORY_DIR="${PROJECT_DIR}/etcd-service"

export ETCDCTL_API=3
export ETCD_CLIENT_ENDPOINT_1=http://localhost:2379
export ETCD_PEER_ENDPOINT_1=http://localhost:2380
export ETCD_CLIENT_ENDPOINT_2=http://localhost:17073
export ETCD_PEER_ENDPOINT_2=http://localhost:17074
export ETCD_CLIENT_ENDPOINT_3=http://localhost:17075
export ETCD_PEER_ENDPOINT_3=http://localhost:17076
```

```console
```

## Development

### Dependencies

#### Set environment variables

```console
export GOPATH="${HOME}/go"
export PATH="${PATH}:${GOPATH}/bin:/usr/local/go/bin"
export PROJECT_DIR="${GOPATH}/src/github.com/docktermj"
export REPOSITORY_DIR="${PROJECT_DIR}/go-etcd-service"
```

#### Download project

```console
mkdir -p ${PROJECT_DIR}
cd ${PROJECT_DIR}
git clone git@github.com:docktermj/go-etcd-service.git
```

#### Download dependencies

```console
cd ${REPOSITORY_DIR}
make dependencies
```

Fix [bug](https://github.com/coreos/etcd/issues/8715).

```console
rm  ${REPOSITORY_DIR}/vendor/github.com/coreos/etcd/client/keys.generated.go
```

### Build

#### Local build

```console
cd ${REPOSITORY_DIR}
make
```

The results will be in the `${GOPATH}/bin` directory.

#### Docker build

```console
cd ${REPOSITORY_DIR}
make build
```

The results will be in the `.../target` directory.

### Test

```console
cd ${REPOSITORY_DIR}
make test-local
```

### Cleanup

```console
cd ${REPOSITORY_DIR}
make clean
```

### APIs

Get Swagger / Open API

```console
wget https://coreos.com/etcd/docs/3.2.11/dev-guide/apispec/swagger/rpc.swagger.json
```

1. Visit [swagger.io editor](https://editor.swagger.io/)
   1. File > Import File > {Choose downloaded rpc.swagger.json}

### References

1. [https://coreos.com/etcd/docs/3.2.11/index.html](https://coreos.com/etcd/docs/3.2.11/index.html)
   1. [https://godoc.org/github.com/coreos/etcd/embed](https://godoc.org/github.com/coreos/etcd/embed)
1. [Curl API](https://coreos.com/etcd/docs/latest/v2/api.html)
1. etcd [Swagger / Open API](https://coreos.com/etcd/docs/3.2.11/dev-guide/apispec/swagger/rpc.swagger.json)
   1. [In editor](http://editor.swagger.io/?url=https://coreos.com/etcd/docs/3.2.11/dev-guide/apispec/swagger/rpc.swagger.json)
1. Client options:
   1. [etcd/clientv3](https://github.com/coreos/etcd/blob/master/clientv3/README.md)
      1. [Go Doc](https://godoc.org/github.com/coreos/etcd/clientv3)
   1. [http requests / curl calls](https://coreos.com/etcd/docs/3.2.11/dev-guide/api_grpc_gateway.html)
      1. [v2](https://coreos.com/etcd/docs/latest/v2/api.html)
      1. [cluster management](https://coreos.com/etcd/docs/latest/v2/members_api.html)
      1. [examples](https://www.digitalocean.com/community/tutorials/how-to-use-etcdctl-and-etcd-coreos-s-distributed-key-value-store#etcd-httpjson-api-usage)
