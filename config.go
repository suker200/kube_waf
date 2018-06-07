package main

import (
	"fmt"
	// "io/ioutil"
	// "k8s.io/api/core/v1"
	"encoding/json"
	"github.com/imdario/mergo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"os"
	"github.com/fsnotify/fsnotify"
	// rest "k8s.io/client-go/rest"
	// k8sapi "k8s.io/kubernetes/pkg/api"
	// "k8s.io/apimachinery/pkg/watch"
	// "k8s.io/client-go/pkg/api"
)

// func Get(rclient *dynamic.ResourceClient) {
// 	a, err := rclient.Get("build.int.suker200.com", metav1.GetOptions{})
// 	if err != nil {
// 		fmt.Println("Failed")
// 		fmt.Println(err.Error())
// 	} else {
// 		// fmt.Println(a.Object)
// 		data, err := a.MarshalJSON()
// 		if err != nil {
// 			fmt.Println(err.Error())
// 		} else {
// 			// fmt.Println(data)
// 			var content CRD
// 			if err := json.Unmarshal(data, &content); err != nil {
// 				fmt.Println(err.Error())
// 			} else {
// 				// fmt.Println(content.Spec.Domains)
// 				// server2 := &Nginx_Server_Config{}
// 				fmt.Println(server)
// 				err := mergo.MergeWithOverwrite(&server, content.Spec)
// 				if err != nil {
// 					fmt.Println(err.Error())
// 				}
// 				fmt.Println(server)
// 				cfg.Servers = append(cfg.Servers, server)
// 				cfg.GenerateConfig()
// 			}
// 		}
// 	}
// }

func ConfigWatch(rclient *dynamic.ResourceClient, c chan string) {
	timeoutSecond := int64(60)
	a, err := rclient.Watch(metav1.ListOptions{Watch: true, TimeoutSeconds: &timeoutSecond, ResourceVersion: "0", LabelSelector: "waf_name=" + WAFName})
	if err != nil {
		fmt.Println("hello failed error")
		fmt.Println(err.Error())
		fmt.Println("hello failed error")
	} else {
		fmt.Println(a)
		b := a.ResultChan()
		for {
			fmt.Println("new start")
			msg := <-b
			if msg.Type != "" {
				if msg.Type != "ERROR" {
					fmt.Println(msg)
					ConfigCreate(rclient)
				}
			} else {
				// Cause of http session is closed, re-new session
				// return
				c <- "ConfigWatch"
				return
			}
		}
	}
}

func (cfg *Nginx_Config) ConfigList(rclient *dynamic.ResourceClient) []CRD {
	var content CRDList
	a, err := rclient.List(metav1.ListOptions{})
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("====")
		b := a.GetObjectKind()
		d := b.(*unstructured.UnstructuredList)
		data, err := d.MarshalJSON()
		if err != nil {
			fmt.Println("Do something when errors appear") // telegram and rollback
			fmt.Println(err.Error())
			// panic(err)
		} else {
			if err := json.Unmarshal(data, &content); err != nil {
				fmt.Println(err.Error())
				fmt.Println("Do something when errors appear")
			}
		}
	}
	return content.Items
}

func ConfigCreate(rclient *dynamic.ResourceClient) {
	var cfg Nginx_Config
	var server Nginx_Server_Config
	cfg.Http = InitDefaultHTTPConfig()
	server = InitDefaultServerConfig()
	cfg.Servers = append(cfg.Servers, server)

	var crdList []CRD
	crdList = cfg.ConfigList(rclient)
	for _, crd := range crdList {
		if crd.Spec.WAFName == WAFName {
			// fmt.Println(crd)
			if crd.Spec.BlockType == "http" {
				if err := mergo.MergeWithOverwrite(&cfg.Http, crd.Spec.Http); err != nil {
					fmt.Println(err.Error())
					fmt.Println("Do something when errors appear") // telegram notification and bypass
				}
			} else if crd.Spec.BlockType == "server" {
				if len(crd.Spec.Server.Domains) != 0 { // If certificates do not specific domains info, bypass
					server := InitDefaultServerConfig()
					err := mergo.MergeWithOverwrite(&server, crd.Spec.Server)
					if err != nil {
						fmt.Println(err.Error())
						fmt.Println("Do something when errors appear") // telegram notification and bypass
					}

					if TestMode {
						Cert_path = "./"
					}

					if _, err := os.Stat(Cert_path + crd.Spec.Server.SSLSecret + ".pem"); os.IsNotExist(err) {
						fmt.Print(err.Error())
						// path/to/whatever does not exist
					} else {
						server.SSLCertificate = Cert_path + crd.Spec.Server.SSLSecret + ".pem"
					}
					// if err := GenerateCerts(Config, CRD_Namespace, crd.Spec.Server.SSLSecret); err != nil {
					// 	fmt.Print(err.Error())
					// } else {
					// 	server.SSLCertificate = "/etc/nginx/certs" + crd.Spec.Server.SSLSecret + ".pem"
					// }

					cfg.Servers = append(cfg.Servers, server)
				}
			} else {
			}
		}
	}

	// fmt.Println(cfg)
	cfgBuffer := cfg.GenerateConfig()

	if !TestMode {
		if err := NginxConfigTest(cfgBuffer); err != nil {
			fmt.Println(err.Error())
			return
		}

		if err := NginxReload(cfgBuffer); err != nil {
			fmt.Println(err.Error())
			return
		}
	}

}
