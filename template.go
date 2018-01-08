package main

import (
	"bytes"
	"fmt"
	"github.com/miekg/dns"
	// "html/template"
	"io/ioutil"
	"text/template"
	// "reflect"
	"strings"
)

// const (
var nginx_template = "/etc/nginx/template/nginx.tmpl"

// nginx_template = "./nginx.tmpl"
// )

func InitDefaultHTTPConfig() Nginx_HTTP_Config {
	var nginx_http Nginx_HTTP_Config
	nginx_http = Nginx_HTTP_Config{
		Resolver:                 "127.0.0.1",
		LogFormat:                `'$the_x_forwarded_for - [$the_real_ip] - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent" $request_length $request_time [$proxy_upstream_name] $upstream_addr $upstream_response_length $upstream_response_time $upstream_status';`,
		ShowServerTokens:         false,
		UseHTTP2:                 true,
		BacklogSize:              16800,
		KeepAlive:                60,
		KeepAliveRequests:        100,
		ProxyRealIPCIDR:          "10.0.0.0/8",
		ClientHeaderBufferSize:   "1k",
		LargeClientHeaderBuffers: "4 8k",
		ClientBodyBufferSize:     "8k",
		UseProxyProtocol:         false,
		HTTP2MaxFieldSize:        "4k",
		HTTP2MaxHeaderSize:       "16k",

		ServerNameHashMaxSize:    2048,
		ServerNameHashBucketSize: 64,
		MapHashBucketSize:        64,

		VariablesHashBucketSize:    64,
		VariablesHashMaxSize:       2048,
		EnableUnderscoresInHeaders: true,
		IgnoreInvalidHeaders:       false,
		GzipTypes: `font/opentype
					image/svg+xml
					image/x-icon
					text/x-component
					text/plain
					text/html
					text/xml
					text/css
					application/xml
					application/xhtml+xml
					application/rss+xml
					application/atom_xml
					application/javascript
					application/x-javascript
					application/x-httpd-php
					application/vnd.ms-fontobject
					application/x-font-ttf
					application/x-web-app-manifest+json`,
		DisableAccessLog:    true,
		SSLProtocols:        "TLSv1 TLSv1.1 TLSv1.2",
		SSLSessionCache:     true,
		SSLSessionCacheSize: "10m",
		SSLSessionTimeout:   "10m",
		SSLSessionTickets:   true,
		SSLBufferSize:       "16k",
		SSLCiphers:          `ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:DHE-DSS-AES128-GCM-SHA256:kEDH+AESGCM:ECDHE-RSA-AES128-SHA256:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA:ECDHE-ECDSA-AES128-SHA:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA:ECDHE-ECDSA-AES256-SHA:DHE-RSA-AES128-SHA256:DHE-RSA-AES128-SHA:DHE-DSS-AES128-SHA256:DHE-RSA-AES256-SHA256:DHE-DSS-AES256-SHA:DHE-RSA-AES256-SHA:AES128-GCM-SHA256:AES256-GCM-SHA384:AES128-SHA256:AES256-SHA256:AES128-SHA:AES256-SHA:AES:CAMELLIA:DES-CBC3-SHA:!aNULL:!eNULL:!EXPORT:!DES:!RC4:!MD5:!PSK:!aECDH:!EDH-DSS-DES-CBC3-SHA:!EDH-RSA-DES-CBC3-SHA:!KRB5-DES-CBC3-SHA`,
		SSLECDHCurve:        "secp384r1",
	}

	return nginx_http
}

func InitDefaultServerConfig() Nginx_Server_Config {
	var nginx_server Nginx_Server_Config
	nginx_server = Nginx_Server_Config{
		Domains:             []string{"__"},
		SSLCertificate:      "/etc/nginx/certs/default.pem",
		BodySize:            "30m",
		ProxySetHeaders:     make(map[string]string),
		ProxyConnectTimeout: 75,
		ProxySendTimeout:    300,
		ProxyReadTimeout:    300,
		ProxyBufferSize:     "4k",
		Waf: Waf_Config{
			Mode:                "SIMULATE",
			LogDebug:            "true",
			LogLevel:            "ngx.DEBUG",
			EventLogLevel:       "ngx.DEBUG",
			LogFieldsAddition:   []string{"host", "request_id"},
			EventLogBufferSize:  8192,
			EventLogRequestBody: "false",
			ScoreThreshold:      10,
		},
	}
	return nginx_server
}

var (
	funcMap = template.FuncMap{
		"empty": func(input interface{}) bool {
			check, ok := input.(string)
			if ok {
				return len(check) == 0
			}
			return true
		},
		"emptyMap": func(input interface{}) bool {
			check, ok := input.(map[string]string)
			if ok {
				return len(check) == 0
			}
			return true
		},
		"buildResolvers": buildResolvers,
		"mirrorModuleCheck": NginxMirrorModuleCheck,
	}
)

// buildResolvers returns the resolvers reading the /etc/resolv.conf file
func buildResolvers() string {
	r := []string{"resolver"}
	config, _ := dns.ClientConfigFromFile("/etc/resolv.conf")
	for _, ns := range config.Servers {
		r = append(r, ns)
	}
	r = append(r, "valid=30s;")
	return strings.Join(r, " ")
}

func (cfg Nginx_Config) GenerateConfig() []byte {
	var configContent []byte

	if TestMode {
		nginx_template = "./nginx.tmpl"
	}

	dat, err := ioutil.ReadFile(nginx_template)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		var templateBody bytes.Buffer

		general := template.New("File template")
		general.Funcs(funcMap)
		general, err = general.Parse(string(dat))

		if err != nil {
			fmt.Println("Wrong template format is : " + err.Error())
		} else {
			// buffer := new(bytes.Buffer)
			general.Execute(&templateBody, cfg)
			configContent = templateBody.Bytes()
			// configContent := templateBody
			fmt.Println(templateBody.String())
			// fmt.Println(reflect.TypeOf(configContent))

		}
	}
	return configContent
}
