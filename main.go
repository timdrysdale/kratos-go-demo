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
			fmt.Printf("Session error: %s\n", err)
			c.Abort()
			return
		}
		fmt.Printf("Should not get to here if there is a missing cookie")
		if !*session.Active { //if session is not active?
			c.Redirect(http.StatusMovedPermanently, "https://core.prac.io/book")
			fmt.Println("Session not active")
			c.Abort()
			return
		}
		fmt.Println("Going to c.Next()")
		c.Next()
	}
}
func (k *kratosMiddleware) validateSession(r *http.Request) (*client.Session, error) {
	cookie, err := r.Cookie("ory_kratos_session")
	if err != nil {
		fmt.Printf("Error getting cookie: %s\n", err)
		return nil, err
	}
	if cookie == nil {
		fmt.Printf("no session found in cookie")
		return nil, errors.New("no session found in cookie")
	}
	resp, _, err := k.client.V0alpha2Api.ToSession(context.Background()).Cookie(cookie.String()).Execute()
	if err != nil {
		fmt.Printf("Error k.client.V0alpha2Api.ToSession: %s", err)
		return nil, err
	}
	fmt.Printf("Resp: %s\n", resp)
	return resp, nil
}
func main() {

	r := gin.Default()
	k := NewMiddleware()

	r.Use(k.Session())
	r.GET("/ping", func(c *gin.Context) {
		fmt.Println("In context, doing pong")
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
