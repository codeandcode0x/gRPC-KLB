package route

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/valyala/fasthttp"
)

type HttpRequest struct {
	Header    http.Header
	Method    string
	To        string
	Query     interface{}
	TimeOut   int
	CacheTime int
}

func Run(c *gin.Context, to string) {
	result := make(chan string)
	resultErr := make(chan error)
	var queryStr interface{}

	queryStr = GetRequestUrl(to, c)
	if strings.ToUpper(c.Request.Method) == "POST" {
		c.Request.ParseForm()
		queryStr = c.Request.PostForm
	}

	httpReq := HttpRequest{
		Header:    c.Request.Header,
		Method:    c.Request.Method,
		To:        to,
		Query:     queryStr,
		TimeOut:   10,
		CacheTime: 10,
	}

	t := time.NewTimer(30 * time.Second)

	go func() {
		bodyStr, err := httpReq.Request()
		if err != nil {
			resultErr <- err
		}
		result <- bodyStr
	}()

	select {
	case res := <-result:
		c.String(http.StatusOK, res)
	case err := <-resultErr:
		c.String(http.StatusInternalServerError, fmt.Sprintln(err))
	case <-t.C:
		c.String(http.StatusNotFound, "request time out")
	}

}

func (httpReq *HttpRequest) Request() (string, error) {
	var body string = ""
	var err error

	method := strings.ToUpper(httpReq.Method)
	switch method {
	case "GET":
		body, err = getRequest(httpReq.Query.(string), httpReq.TimeOut)
	case "POST":
		body, err = postRequest(httpReq.To, httpReq.Query.(url.Values), httpReq.TimeOut, httpReq.Header)
	default:
		err = errors.New("http request any method")
	}

	return body, err
}

// get request uri
func GetRequestUri(c *gin.Context) string {
	c.Request.ParseForm()
	u, _ := url.Parse(c.Request.RequestURI)

	return u.Path
}

// get request url
func GetRequestUrl(to string, c *gin.Context) string {
	query, method := "", c.Request.Method
	switch method {
	case "GET":
		query = c.Request.URL.RawQuery
	case "POST":
		c.Request.ParseForm()
		param := c.Request.PostForm
		if len(param) > 0 {
			query = param.Encode()
		}
	default:
		break
	}

	queryStr := to
	if query != "" {
		queryStr = fmt.Sprintf("%s?%s", to, query)
	}
	return queryStr
}

// get
func getRequest(u string, timeOut int) (string, error) {
	timeout := time.Duration(timeOut) * time.Second

	cli := fasthttp.Client{
		MaxConnsPerHost: 200, //最大链接数
		ReadTimeout:     timeout,
	}

	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()

	req.Header.SetContentType("application/json")
	req.Header.SetMethod("GET")
	req.SetRequestURI(u)

	if err := cli.DoTimeout(req, res, timeout); err != nil {
		return "", err
	}

	body := res.Body()
	bodyStr := string(body)

	return bodyStr, nil
}

// post

func postRequest(to string, param map[string][]string, timeOut int, header http.Header) (string, error) {
	timeout := time.Duration(timeOut) * time.Second

	cli := fasthttp.Client{
		MaxConnsPerHost: 200,     //最大链接数
		ReadTimeout:     timeout, //主动断开时间
	}

	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()

	for k, v := range header {
		for _, value := range v {
			req.Header.Add(k, value)
		}
	}

	req.Header.SetMethod("POST")
	req.SetRequestURI(to)
	args := req.PostArgs()
	for k, v := range param {
		for _, value := range v {
			args.Add(k, value)
		}
	}

	if err := cli.DoTimeout(req, res, timeout); err != nil {
		return "", err
	}

	body := res.Body()
	bodyStr := string(body)

	return bodyStr, nil
}
