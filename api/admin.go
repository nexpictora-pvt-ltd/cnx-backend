package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/nexpictora-pvt-ltd/cnx-backend/db/sqlc"
	"github.com/nexpictora-pvt-ltd/cnx-backend/util"
)

type createAdminRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required"`
	Address  string `json:"address" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type adminResponse struct {
	AdminID           int32     `json:"admin_id"`
	Name              string    `json:"name"`
	Email             string    `json:"email"`
	Phone             string    `json:"phone"`
	Address           string    `json:"address"`
	TotalOrders       int32     `json:"total_orders"`
	CreatedAt         time.Time `json:"created_at"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
}

func newAdminResponse(admin db.Admin) adminResponse {
	return adminResponse{
		AdminID:           admin.AdminID,
		Name:              admin.Name,
		Email:             admin.Email,
		Phone:             admin.Phone,
		Address:           admin.Address,
		PasswordChangedAt: admin.PasswordChangedAt,
		CreatedAt:         admin.CreatedAt,
	}
}

func (server *Server) createAdmin(ctx *gin.Context) {
	var req createAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	arg := db.CreateAdminParams{
		Name:           req.Name,
		Email:          req.Email,
		Phone:          req.Phone,
		Address:        req.Address,
		HashedPassword: hashedPassword,
	}

	admin, err := server.store.CreateAdmin(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	res := newAdminResponse(admin)
	ctx.JSON(http.StatusOK, res)
}

func (server *Server) addAdmin(ctx *gin.Context) {
	var req createAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	arg := db.AddAdminParams{
		Name:           req.Name,
		Email:          req.Email,
		Phone:          req.Phone,
		Address:        req.Address,
		HashedPassword: hashedPassword,
	}

	admin, err := server.store.AddAdmin(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	res := newAdminResponse(admin)
	ctx.JSON(http.StatusOK, res)
}

type listAdminsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listAdmins(ctx *gin.Context) {
	var req listAdminsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListAdminsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}
	admins, err := server.store.ListAdmins(ctx, arg)
	if err != nil {

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, admins)

}

type loginAdminRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type loginAdminResponse struct {
	SessionID             uuid.UUID     `json:"session_id"`
	AccessToken           string        `json:"access_token"`
	AccessTokenExpiresAt  time.Time     `json:"access_token_expires_at"`
	RefreshToken          string        `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time     `json:"refresh_token_expires_at"`
	Admin                 adminResponse `json:"admin"`
}

func (server *Server) loginAdmin(ctx *gin.Context) {
	var req loginAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	admin, err := server.store.GetAdminByEmail(ctx, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = util.CheckPassword(req.Password, admin.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		admin.Email,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
		admin.Email,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Email:        admin.Email,
		RefreshToken: refreshToken,
		UserAgent:    ctx.Request.UserAgent(),
		CleintIp:     ctx.ClientIP(),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	res := loginAdminResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		Admin:                 newAdminResponse(admin),
	}

	ctx.JSON(http.StatusOK, res)
}
