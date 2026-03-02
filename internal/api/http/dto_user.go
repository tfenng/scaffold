package http

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	sqlc "github.com/tfenng/scaffold/internal/gen/sqlc"
	"github.com/tfenng/scaffold/internal/repo"
)

type userResponse struct {
	ID        int64   `json:"id"`
	Uid       string  `json:"uid"`
	Email     *string `json:"email"`
	Name      string  `json:"name"`
	UsedName  *string `json:"used_name"`
	Company   *string `json:"company"`
	Birth     *string `json:"birth"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

func toUserResponse(u sqlc.User) userResponse {
	return userResponse{
		ID:        u.ID,
		Uid:       u.Uid,
		Email:     textPtr(u.Email),
		Name:      u.Name,
		UsedName:  textPtr(u.UsedName),
		Company:   textPtr(u.Company),
		Birth:     datePtr(u.Birth),
		CreatedAt: timestampString(u.CreatedAt),
		UpdatedAt: timestampString(u.UpdatedAt),
	}
}

func toUserPageResponse(p repo.Page[sqlc.User]) repo.Page[userResponse] {
	items := make([]userResponse, len(p.Items))
	for i, item := range p.Items {
		items[i] = toUserResponse(item)
	}

	return repo.Page[userResponse]{
		Items:      items,
		Total:      p.Total,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalPages: p.TotalPages,
	}
}

func textPtr(v pgtype.Text) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}

func datePtr(v pgtype.Date) *string {
	if !v.Valid {
		return nil
	}
	s := v.Time.Format("2006-01-02")
	return &s
}

func timestampString(v pgtype.Timestamptz) string {
	if !v.Valid {
		return ""
	}
	return v.Time.UTC().Format(time.RFC3339)
}
