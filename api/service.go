package api

import (
	"database/sql"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	db "github.com/nexpictora-pvt-ltd/cnx-backend/db/sqlc"
)

// func (server *Server) uploadToS3(fileHeader *multipart.FileHeader) (string, error) {
// 	file, err := fileHeader.Open()
// 	if err != nil {
// 		return "", err
// 	}
// 	defer file.Close()

// 	// Upload the file to S3
// 	uploadedURL, err := saveFile(file, fileHeader)
// 	if err != nil {
// 		return "", err
// 	}

// 	return uploadedURL, nil
// }

type createServiceRequest struct {
	ServiceName  string                `json:"service_name" binding:"required"`
	ServicePrice int64                 `json:"service_price" binding:"required"`
	ServiceImage *multipart.FileHeader `form:"service_image" binding:"required"`
}

func (server *Server) createService(ctx *gin.Context) {
	// Parse the multipart form data
	err := ctx.Request.ParseMultipartForm(10 << 20) // 10 MB is the maximum size of the uploaded file
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Access other form fields
	serviceName := ctx.Request.FormValue("service_name")
	servicePriceStr := ctx.Request.FormValue("service_price")

	// Convert servicePrice to int64
	servicePrice, err := strconv.ParseInt(servicePriceStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Access the file
	file, fileHeader, err := ctx.Request.FormFile("service_image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	defer file.Close()

	// Upload the image to S3
	imageURL, err := server.uploadToS3(fileHeader)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Create service with imageURL
	arg := db.CreateServiceParams{
		ServiceName:  serviceName,
		ServicePrice: servicePrice,
		ServiceImage: imageURL,
	}

	service, err := server.store.CreateService(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, service)
}

// func (server *Server) uploadToS3(fileHeader *multipart.FileHeader) (string, error) {
// 	file, err := fileHeader.Open()
// 	if err != nil {
// 		return "", err
// 	}
// 	defer file.Close()

// 	// Upload the file to S3
// 	uploadedURL, err := saveFile(file, fileHeader, server.s3Uploader)
// 	if err != nil {
// 		return "", err
// 	}

// 	return uploadedURL, nil
// }

type getServiceRequest struct {
	ServiceID int64 `uri:"service_id" binding:"required,min=1"`
}

func (server *Server) getService(ctx *gin.Context) {
	var req getServiceRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	service, err := server.store.GetService(ctx, int32(req.ServiceID))
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, service)
}

type listServicesRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listLimitedServices(ctx *gin.Context) {
	var req listServicesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListLimitedServicesParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}
	services, err := server.store.ListLimitedServices(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, services)
}

type updateServiceRequest struct {
	ServiceID    int64  `uri:"service_id" binding:"required,min=1"`
	ServiceName  string `json:"service_name" binding:"required"`
	ServicePrice int64  `json:"service_price" binding:"required"`
	ServiceImage string `json:"service_image" binding:"required"`
}

func (server *Server) updateService(ctx *gin.Context) {
	var req updateServiceRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	updateParams := db.UpdateServiceParams{
		ServiceID:    int32(req.ServiceID),
		ServiceName:  req.ServiceName,
		ServicePrice: req.ServicePrice,
		ServiceImage: req.ServiceImage,
	}

	service, err := server.store.UpdateService(ctx, updateParams)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, service)
}

type deleteServiceRequest struct {
	ServiceID int64 `uri:"service_id" binding:"required,min=1"`
}

func (server *Server) deleteService(ctx *gin.Context) {
	var req deleteServiceRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	err := server.store.DeleteService(ctx, int32(req.ServiceID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, "Service Deleted Successfully")
}

func (server *Server) listServices(ctx *gin.Context) {
	service, err := server.store.ListAllServices(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, service)
}
