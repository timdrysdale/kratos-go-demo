package main

import (
	"context"
	"errors"
	"fmt"
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
			URL: "http://ory-quickstart-kratos-admin.default.svc.cluster.local", // Kratos Admin API
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
			c.Redirect(http.StatusFound, "https://core.prac.io/accounts/login")
			c.Abort()
			return
		}

		if !*session.Active { //if session is not active, we need to login again
			c.Redirect(http.StatusFound, "https://core.prac.io/accounts/login")
			c.Abort()
			return
		}

		c.Next()
	}
}

func (k *kratosMiddleware) validateSession(r *http.Request) (*client.Session, error) {

	cookie, err := r.Cookie("ory_kratos_session")
	if err != nil {
		fmt.Printf("Retrieve cookie error: %s\n", err)
		return nil, err
	}

	if cookie == nil {
		fmt.Print("No session found in cookie")
		return nil, errors.New("no session found in cookie")
	}
	resp, _, err := k.client.V0alpha2Api.ToSession(context.Background()).Cookie(cookie.String()).Execute()

	if err != nil {
		fmt.Printf("Verfying session error: %s\n", err)
		return nil, err
	}
	fmt.Printf("Found session: %+v\n", (*resp))
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
