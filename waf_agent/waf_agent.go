package main

import(
	"os"
	netContext "golang.org/x/net/context"
	"github.com/coreos/etcd/client"
	"io/ioutil"
	"time"
	"net"
	"net/http"
	"context"
	"strings"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"flag"
	"fmt"
)

type DockerExecRequest struct {
	Id string `json:"Id"`
}

type DockerExecInspect struct {
	ExitCode int `json:"ExitCode"`
}

func BackupConf(nginx_file_path string) error {
	orginFile, err := ioutil.ReadFile(nginx_file_path)
    if err != nil {
    	return err
    }
    
    backup_file := nginx_file_path + ".bak"
	err = ioutil.WriteFile(backup_file, orginFile, 0644)
	return err
}

func RestoreConf(nginx_file_path string) error {
	bakFile, err := ioutil.ReadFile(nginx_file_path + ".bak")
    if err != nil {
    	return err
    }
    
	err = ioutil.WriteFile(nginx_file_path, bakFile, 0644)
	return err
}

func ReloadNginx(execCommand string) error {
	fmt.Println("We start reload Nginx")
	httpc := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", "/var/run/docker.sock")
			},
		},
	}

	post := `
		{
		  "AttachStdin": false,
		  "AttachStdout": true,
		  "AttachStderr": true,
		  "DetachKeys": "ctrl-p,ctrl-q",
		  "Tty": false,
		  "Cmd": [
		  	"sh",
		  	"-c",
		  	` + execCommand + `
		  ]
		}
	`

	fmt.Println(post)
	var response *http.Response
	var err error

	// Register exec command
	response, err = httpc.Post("http://unix/containers/" + os.Getenv("remote_exec_container") + "/exec", "application/json", strings.NewReader(post))

	if err != nil {
		fmt.Println(err.Error())
		log.WithFields(log.Fields{
		  "message": err.Error(),
		  "type": "request",
		  "statusCode": 999,
		}).Error("execRequest")
		return err
	}

	if response.StatusCode != 201 {
		log.WithFields(log.Fields{
		  "message": response.Status,
		  "type": "request",
		  "statusCode": response.Status,
		}).Error("execRequest")
		return err
	}

	body, _ := ioutil.ReadAll(response.Body)

	var execRequest DockerExecRequest
	err = json.Unmarshal(body, &execRequest)
	if err != nil {
		log.Println(err.Error())
		log.WithFields(log.Fields{
		  "message": err.Error(),
		  "type": "jsonDecode",
		}).Error("execRequest")
		return err
	}


	post = `
		{
		  "DetachKeys": false,
		  "Tty": false
		}
	`


	// Run exec command
	response, err = httpc.Post("http://unix/exec/" + execRequest.Id + "/start", "application/json", strings.NewReader(post))
	if err != nil {
		fmt.Println(err.Error())
		log.WithFields(log.Fields{
		  "message": err.Error(),
		  "type": "request",
		  "statusCode": 999,
		}).Error("execRun")
		return err
	}


	execResponse, _ := ioutil.ReadAll(response.Body)

	if response.StatusCode != 200 {
		fmt.Println(response.StatusCode)
		log.WithFields(log.Fields{
		  "message": response.Status,
		  "type": "request",
		  "statusCode": response.StatusCode,
		}).Error("execRun")
		return err
	}



	// Get exec command status
	response, err = httpc.Get("http://unix/exec/" + execRequest.Id + "/json")
	if err != nil {
		fmt.Println(err.Error())
		log.WithFields(log.Fields{
		  "message": err.Error(),
		  "type": "request",
		  "statusCode": 999,
		}).Error("execStatus")
		return err
	}


	if response.StatusCode != 200 {
		fmt.Println(response.StatusCode)
		log.WithFields(log.Fields{
		  "message": response.Status,
		  "type": "request",
		  "statusCode": response.StatusCode,
		}).Error("execStatus")
		return err
	}


	var execStatus DockerExecInspect
	body, _ = ioutil.ReadAll(response.Body)
	_ = json.Unmarshal(body, &execStatus)

	if execStatus.ExitCode != 0 {
		fmt.Println(execStatus.ExitCode)
		log.WithFields(log.Fields{
		  "message": string(execResponse),
		  "type": "request",
		  "statusCode": execStatus.ExitCode,
		}).Error("execStatus")		
		return err
	}

	return nil
}

func Conf(kAPI client.KeysAPI, etcdkey, nginx_file_path, execCommand string) {
	resp, err := kAPI.Get(netContext.Background(), etcdkey, &client.GetOptions{})

	if err != nil {
		log.WithFields(log.Fields{
		  "message": err.Error(),
		  "type": "etcd",
		  "statusCode": 999,
		}).Error("etcdGet")		
		return
	}

	
	if resp.Node.Value == "" {
		log.WithFields(log.Fields{
		  "message": etcdkey + " key is empty",
		  "type": "etcd",
		  "statusCode": 999,
		}).Error("etcdGet")		
		return
	}


	if err := BackupConf(nginx_file_path); err != nil {
		log.WithFields(log.Fields{
		  "message": err.Error(),
		  "type": "operation",
		  "statusCode": 1000,
		}).Error("BackupConf")	
		return
	}

    err = ioutil.WriteFile(nginx_file_path, []byte(resp.Node.Value), 0644)
    if err != nil {
		log.WithFields(log.Fields{
		  "message": err.Error(),
		  "type": "operation",
		  "statusCode": 1000,
		}).Error("writeFile")
		if err := RestoreConf(nginx_file_path); err != nil {
			log.WithFields(log.Fields{
			  "message": err.Error(),
			  "type": "operation",
			  "statusCode": 1000,
			}).Error("RestoreConf")
		}
		return		    	
    }

    if err := ReloadNginx(execCommand); err != nil {
    	if err := RestoreConf(nginx_file_path); err != nil {
			log.WithFields(log.Fields{
			  "message": err.Error(),
			  "type": "operation",
			  "statusCode": 1000,
			}).Error("RestoreConf")	
    	}
    }	
}

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	f, err := os.OpenFile("/var/log/waf_agent.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		panic(err)
	}
	log.SetOutput(f)
}

func main() {
	var execCommand string
    flag.StringVar(&execCommand, "command", "", "command run at remote_container")
    flag.Parse()

	var etcd_endpoint = "http://127.0.0.1:2379"
	if os.Getenv("etcd_endpoint") != "" {
		etcd_endpoint = os.Getenv("etcd_endpoint")
	}
	cli, err := client.New(client.Config{
		Endpoints:   []string{etcd_endpoint},
		Transport:   client.DefaultTransport,
		Username: 	 os.Getenv("etcd_user_name"),
		Password: 	 os.Getenv("etcd_user_password"),
		HeaderTimeoutPerRequest: 50 * time.Second, // Incase Etcd behind proxy, this value must be lesser than proxy timeout
	})

	if err != nil {
		panic(err)
	}
	kAPI := client.NewKeysAPI(cli)

	key := os.Getenv("nginx_file_key")
	nginx_file_path := os.Getenv("nginx_file_path")

	if os.Getenv("nginx_file_key") != "" || os.Getenv("nginx_file_path") != "" || os.Getenv("remote_exec_container") != "" {
		Conf(kAPI, key, nginx_file_path, execCommand)

	 	w := kAPI.Watcher(key, &client.WatcherOptions{AfterIndex: uint64(0)})
		for {
			resp, err := w.Next(netContext.TODO())
			if err != nil {
				log.Println(err.Error())
				log.WithFields(log.Fields{
				  "message": key + " key is missing",
				  "type": "etcd",
				  "statusCode": 999,
				}).Error("etcdWatch")
				time.Sleep(time.Duration(2) * time.Second)
				continue
			}
			Conf(kAPI, resp.Node.Key, nginx_file_path, execCommand)
		}		
	} else {
		log.WithFields(log.Fields{
		  "message": "nginx_file_key or nginx_file_path or remote_exec_container are missing",
		  "type": "operation",
		  "statusCode": 1000,
		}).Fatal("init")	
	}
}