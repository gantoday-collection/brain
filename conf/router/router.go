package router

import (
	"github.com/labstack/echo"
)
func InitRouter() {
	e := echo.New()
	//token验证
	//e.Use(interceptor.TokenInterceptor)
	//这样会将所有访问/static/*的请求去访问views目录。
	e.File("/", "views/index.html")
	e.Static("/static", "views")
	//e.GET("/user",userApi.GET)
	//e.GET("/user/:token",userApi.GET)
	//e.POST("/user/:token",userApi.POST)
	//e.PUT("/user/:token",userApi.PUT)
	//e.GET("/telcode",telcodeApi.GET)
	e.Logger.Fatal(e.Start(":8080"))

}
