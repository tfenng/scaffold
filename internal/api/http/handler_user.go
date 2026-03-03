package http

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tfenng/scaffold/internal/domain"
	"github.com/tfenng/scaffold/internal/repo"
	"github.com/tfenng/scaffold/internal/service"
)

type UserHandler struct{ Svc *service.UserService }

func parsePositiveID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(domain.Invalid("id must be a positive integer"))
		return 0, false
	}
	return id, true
}

func (h *UserHandler) Get(c *gin.Context) {
	id, ok := parsePositiveID(c)
	if !ok {
		return
	}
	u, err := h.Svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, toUserResponse(u))
}

type createUserReq struct {
	Uid      string  `json:"uid" binding:"required"`
	Email    *string `json:"email"`
	Name     string  `json:"name" binding:"required"`
	UsedName *string `json:"used_name"`
	Company  *string `json:"company"`
	Birth    *string `json:"birth"`
}

func (h *UserHandler) Create(c *gin.Context) {
	var req createUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(domain.Invalid(err.Error()))
		return
	}

	birth, err := parseBirth(req.Birth)
	if err != nil {
		c.Error(err)
		return
	}

	u, err := h.Svc.Create(c.Request.Context(), req.Uid, req.Email, req.Name, req.UsedName, req.Company, birth)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, toUserResponse(u))
}

type listUsersQuery struct {
	Email    *string `form:"email"`
	NameLike *string `form:"name_like"`
	Page     int32   `form:"page"`
	PageSize int32   `form:"page_size"`
}

func (h *UserHandler) List(c *gin.Context) {
	var q listUsersQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.Error(domain.Invalid(err.Error()))
		return
	}

	out, err := h.Svc.List(c.Request.Context(), repo.UserListFilter{
		Email: q.Email, NameLike: q.NameLike, Page: q.Page, PageSize: q.PageSize,
	})
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, toUserPageResponse(out))
}

type updateUserReq struct {
	Email    *string `json:"email"`
	Name     string  `json:"name" binding:"required"`
	UsedName *string `json:"used_name"`
	Company  *string `json:"company"`
	Birth    *string `json:"birth"`
}

func (h *UserHandler) Update(c *gin.Context) {
	id, ok := parsePositiveID(c)
	if !ok {
		return
	}

	var req updateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(domain.Invalid(err.Error()))
		return
	}

	birth, err := parseBirth(req.Birth)
	if err != nil {
		c.Error(err)
		return
	}

	u, err := h.Svc.Update(c.Request.Context(), id, req.Email, req.Name, req.UsedName, req.Company, birth)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, toUserResponse(u))
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, ok := parsePositiveID(c)
	if !ok {
		return
	}
	err := h.Svc.Delete(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

func parseBirth(birth *string) (*time.Time, error) {
	if birth == nil {
		return nil, nil
	}
	v := strings.TrimSpace(*birth)
	if v == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return nil, domain.Invalid("birth must be in YYYY-MM-DD format")
	}
	return &t, nil
}
