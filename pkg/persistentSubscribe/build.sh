CGO_ENABLED=0 go build main.go
docker rmi event-redirect:v1
docker build -t event-redirect:v1 .
docker tag event-redirect:v1 ccr.ccs.tencentyun.com/k8s-test/auth:event-redirect-v3
docker push ccr.ccs.tencentyun.com/k8s-test/auth:event-redirect-v3