package route

import (
	"fmt"
	"net/http"
	"pod-api-gin/middleware"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	TIME_DURATION = 10
)

func DefinitionRoute(router *gin.Engine) {
	// set run mode
	gin.SetMode(gin.DebugMode)
	// middleware
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(middleware.Tracing())
	router.Use(middleware.UseCookieSession())
	router.Use(middleware.TimeoutHandler(time.Second * TIME_DURATION))
	// no route
	router.NoRoute(NoRouteResponse)

	//web
	// router.Static("/web/assets", "./web/assets")
	// router.StaticFS("/web/upload", http.Dir("/web/upload"))
	// router.LoadHTMLGlob("web/*.tmpl")

	// // gateway auth
	// auth := router.Group("/")
	// auth.Use(middleware.AuthMiddle())
	// {}

	// route
	getRouter(router)

	// api doc
	router.GET("/swagger/*any", ginSwagger.DisablingWrapHandler(swaggerFiles.Handler, "USE_SWAGGER"))

}

func getRouter(r *gin.Engine) {
	rv := viper.New()
	rv.SetConfigName("route")
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
