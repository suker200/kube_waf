#!/bin/bash

minikube start --memory=4096 --kubernetes-version v1.8.5 --bootstrapper kubeadm

# this for loop waits until kubectl can access the api server that Minikube has created
for i in {1..300}; do # timeout for 5 minutes
	kubectl get po &> /dev/null
	if [ $? -ne 1 ]; then
		break
	fi
	
	sleep 2
	
	if [ $i -eq 300 ]; then
		echo "Minikube start timeout"
		exit 1
	fi
done

# init helm
kubectl -n kube-system create sa tiller
kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller
helm init --service-account tiller

for i in {1..150}; do
	helm list &> /dev/null
	if [ $? -ne 1 ]; then
		break
	fi
	
	sleep 2

	if [ $i -eq 150 ]; then
		echo "helm init failed"
		exit 1
	fi
done


# Create CRD 
kubectl apply -f crd.yaml

# Deploy kube_waf
helm upgrade -i --namespace=kube-system nginx-ingress-controller-proxy-protocol nginx-ingress-controller-proxy-protocol

for i in {1..300}; do
	kubectl get pods -n kube-system -l "app=nginx-ingress-controller-proxy-protocol-test" | grep -i "running" | grep "2/2" &> /dev/null
	if [ $? -ne 1 ]; then
		echo "Deploy ingress successfully"
		break
	fi

	sleep 2

	if [ $i -eq 150 ]; then
		echo "Deploy ingress failed"
		exit 1
	fi

done

# Gen test cert
sh gen_secret_cert_test.sh test_cert test.suker200.com

kubectl apply -f cert_secret.yaml

# Deploy nginx application
kubectl apply -f nginx.yaml



for i in {1..300}; do
	kubectl -n devops get po -l app=nginx | grep Running &> /dev/null
	if [ $? -ne 1 ]; then
		break
	fi
	
	sleep 2

	if [ $i -eq 300 ]; then
		echo "Deploy nginx application failed"
		exit 1
	fi
done


for i in {1..10}; do
	curl -s -I $(minikube service -n devops nginx --url) | grep "HTTP/1.1 200 OK"
	if [ $? -eq 0 ]; then
		echo "Deploy nginx Successfully"
		break
	fi

	sleep 2

	if [ $i -eq 10 ]; then
		echo "Deploy nginx application failed"
		exit 1
	fi
done

# Test nginx application with waf
waf_endpoint=$(minikube service -n kube-system nginx-ingress-controller-proxy-protocol-test --url | cut -d '/' -f3)

curl -s -k -I -H Host:test.suker200.com https://$waf_endpoint | grep "HTTP/1.1 200 OK"
if [ $? -eq 0 ]; then
	echo "Deploy nginx with WAF Successfully"
else
	echo "Test Nginx with WAF Failed"
	exit 1
fi

# Check certificate 
openssl s_client -showcerts -servername test.suker200.com -connect $waf_endpoint </dev/null | grep "CN=test.suker200.com"
if [ $? -eq 0 ]; then
	echo "Test Certificate test.suker200.com Successfully"
else
	echo "Test Certs with WAF Failed"
	exit 1
fi

