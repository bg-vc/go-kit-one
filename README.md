## biz_rest: http restful api
* server
```shell script
cd $GOPATH/src/go-kit-one/biz_rest
go run *.go
```
* client
```shell script
curl -X POST -H "Content-Type:application/json" \
http://127.0.0.1:8000/biz/add/1/2
```

## biz_log: 日志功能
* server
```shell script
cd $GOPATH/src/go-kit-one/biz_log
go run *.go
```
* client
```shell script
curl -X POST -H "Content-Type:application/json" \
http://127.0.0.1:8000/biz/add/1/2
```

## biz_rate: 令牌桶算法限流(juju/ratelimit方案、gokit内置实现方案)
* server
```shell script
cd $GOPATH/src/go-kit-one/biz_rate
go run *.go
```
* client
```shell script
curl -X POST -H "Content-Type:application/json" \
http://127.0.0.1:8000/biz/add/1/2
```

## biz_monitor: 服务监控
* docker prometheus
```shell script
docker search prometheus
docker pull prom/prometheus
docker run --name vc-prometheus -d -p 9090:9090 prom/prometheus
```
```yaml
/etc/prometheus/prometheus.yaml
global:
    scrape_interval: 15s
    external_labels:
      monitor: 'vc-monitor'

scrape_configs:
  - job_name: 'prometheus'
    scrape_interval: 5s
    static_configs:
      - targets: ['localhost:9090']
        labels:
          group: 'local'

  - job_name: 'vc-biz-service'
    scrape_interval: 5s
    static_configs:
      - targets: ['192.168.65.2:8000']
        labels:
          group: 'biz-service'
```
* docker grafana
```shell script
docker search grafana
docker pull grafana/grafana
docker run --name vc-grafana -d -p 3000:3000 grafana/grafana
```
* server
```shell script
cd $GOPATH/src/go-kit-one/biz_monitor
go run *.go
```
* client
```shell script
curl -X POST -H "Content-Type:application/json" \
http://127.0.0.1:8000/biz/add/1/2
```
## biz_consul: 服务注册与发现
* docker consul
```shell script
dokcer search consul
docker pull consul
docker run --name vc-consul -d -p 8500:8500 consul
```
* server register
```shell script
cd $GOPATH/src/go-kit-one/biz_consul/register
go build
./register -consul.host localhost -consul.port 8500 -service.host 192.168.0.103 -service.port 8000
./register -consul.host localhost -consul.port 8500 -service.host 192.168.0.103 -service.port 8001
``` 

* server discover
```shell script
cd $GOPATH/src/go-kit-one/biz_consul/discover
go build
./discover -consul.host localhost -consul.port 8500
curl -XPOST -H "Content-Type:application/json" \
http://localhost:8002/biz \
-d'{"type":"add","a":1,"b":2}'
```
* gateway proxy
```shell script
cd $GOPATH/src/go-kit-one/biz_consul/gateway
go run *.go
curl -XPOST -H "Content-Type:application/json" \
http://127.0.0.1:8003/biz/biz/add/1/2 
```
## biz_trace: 服务链路跟踪
* docker zipkin
```shell script
docker search zipkin
dokcer pull openzipkin/zipkin
docker run --name vc-zipkin -d -p 9411:9411 openzipkin/zipkin
```
* server register
```shell script
cd $GOPATH/src/go-kit-one/biz_trace/register
go build
./register -consul.host localhost -consul.port 8500 -service.host 192.168.0.103 -service.port 8000
./register -consul.host localhost -consul.port 8500 -service.host 192.168.0.103 -service.port 8001
```
* gateway proxy 
```shell script
cd $GOPATH/src/go-kit-one/biz_trace/gateway
go run *.go
curl -XPOST -H "Content-Type:application/json" \
http://127.0.0.1:8003/biz/biz/add/1/2 
```
## biz_circuitbreaker: 服务熔断
* docker hystrix-dashboard
```shell script
docker search hystrix-dashboard
docker pull mlabouardy/hystrix-dashboard
docker run --name vc-hystrix-dashboard -d -p 9002:9002 mlabouardy/hystrix-dashboard
```
* server register
```shell script
cd $GOPATH/src/go-kit-one/biz_circuitbreaker/register
go build
./register -consul.host localhost -consul.port 8500 -service.host 192.168.0.103 -service.port 8000
./register -consul.host localhost -consul.port 8500 -service.host 192.168.0.103 -service.port 8001
```
* gateway proxy
```shell script
cd $GOPATH/src/go-kit-one/biz_circuitbreaker/gateway
go run *.go
curl -XPOST -H "Content-Type:application/json" \
http://127.0.0.1:8003/biz/biz/add/1/2
```

## biz_jwt: JWT身份认证
* server register
```shell script
cd $GOPATH/src/go-kit-one/biz_jwt/register
go build
./register -consul.host localhost -consul.port 8500 -service.host 192.168.0.103 -service.port 8000
./register -consul.host localhost -consul.port 8500 -service.host 192.168.0.103 -service.port 8001
```
* gateway proxy
```shell script
cd $GOPATH/src/go-kit-one/biz_jwt/gateway
go run *.go
curl -XPOST -H "Content-Type:application/json" \
http://127.0.0.1:8003/biz/login -d'{"name":"admin","pwd":"admin"}'
curl -XPOST -H "Content-Type:application/json" \
http://127.0.0.1:8003/biz/biz/add/1/2 \
-H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySUQiOiJhZG1pbiIsIm5hbWUiOiJhZG1pbiIsImV4cCI6MTU2ODEwMzg0NiwiaXNzIjoic3lzdGVtIn0.NjKW_DqfEhk9QHmgM0hsOCYxN6wX7O4by0h2NvhPP_c"
```