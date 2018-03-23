package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func MiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Server", "GMFE 0.9.1")
	}
}

func worker() http.Handler {
	router := gin.New()
	router.Use(MiddleWare())
	router.LoadHTMLGlob("templates/*")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	router.StaticFS("/src", http.Dir("src"))
	router.StaticFS("/images", http.Dir("images"))

	router.GET("/fe", func(c *gin.Context) {
		c.HTML(http.StatusOK, "fe.html", gin.H{})
	})
	router.GET("/tricks", func(c *gin.Context) {
		c.HTML(http.StatusOK, "tricks.html", gin.H{})
	})
	router.GET("/pure", func(c *gin.Context) {
		c.HTML(http.StatusOK, "pure.html", gin.H{})
	})
	return router
}
