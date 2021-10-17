package proxy

import "net/http"

type RpcRequest struct {
	Header    http.Header
	Method    string
	To        string
	Query     interface{}
	TimeOut   int
	CacheTime int
}
