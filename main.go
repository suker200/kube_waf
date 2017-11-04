package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	// "net/http"
	"net"
	"time"
	// "k8s.io/apimachinery/pkg/runtime"
	// apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	// apierrors "k8s.io/apimachinery/pkg/api/errors"
	// "encoding/json"
	// "github.com/imdario/mergo"
	// "github.com/jinzhu/copier"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	rest "k8s.io/client-go/rest"
	// "k8s.io/kubernetes/pkg/api"
	// "reflect"
	// k8sapiV1 "k8s.io/kubernetes/pkg/api/v1"
	// clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	// "k8s.io/client-go/kubernetes"
)

// var NginxConfig_Backup *Nginx_Config

// func (s *Nginx_Server_Config) ServerInfo() {
// 	var serverInfo
// }

// var Client *clientset.Clientset

const (
	terminatedFile = "/tmp/terminated"
)

func APISOutClusterConfig() *rest.Config {
	certData, _ := ioutil.ReadFile("/data/suker/git/minikube/.minikube/apiserver.crt")

	keyData, _ := ioutil.ReadFile("/data/suker/git/minikube/.minikube/apiserver.key")

	config := &rest.Config{
		Host:    "https://192.168.99.100:8443",
		APIPath: "/apis",
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
			CertFile: "/data/suker/git/minikube/.minikube/apiserver.crt",
			KeyFile:  "/data/suker/git/minikube/.minikube/apiserver.key",
			CertData: certData,
			KeyData:  keyData,
		},
		ContentConfig: rest.ContentConfig{
			GroupVersion: &schema.GroupVersion{
				Group:   "suker200.com",
				Version: "v1",
			},
		},
	}

	return config
}

func OutClusterConfig() *rest.Config {
	certData, _ := ioutil.ReadFile("/data/suker/git/minikube/.minikube/apiserver.crt")

	keyData, _ := ioutil.ReadFile("/data/suker/git/minikube/.minikube/apiserver.key")

	config := &rest.Config{
		Host:    "https://192.168.99.100:8443",
		APIPath: "/api",
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
			CertFile: "/data/suker/git/minikube/.minikube/apiserver.crt",
			KeyFile:  "/data/suker/git/minikube/.minikube/apiserver.key",
			CertData: certData,
			KeyData:  keyData,
		},
		ContentConfig: rest.ContentConfig{
			GroupVersion: &schema.GroupVersion{
				Version: "v1",
			},
		},
	}

	return config
}

func APISInClusterConfig() *rest.Config {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	config.APIPath = "/apis"

	config.ContentConfig = rest.ContentConfig{
		GroupVersion: &schema.GroupVersion{
			Group:   CRD_Group,
			Version: CRD_Group_Version,
		},
	}

	return config
}

func InClusterConfig() *rest.Config {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	config.APIPath = "/api"

	config.ContentConfig = rest.ContentConfig{
		GroupVersion: &schema.GroupVersion{
			// Group:   "v1",
			Version: "v1",
		},
	}

	return config
}

var TestMode bool
var CRD_Namespace string
var CRD_Plural string
var CRD_Kind string
var CRD_Group string
var CRD_Group_Version string
var WAFName string

var APISConfig *rest.Config
var Config *rest.Config
var ApisClient *dynamic.Client
var Client *dynamic.Client
var Config_Info *ConfigInfo
var TCPSock net.Listener

// var ClientCMD *clientset.Clientset

func ConfigWorker(c chan string) {
	fmt.Println("We start ConfigWorker")
	var resourceClient *dynamic.ResourceClient
	resource := &metav1.APIResource{
		Name:       CRD_Plural, //"nginxcerts",
		Kind:       CRD_Kind,   // "Nginxcert",
		Namespaced: true,
		Verbs:      metav1.Verbs{"get"},
	}
	resourceClient = ApisClient.Resource(resource, CRD_Namespace)
	// Init(resourceClient)
	ConfigWatch(resourceClient, c)
}

func CertWorker(c chan string) {
	fmt.Println("We start CertWorker")
	var resourceClient *dynamic.ResourceClient
	resource := &metav1.APIResource{
		Name:       "secrets", //"nginxcerts",
		Kind:       "secret",  // "Nginxcert",
		Namespaced: true,
		Verbs:      metav1.Verbs{"get"},
	}

	resourceClient = Client.Resource(resource, CRD_Namespace)
	if !CertInitRunOnce {
		CertInit(resourceClient)
	}
	CertWatch(resourceClient, c)
}

func Worker() {
	// Start CRD Watch
	var err error
	cInfo := make(chan string)

	if ApisClient, err = dynamic.NewClient(APISConfig); err != nil {
		panic(err)
	}
	if Client, err = dynamic.NewClient(Config); err != nil {
		panic(err)
	}

	// ClientCMD = clientset.NewForConfigOrDie(Config)

	go ConfigWorker(cInfo)
	go CertWorker(cInfo)

	for {
		c := <-cInfo
		if c == "ConfigWatch" {
			go ConfigWorker(cInfo)
		} else if c == "CertWatch" {
			go CertWorker(cInfo)
		}
		time.Sleep(time.Duration(1) * time.Second)
	}
}

func ConfigLoad() {
	data, err := ioutil.ReadFile("/config/config.yaml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(data, &Config_Info)
	if err != nil {
		panic(err)
	}
}

func CheckTerminated() bool {
	if _, err := os.Stat(terminatedFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func HealthCheckIngress() {
	for {
		fmt.Println("We start HealthChecking Ingress")
		resp, err := http.Get("http://127.0.0.1:10254/healthz")
		if err == nil {
			if resp.StatusCode == 200 {
				go TcpListen()
				break
			}
		}
		time.Sleep(time.Duration(1) * time.Second)
	}

}

func TcpListen() {
	var err error
	TCPSock, err = net.Listen("tcp4", ":9999")
	if err != nil {
		panic(err)
	}

	for {
		_, err := TCPSock.Accept()
		if err != nil {
			panic(err)
		}

		if ok := CheckTerminated(); ok {
			fmt.Print("We shutdown socket for terminated process")

			if err := TCPSock.Close(); err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println("The socket stopped")
				fmt.Println(TCPSock)
				break
			}
		}
	}
}

func main() {
	ConfigLoad()
	local_test := flag.Bool("test", false, "local test mode")
	waf_name := flag.String("waf_name", "waf-abc", "waf_name for applying change config")
	flag.Parse()

	fmt.Println(Config_Info)
	CRD_Namespace = Config_Info.CRD_Info.Namespace
	CRD_Kind = Config_Info.CRD_Info.Kind
	CRD_Plural = Config_Info.CRD_Info.Plural
	CRD_Group = Config_Info.CRD_Info.Group
	CRD_Group_Version = Config_Info.CRD_Info.Version
	WAFName = Config_Info.CRD_Info.Waf_Name

	// var err error
	// var config *rest.Config
	if *local_test {
		APISConfig = APISOutClusterConfig()
		Config = OutClusterConfig()
		TestMode = true
		CRD_Namespace = "devops"
		CRD_Kind = "Nginxcert"
		CRD_Plural = "nginxcerts"
		CRD_Group = "suker200.com"
		CRD_Group_Version = "v1"
		WAFName = *waf_name
	} else {
		APISConfig = APISInClusterConfig()
		Config = InClusterConfig()

	}

	// go TcpListen()
	go HealthCheckIngress()
	Worker()

	// time.Sleep(time.Duration(10) * time.Second)
	// os.Exit(100)
	// ApisClient, err := dynamic.NewClient(APISConfig)

	// if err != nil {
	// 	panic(err)
	// } else {
	// 	var resourceClient *dynamic.ResourceClient
	// 	resource := &metav1.APIResource{
	// 		Name:       CRD_Plural, //"nginxcerts",
	// 		Kind:       CRD_Kind,   // "Nginxcert",
	// 		Namespaced: true,
	// 		Verbs:      metav1.Verbs{"get"},
	// 	}
	// 	resourceClient = ApisClient.Resource(resource, CRD_Namespace)
	// 	Init(resourceClient)
	// 	for {
	// 		Watch(resourceClient)
	// 		time.Sleep(time.Duration(1) * time.Second)
	// 	}

	// }
}
