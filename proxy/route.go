package proxy

import (
	"fmt"
	"grpc-klb/plugins"
	"log"
	"net/http"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

const (
	TIME_DURATION = 10
)

func DefinitionRoute(router *gin.Engine) {
	// set run mode
	gin.SetMode(gin.DebugMode)
	// middleware
	router.Use(gin.Recovery())
	// router.Use(middleware.Tracing())
	router.Use(plugins.UseCookieSession())
	router.Use(plugins.TimeoutHandler(time.Second * TIME_DURATION))
	// no route
	router.NoRoute(NoRouteResponse)
	// route
	getRouter(router)
}

func getRouter(r *gin.Engine) {
	routerViper := viper.New()
	routerViper.SetConfigName("RouteConfig")
	routerViper.SetConfigType("yaml")
	routerViper.AddConfigPath("config/")
	routerViper.WatchConfig()
	routerViper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("router config reload ...", e.Name)
		getRouterTask(r)
	})

}

func getRouterTask(r *gin.Engine) {
	rv := viper.New()
	rv.SetConfigName("RouteConfig")
	rv.SetConfigType("yaml")
	rv.AddConfigPath("config/")

	err := rv.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}
	routeRoot := rv.AllSettings()
	routeMap := routeRoot["route"]
	for _, v := range routeMap.([]interface{}) {
		addRoute(v, r)
	}
}

func addRoute(v interface{}, r *gin.Engine) {
	rmap := v.(map[interface{}]interface{})
	rmapStr := make(map[string]string)
	for k, v := range rmap {
		strKey := fmt.Sprintf("%v", k)
		strValue := fmt.Sprintf("%v", v)
		rmapStr[strKey] = strValue
	}

	if rmap["method"] == "get" {
		r.GET(rmapStr["path"], func(c *gin.Context) {
			Run(c, rmapStr["to"])
		})
	}

	if rmap["method"] == "post" {
		r.POST(rmapStr["path"], func(c *gin.Context) {
			Run(c, rmapStr["to"])
		})
	}
}

// no route
func NoRouteResponse(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"code":  404,
		"error": "oops, page not exists!",
	})
}
