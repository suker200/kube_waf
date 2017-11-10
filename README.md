# Kubernetes Web Application Firewall

# Diagram flow : kube_waf_2.png

![alt text](https://raw.githubusercontent.com/suker200/kube_waf/master/kube_waf_2.png)

# Requirement:
 - k8s 1.7
 - k8s 1.8 (in testing): we use library from kube-client so I think it will be fine :)

# Reason:
 - We want to apply custom Nginx WAF rule/Nginx config in front of the Kubernetes Ingress Controller for all virtual host
 - We want to support multi Certs (we're running in AWS, and ELB just support one Cert when we decide support this feature, so we use TCP/Proxy Proxy-protcol to keep ClientIP). AWS Support Multi certs now.
 - We want to support custom nginx config per virtual domain, domain redirect etc.
 - Zero downtime deploy
 - We want to write own WAF rule beside default WAF rule

# Elements:
 - Nginx WAF core https://github.com/p0pr0ck5/lua-resty-waf (Big thanks to p0pr0ck5)
 - Nginx Config Watcher: watching and generate Certs + Nginx Config
 - k8s CRD (Custom Resource Definition)

# Target:
- Certs must be generate base on CRD update/edit/delete (k8s Custom Resource Definition) for Nginx WAF
- Support custom config base on CRD
- Zero Downtime in deploying/updating via Nginx configtest + reload + k8s deployment
- Run as a deployment in kubernetes
- Nginx Gateway for K8s ingress controller
- Apply WAF rule for all request to k8s ingress controller


# How to:
- Healtcheck endpoint: TCP 9999
	+ When k8s update/delete pod, it's run post start script, which notify the nginx-config-watcher pod going down, and sleep a period time (we defined in kubernetes helm chart: terminationGracePeriodSeconds)
	+ nginx-config-watcher disable listen on TCP 9999
	+ ELB (Nginx) health check over TCP 9999 recieves refuse response from TCP 9999, and detach this pod from ELB

- Cause of when ELB wasn't support multi certs, so we run ELB with Proxy-Protocol, so we must enable it via CRD when config domain point to this ELB. You can check the config in chart folder

- list CRD info (list certs info):
	+ kubectl -n devops get nginxcerts

Note: in charts/crd.yaml + charts/values.yaml we defined namespace = devops

# Usage:
- Build image
- Create CRD resource
- Deploy kubernetes
- Register Certs
	+ sh scripts/gen_cert.sh test_cert test.suker200.com
	+ cert_base64=$(cat test_cert_KEY.pem >> test_cert.pem && cat test_cert.pem | base64 "
	+ update $(cert_base64) content to scripts/test_cert_secret.yaml
	+ kubect apply -f  scripts/test_cert_secret.yaml
	+ update host nginx config scripts/domain_config.yaml
	+ kubectl apply -f scripts/domain_config.yaml
- Update your service with k8s ingress as normal

# Build

- Build kube_waf

```
glide up --strip-vendor

CGO_ENABLED=0 env GOOS=linux go build

```

- Build kube_waf docker image

```
cp kube_waf ./nginx-loadbalancer/nginx-config-watcher && cd ./nginx-loadbalancer

docker build -t kube_waf -f Dockerfile_watcher
```

- Create k8s CRD:

```
kubectl create -f charts/crd.yaml
```

- Update helm chart (We using helm chart for deploying application). You can found it in charts folder

- we can test with minikube :)
