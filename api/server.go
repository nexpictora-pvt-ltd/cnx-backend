package api

import (
	"fmt"
	"io"
	"mime/multipart"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
	db "github.com/nexpictora-pvt-ltd/cnx-backend/db/sqlc"
	"github.com/nexpictora-pvt-ltd/cnx-backend/token"
	"github.com/nexpictora-pvt-ltd/cnx-backend/util"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	config     util.Config
	store      *db.Store
	tokenMaker token.Maker
	router     *gin.Engine
	s3Uploader *s3manager.Uploader
}

func NewServer(config util.Config, store *db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)

	//By uncommenting this line and commenting above line we can use JWT token as access token as both of them use same maker interfaces
	// tokenMaker, err := token.NewJWTMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	awsSession, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String("ap-south-1"), // Replace with your AWS S3 region
			Credentials: credentials.NewStaticCredentials(
				config.AWSAccessKey, // Replace with your AWS access key
				config.AWSSecretKey, // Replace with your AWS secret key
				"",                  // Optional: Replace with your AWS session token if using temporary credentials
			),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create AWS session: %w", err)
	}

	s3Uploader := s3manager.NewUploader(awsSession)

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
		s3Uploader: s3Uploader,
	}

	server.setupRouter()
	return server, nil
}

func (server *Server) uploadToS3(fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Upload the file to S3
	uploadedURL, err := saveFile(file, fileHeader, server.s3Uploader)
	if err != nil {
		return "", err
	}

	return uploadedURL, nil
}

func saveFile(fileReader io.Reader, fileHeader *multipart.FileHeader, uploader *s3manager.Uploader) (string, error) {
	S3Bucket := aws.String("ctt-test-001")
	// Upload the file to S3 using the provided uploader
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: S3Bucket, // Replace with your S3 bucket name
		Key:    aws.String(fileHeader.Filename),
		Body:   fileReader,
	})
	if err != nil {
		return "", err
	}

	// Get the URL of the uploaded file
	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", *S3Bucket, fileHeader.Filename) // Replace with your S3 bucket name

	return url, nil
}

func (server *Server) setupRouter() {

	router := gin.Default()
	router.Use(corsMiddleware())
	// // Apply CORS middleware
	// config := cors.DefaultConfig()
	// config.AllowAllOrigins = true
	// config.AllowCredentials = true
	// config.AddAllowHeaders("Authorization")

	// router.Use(corsMiddleware())
	simpleRoutes := router.Group("/").Use(corsMiddleware())
	// Swagger Documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// Anyone can use or create user or login routes
	router.POST("/users", server.createUser)
	simpleRoutes.POST("/users/login", server.loginUser)

	// Company can create admin or Admin can add new admins

	router.POST("/create-new-admin", server.createAdmin)
	router.POST("/admin/login", server.loginAdmin)

	//Here the endpoint is used to renew access token for user session
	router.POST("/tokens/renew_access", server.renewAccessToken)
	// Here we are grouping all the routes and making them protected
	userAuthRoutes := router.Group("/").Use(userAuthMiddleWare(server.tokenMaker))
	adminAuthRoutes := router.Group("/").Use(adminAuthMiddleWare(server.tokenMaker))
	// These Routes should be protected as not everyone should have access to it
	userAuthRoutes.GET("/users/:user_id", server.getUser)
	router.GET("/users", server.listUser)

	userAuthRoutes.GET("/users/all", server.listAllUsers)

	// Admin can add new admins
	adminAuthRoutes.POST("/admin/new", server.addAdmin)

	// Admin can Create/Add service
	adminAuthRoutes.POST("/services", server.createService)

	// User Endpoints
	userAuthRoutes.GET("/services/:service_id", server.getService)
	userAuthRoutes.GET("/services/user/all", server.listServices)
	// adminAuthRoutes.GET("/services/admin/all", server.listServices)

	//Admin Endpoints
	userAuthRoutes.GET("/services/preview", server.listServices)
	adminAuthRoutes.GET("/services", server.listLimitedServices)
	adminAuthRoutes.PUT("/services/:service_id", server.updateService)
	adminAuthRoutes.DELETE("/services/:service_id", server.deleteService)

	userAuthRoutes.POST("/orders", server.createOrder)
	adminAuthRoutes.PUT("/orders/status", server.updateOrderStatus)
	adminAuthRoutes.PUT("/orders/delivery", server.updateOrderDelivered)
	adminAuthRoutes.GET("/orders/:order_id", server.getOrder)
	userAuthRoutes.GET("/orders/users/:order_id", server.getOrder)
	adminAuthRoutes.GET("/orders", server.listOrders)
	adminAuthRoutes.GET("/orders/all", server.listAllOrders)

	server.router = router
}

// Start the http server on the input/specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
