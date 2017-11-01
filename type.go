package main

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigInfo struct {
	CRD_Info struct {
		Namespace string `yaml:"namespace"`
		Kind      string `yaml:"kind"`
		Plural    string `yaml:"plural"`
		Group     string `yaml:"group"`
		Version   string `yaml:"version"`
		Waf_Name  string `yaml:"waf_name"`
	} `yaml:"crd"`
}

type KubeSecretList struct {
	Items []KubeSecret `json:items`
	// Metadata metav1.ObjectMeta `json:metadata`
	// Data     map[string][]byte `json:data`
	// Type     string
}

type KubeSecret struct {
	Metadata metav1.ObjectMeta `json:metadata`
	Data     map[string][]byte `json:data`
	Type     string
}

type Nginx_Config struct {
	Http    Nginx_HTTP_Config
	Servers []Nginx_Server_Config
}

type CRDList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []CRD `json:"items"`
}

type CRD struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Spec            CRDConfig `json:"spec"`
}

type CRDConfig struct {
	WAFName   string              `json:"waf_name"`
	BlockType string              `json:"block_type"`
	Http      Nginx_HTTP_Config   `json:"http"`
	Server    Nginx_Server_Config `json:"server"`
}

type Nginx_HTTP_Config struct {
	Resolver                 string
	ShowServerTokens         bool   `json:"show-server-tokens, omitempty"`
	LogFormat                string `json:"log-format, omitempty"`
	UseHTTP2                 bool   `json:"useHTTP2,omitempty"`
	HSTS                     bool   `json:"hsts, omitempty"`
	HSTSMaxAge               int    `json:"hstsMaxAge,omitempty"`
	BacklogSize              int    `json:"backlogSize, omitempty"`
	KeepAlive                int    `json:"keep-alive, omitempty"`
	KeepAliveRequests        int    `json:"keep-alive-request, omitempty"`
	UseProxyProtocol         bool   `json:"proxy-protocol, omitempty"`
	ProxyRealIPCIDR          string `json:"set-real-ip-from, omitempty"`
	ClientHeaderBufferSize   string `json:"Client-header-buffer-size, omitempty"`
	LargeClientHeaderBuffers string `json:"client-header-buffers, omitempty"`
	ClientBodyBufferSize     string `json:"client-body-buffer-size, omitempty"`
	HTTP2MaxFieldSize        string `json:"http2-max-field-size, omitempty"`
	HTTP2MaxHeaderSize       string `json:"http2-max-header-size, omitempty"`

	ServerNameHashMaxSize    int `json:"server-names-hash-max-size, omitempty"`
	ServerNameHashBucketSize int `json:"server-names-hash-bucket-size, omitempty"`
	MapHashBucketSize        int `json:"map-hash-bucket-size, omitempty"`

	VariablesHashBucketSize    int    `json:"variables-hash-bucket-size, omitempty"`
	VariablesHashMaxSize       int    `json:"variables-hash-max-size, omitempty"`
	EnableUnderscoresInHeaders bool   `json:"enable-under-scores-header, omitempty"`
	IgnoreInvalidHeaders       bool   `json:"variables-hash-max-size, omitempty"`
	GzipTypes                  string `json:"gzip-types, omitempty"`
	DisableAccessLog           bool   `json:"disable-access-log, omitempty"`

	SSLProtocols        string `json:"ssl-protocols, omitempty"`
	SSLSessionCache     bool   `json:"ssl-session-cache, omitempty"`
	SSLSessionCacheSize string `json:"ssl-session-cache-size, omitempty"`
	SSLSessionTimeout   string `json:"ssl-session-timeout, omitempty"`
	SSLSessionTickets   bool   `json:ssl-session-tickets, omitempty"`
	SSLBufferSize       string `json:"ssl-buffer-size, omitempty"`
	SSLCiphers          string `json:"ssl-ciphers, omitempty"`
	SSLECDHCurve        string `json:ssl-ecdh-curve, omitempty"`
}

type Nginx_Server_Config struct {
	SSLSecret           string            `json:"ssl-secret, omitempty"`
	Domains             []string          `json:"domains"`
	Domain_redirect     map[string]string `json:"domain_redirect"`
	SSLCertificate      string
	BodySize            string            `json:"client-max-body-size, omitempty"`
	ProxySetHeaders     map[string]string `json:"proxy-set-headers, omitempty"`
	ProxyConnectTimeout int               `json:"proxy-connect-timeout, omitempty"`
	ProxySendTimeout    int               `json:"proxy-send-timeout, omitempty"`
	ProxyReadTimeout    int               `json:"proxy-read-timeout, omitempty"`
	ProxyBufferSize     string            `json:"proxy-buffer-size, omitempty"`
	EnableCORS          string            `json:"enable-cors, omitempty"`
	Waf                 Waf_Config
}

type Waf_Config struct {
	Mode                string   `json:"waf-mode, omitempty"`
	LogDebug            string   `json:"waf-log-mode, omitempty"`
	LogLevel            string   `json:"waf-log-level, omitempty"`
	EventLogLevel       string   `json:"waf-event-log-level, omitepty"`
	LogFieldsAddition   []string `json:"waf-log-fields-addtion, omitempty"`
	EventLogBufferSize  int      `json:"waf-event-log-buffer-size, omitempty"`
	EventLogRequestBody string   `json:"waf-event-log-request-body, omitempty"`
	ScoreThreshold      int      `json:"waf-score-threshold, omitempty"`
}
