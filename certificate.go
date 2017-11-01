package main

import (
	"fmt"
	"io/ioutil"
	// "k8s.io/api/core/v1"
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	// rest "k8s.io/client-go/rest"
	"k8s.io/apimachinery/pkg/watch"
	// "k8s.io/kubernetes/pkg/api"
	// "k8s.io/client-go/pkg/api"
)

var Cert_path = "/etc/nginx/certs/"
var CertInitRunOnce = false

func CertWatch(rclient *dynamic.ResourceClient, c chan string) {
	timeoutSecond := int64(60)
	a, err := rclient.Watch(metav1.ListOptions{Watch: true, TimeoutSeconds: &timeoutSecond, ResourceVersion: "0", LabelSelector: "type=certfile"})
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(a)
		b := a.ResultChan()
		for {
			fmt.Println("new start watch cert")
			msg := <-b
			if msg.Type != "" {
				if msg.Type != "ERROR" {
					fmt.Println(msg)
					CertCreate(rclient, msg)
				}
			} else {
				// Cause of http session is closed, re-new session
				c <- "CertWatch"
				return
			}
		}
	}
}

func CertCreate(rclient *dynamic.ResourceClient, event watch.Event) {
	b := event.Object.GetObjectKind()
	d := b.(*unstructured.Unstructured)
	var secret *KubeSecret
	// var hehe *test_secret
	data, err := d.MarshalJSON()
	if err != nil {
		fmt.Println("Do something when errors appear") // telegram and rollback
		fmt.Println(err.Error())
		return
	} else {
		if err := json.Unmarshal(data, &secret); err != nil {
			fmt.Println("Do something when errors appear") // telegram and rollback
			fmt.Println(err.Error())
			return
		}
	}

	if TestMode {
		Cert_path = "./"
	}

	if err := ioutil.WriteFile(Cert_path+secret.Metadata.Name+".pem", secret.Data["cert"], 0644); err != nil {
		fmt.Println(err.Error())
		fmt.Println("Do something when errors appear")
		return
	}

	// if !TestMode {
	if event.Type == "MODIFIED" {
		if !TestMode {
			NginxConfigTestOnly()
			NgixnReloadOnly()
		}
	} else if event.Type == "ADDED" {
		var resourceClient *dynamic.ResourceClient
		resource := &metav1.APIResource{
			Name:       CRD_Plural, //"nginxcerts",
			Kind:       CRD_Kind,   // "Nginxcert",
			Namespaced: true,
			Verbs:      metav1.Verbs{"get"},
		}
		resourceClient = ApisClient.Resource(resource, CRD_Namespace)
		ConfigCreate(resourceClient)
	}
	// }
}

func CertInit(rclient *dynamic.ResourceClient) {
	fmt.Println("We start CertInit")
	var secretList *KubeSecretList
	a, err := rclient.List(metav1.ListOptions{LabelSelector: "type=certfile"})
	// secretList, err := ClientCMD.Core().Secrets(CRD_Namespace).List(metav1.ListOptions{LabelSelector: "type=certfile"})
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return

	// }
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("====")
		b := a.GetObjectKind()
		d := b.(*unstructured.UnstructuredList)

		data, err := d.MarshalJSON()
		// fmt.Println(data.String())
		// fmt.Println(string(data))
		if err != nil {
			fmt.Println("Do something when errors appear") // telegram and rollback
			fmt.Println(err.Error())
			// panic(err)
		} else {
			if err := json.Unmarshal(data, &secretList); err != nil {
				fmt.Println(err.Error())
				fmt.Println("Do something when errors appear")
			}
		}
	}

	if TestMode {
		Cert_path = "./"
	}

	for _, secret := range secretList.Items {
		fmt.Println(Cert_path + secret.Metadata.Name + ".pem")
		// fmt.Println("-------")
		// fmt.Println(secret)
		// fmt.Println("-------")
		if err := ioutil.WriteFile(Cert_path+secret.Metadata.Name+".pem", secret.Data["cert"], 0644); err != nil {
			fmt.Println(err.Error())
			fmt.Println("Do something when errors appear")
			return
		}
	}

	CertInitRunOnce = true

}

// func (cfg *Nginx_Config) List(rclient *dynamic.ResourceClient) []CRD {
// 	var content CRDList
// 	a, err := rclient.List(metav1.ListOptions{})
// 	if err != nil {
// 		fmt.Println(err.Error())
// 	} else {
// 		fmt.Println("====")
// 		b := a.GetObjectKind()
// 		d := b.(*unstructured.UnstructuredList)
// 		data, err := d.MarshalJSON()
// 		if err != nil {
// 			fmt.Println("Do something when errors appear") // telegram and rollback
// 			fmt.Println(err.Error())
// 			// panic(err)
// 		} else {
// 			if err := json.Unmarshal(data, &content); err != nil {
// 				fmt.Println(err.Error())
// 				fmt.Println("Do something when errors appear")
// 			}
// 		}
// 	}
// 	return content.Items
// }

func GenerateCerts(client *dynamic.Client, nameSpace, secretCertName string) error {
	resource := &metav1.APIResource{
		Name:       "secrets", //"nginxcerts",
		Kind:       "secret",  // "Nginxcert",
		Namespaced: true,
		Verbs:      metav1.Verbs{"get"},
	}

	resourceClient := client.Resource(resource, nameSpace)
	resp, err := resourceClient.Get(secretCertName, metav1.GetOptions{})
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	b := resp.GetObjectKind()
	d := b.(*unstructured.Unstructured)
	var content *KubeSecret
	data, err := d.MarshalJSON()
	if err != nil {
		fmt.Println("Do something when errors appear") // telegram and rollback
		fmt.Println(err.Error())
		// panic(err)
		return err
	} else {
		if err := json.Unmarshal(data, &content); err != nil {
			fmt.Println(err.Error())
			fmt.Println("Do something when errors appear")
			return err
		}
	}
	// fmt.Println(string(content.Data["cert"]))
	if TestMode {
		Cert_path = "./"
	}

	err = ioutil.WriteFile(Cert_path+secretCertName+".pem", content.Data["cert"], 0644)
	// err := ioutil.WriteFile("./"+secretCertName, cfg, 0644)
	return err
}
