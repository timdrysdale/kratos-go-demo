package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strings"

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
			URL: "http://ory-quickstart-kratos-public.default.svc.cluster.local", // was Kratos Admin API
		},
	}

	// from https://github.com/ory/kratos/blob/cf63a1c14bef86bbb0f0105453677c92cc9c947e/examples/go/pkg/common.go#L28
	cj, _ := cookiejar.New(nil)
	configuration.HTTPClient = &http.Client{Jar: cj}

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

	cookies := make([]string, 0)

	for _, c := range r.Cookies() {
		cookies = append(cookies, c.String())
	}

	joinedCookies := strings.Join(cookies[:], "; ")

	fmt.Printf("Joined Cookies: %s\n", joinedCookies)

	resp, _, err := k.client.V0alpha2Api.ToSession(context.Background()).Cookie(joinedCookies).Execute() //was cookie

	if err != nil {
		fmt.Printf("Verfying session error: %s\n", err)
		return nil, err
	}
	fmt.Printf("Found session: %+v\n", (*resp))
	return resp, nil
}

func main() {

	r := gin.Default()
	//k := NewMiddleware()

	//r.Use(k.Session())
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
	fmt.Println("v0.0.9")
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
