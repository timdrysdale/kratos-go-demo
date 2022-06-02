package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	client "github.com/ory/kratos-client-go"
)

type kratosMiddleware struct {
	client *client.APIClient
}

func NewMiddleware() *kratosMiddleware {
	configuration := client.NewConfiguration()
	configuration.Servers = []client.ServerConfiguration{
		{
			URL: "http://127.0.0.1:4433", // Kratos Admin API
		},
	}
	return &kratosMiddleware{
		client: client.NewAPIClient(configuration),
	}
}

func (k *kratosMiddleware) Session() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := k.validateSession(c.Request)
		if err != nil {
			c.Redirect(http.StatusMovedPermanently, "https://core.prac.io/accounts/login")
			c.Abort()
			return
		}

		if !*session.Active { //if session is not active, we need to login again
			c.Redirect(http.StatusMovedPermanently, "https://core.prac.io/accounts/login")
			c.Abort()
			return
		}

		c.Next()
	}
}

func (k *kratosMiddleware) validateSession(r *http.Request) (*client.Session, error) {

	cookie, err := r.Cookie("ory_kratos_session")
	if err != nil {
		return nil, err
	}

	if cookie == nil {
		return nil, errors.New("no session found in cookie")
	}
	resp, _, err := k.client.V0alpha2Api.ToSession(context.Background()).Cookie(cookie.String()).Execute()

	if err != nil {
		return nil, err
	}
	return resp, nil
}

func main() {

	r := gin.Default()
	k := NewMiddleware()

	r.Use(k.Session())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/foo", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "bar",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
