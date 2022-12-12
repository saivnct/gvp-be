package apphttp

import (
	"encoding/hex"
	"errors"
	"fmt"
	"gbb.go/gvp/appws"
	"gbb.go/gvp/dao"
	"gbb.go/gvp/model"
	"gbb.go/gvp/proto/grpcXVPPb"
	"gbb.go/gvp/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	"log"
	"net/http"
	"strconv"
)

var (
	HTTPAccessDeniedErr         = errors.New("Access Denied")
	HTTPUnauthenticatedTokenErr = errors.New("Invalid Token")
	HTTPUnauthenticateUserdErr  = errors.New("Invalid User")
	ALLOW_ORIGINS               = []string{"https://dzunu.com"}
)

const (
	API_CRED_AUTH = "authorization"
	CTX_KEY_USER  = "authorization"
)

func CORSMiddleware(allowOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", fmt.Sprintf("Accept, Authorization, Content-Type, Content-Length, X-CSRF-Token, Token, session, Origin, Host, Connection, Accept-Encoding, Accept-Language, X-Requested-With, %v", FILE_ID))

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		//c.Request.Header.Del("Origin")

		c.Next()
	}
}

func WSMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		jwtToken := ctx.Query(appws.WS_CRED_AUTH)
		//log.Println("WS - WSMiddleware", jwtToken)

		payload, err := hex.DecodeString(jwtToken)
		if err != nil {
			log.Println("WS - Invalid decode jwtToken", err)
			ctx.AbortWithError(http.StatusUnauthorized, HTTPUnauthenticatedTokenErr)
			return
		}

		claims, err := utils.ParseJWTToken(string(payload))
		if err != nil {
			log.Println("WS - Invalid jwtToken", err)
			ctx.AbortWithError(http.StatusUnauthorized, HTTPUnauthenticatedTokenErr)
			return
		}

		userName := claims.Subject

		_, err = dao.GetUserDAO().FindByUserName(ctx, userName)
		if err != nil {
			log.Println("WSMiddleware - not found user", userName)
			ctx.AbortWithError(http.StatusForbidden, HTTPAccessDeniedErr)
			return
		}

		//fmt.Println("on WS connect, path:", ctx.FullPath())
		//fmt.Println("phoneFull:", phoneFull)
		//ctx.Set(appws.WS_CTX_KEY_USER, user)

		ctx.Next()
	}
}

func APIMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//origin := ctx.Request.Header.Get("Origin")
		//log.Println("GetKey - origin:", origin)
		//log.Println("GetKey - clientIP:", ctx.ClientIP())

		//idx := slices.IndexFunc(ALLOW_ORIGINS, func(t string) bool { return t == origin })
		//if idx < 0 {
		//	if !strings.HasPrefix(origin, "http://localhost:") && !strings.HasPrefix(origin, "http://10.61.60.41:") {
		//		log.Println("APIMiddleware - Invalid origin", origin)
		//		ctx.AbortWithError(http.StatusForbidden, HTTPAccessDeniedErr)
		//		return
		//	}
		//}

		user := &model.User{
			Username:  ctx.ClientIP(),
			Email:     "",
			Role:      grpcXVPPb.USER_ROLE_GUEST,
			Password:  "",
			CreatedAt: utils.UTCNowMilli(),
		}

		jwtToken := ctx.GetHeader(API_CRED_AUTH)
		//log.Println("APIMiddleware", jwtToken)

		if len(jwtToken) > 0 {
			claims, err := utils.ParseJWTToken(jwtToken)

			if err != nil {
				log.Println("APIMiddleware - Invalid jwtToken", err)
				ctx.AbortWithError(http.StatusUnauthorized, HTTPUnauthenticatedTokenErr)
				return
			}

			jwtUsername := claims.Subject

			user, err = dao.GetUserDAO().FindByUserName(ctx, jwtUsername)
			if err != nil {
				log.Println("APIMiddleware - not found user", jwtUsername)
				ctx.AbortWithError(http.StatusForbidden, HTTPAccessDeniedErr)
				return
			}
		}

		ctx.Set(CTX_KEY_USER, user)

		ctx.Next()
	}
}

func StartServer(httpPort int, httpSessionSecret string, ws *socketio.Server) (*gin.Engine, error) {
	newsController := NewsController{}
	//fileHTTPController := FileHTTPController{}

	gin.SetMode(gin.ReleaseMode)

	// Creates a gin router with default middleware:
	// logger and recovery (crash-free) middleware
	router := gin.Default()
	//	https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies
	router.SetTrustedProxies([]string{"localhost"})

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		//AllowOrigins:     []string{"http://localhost:63343", "http://localhost:4200"},
		AllowMethods:     []string{"POST", "OPTIONS", "GET", "PUT", "DELETE"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "Content-Length", "X-CSRF-Token", "Token", "session", "Origin", "Host", "Connection", "Accept-Encoding", "Accept-Language", "X-Requested-With", FILE_ID},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		//AllowOriginFunc: func(origin string) bool {
		//	return origin == "https://github.com"
		//},
		//MaxAge: 12 * time.Hour,
	}))

	//NOTE:
	//to save cookie from server with different domain -> client set xmlHttp.withCredentials = true -> on server side AllowOrigins must not be "*"
	//https://www.techstack4u.com/angular/set-cookie-in-response-header-not-stored-in-the-browser-not-working/
	store := cookie.NewStore([]byte(httpSessionSecret))
	store.Options(sessions.Options{MaxAge: 60 * 60 * 24, HttpOnly: true, Path: "/"})
	router.Use(sessions.Sessions("xvpsession", store))

	//router.Use(CORSMiddleware("*"))
	//router.Use(CORSMiddleware("http://localhost:63343"))

	wsRouter := router.Group("ws")
	wsRouter.Use(WSMiddleware())
	{
		wsRouter.GET("/*any", gin.WrapH(ws))
		wsRouter.POST("/*any", gin.WrapH(ws))
	}

	mediaRouter := router.Group("news")
	mediaRouter.Use(APIMiddleware())
	{
		//mediaRouter.GET("/key", newsController.GetKey)
		mediaRouter.GET("/key/v2/:newsId", newsController.GetKeyV2)
		mediaRouter.GET("/mediaPresignedUrl/:fileId", newsController.GetPresignedUrl)

		//mediaRouter.POST("/uploadSinglePreviewImage", newsController.UploadSinglePreviewImage)
		//mediaRouter.POST("/uploadMultiplePreviewImages", newsController.UploadMultiplePreviewImages)
		//
		//mediaRouter.POST("/uploadSingleMedia", newsController.UploadSingleMedia)
		//mediaRouter.POST("/uploadMultipleMedias", newsController.UploadMultipleMedias)
	}

	router.GET("/healthCheck", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	//download media msg
	//fileRouter := router.Group("/media")
	//{
	//	fileRouter.GET("/get/:msgId/:fileId", fileHTTPController.getFile)
	//}

	//testRouter := router.Group("/test")
	//{
	//	testRouter.GET("/sendmsg/:name", testHTTPController.SendMsg)
	//	testRouter.GET("/test", testHTTPController.Test)
	//}

	go func() {
		go func() {
			err := ws.Serve()
			if err != nil {
				log.Fatalf("Failed to serve ws: %v", err)
			}
		}()
		defer ws.Close()

		fmt.Printf("Starting http server on: %v\n", httpPort)
		err := router.Run(":" + strconv.Itoa(httpPort))
		if err != nil {
			log.Fatalf("Failed to serve http: %v", err)
		}
	}()

	return router, nil
}
