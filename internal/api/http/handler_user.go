package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tfenng/scaffold/internal/repo"
	"github.com/tfenng/scaffold/internal/service"
)

type UserHandler struct{ Svc *service.UserService }

func (h *UserHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	u, err := h.Svc.GetByID(c.Request.Context(), id)
	if err != nil { c.Error(err); return }
	c.JSON(http.StatusOK, u)
}

type createUserReq struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name" binding:"required"`
}

func (h *UserHandler) Create(c *gin.Context) {
	var req createUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(/*domain.Invalid*/ err) // 也可以把 bind 错误映射成 domain.Invalid
		return
	}
	u, err := h.Svc.Create(c.Request.Context(), req.Email, req.Name)
	if err != nil { c.Error(err); return }
	c.JSON(http.StatusCreated, u)
}

type listUsersQuery struct {
	Email    *string `form:"email"`
	NameLike *string `form:"name_like"`
	Page     int32   `form:"page"`
	PageSize int32   `form:"page_size"`
}

func (h *UserHandler) List(c *gin.Context) {
	var q listUsersQuery
	if err := c.ShouldBindQuery(&q); err != nil { c.Error(err); return }

	out, err := h.Svc.List(c.Request.Context(), repo.UserListFilter{
		Email: q.Email, NameLike: q.NameLike, Page: q.Page, PageSize: q.PageSize,
	})
	if err != nil { c.Error(err); return }
	c.JSON(http.StatusOK, out)
}
