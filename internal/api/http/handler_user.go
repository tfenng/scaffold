package http

import (
	"net/http"
	"strconv"
	"time"

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
	Email    string  `json:"email" binding:"required,email"`
	Name     string  `json:"name" binding:"required"`
	UsedName *string `json:"used_name"`
	Company  *string `json:"company"`
	Birth    *string `json:"birth"`
}

func (h *UserHandler) Create(c *gin.Context) {
	var req createUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(/*domain.Invalid*/ err)
		return
	}

	var birth *time.Time
	if req.Birth != nil {
		t, err := time.Parse("2006-01-02", *req.Birth)
		if err != nil {
			c.Error(err)
			return
		}
		birth = &t
	}

	u, err := h.Svc.Create(c.Request.Context(), req.Email, req.Name, req.UsedName, req.Company, birth)
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

type updateUserReq struct {
	Name     string  `json:"name" binding:"required"`
	UsedName *string `json:"used_name"`
	Company  *string `json:"company"`
	Birth    *string `json:"birth"`
}

func (h *UserHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req updateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	var birth *time.Time
	if req.Birth != nil {
		t, err := time.Parse("2006-01-02", *req.Birth)
		if err != nil {
			c.Error(err)
			return
		}
		birth = &t
	}

	u, err := h.Svc.Update(c.Request.Context(), id, req.Name, req.UsedName, req.Company, birth)
	if err != nil { c.Error(err); return }
	c.JSON(http.StatusOK, u)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	err := h.Svc.Delete(c.Request.Context(), id)
	if err != nil { c.Error(err); return }
	c.JSON(http.StatusNoContent, nil)
}
