package pullrequest

import (
	"time"

	"github.com/kreezerit/pr-assigner/internal/domain/pullrequest"
)

// CreatePRRequest HTTP‑запрос для создания PR
type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

// MergePRRequest HTTP‑запрос для merge PR
type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

// ReassignRequest HTTP‑запрос для переназначения ревьювера
type ReassignRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

// PRResponse HTTP‑ответ с данными PR
type PRResponse struct {
	PR PRView `json:"pr"`
}

// PRView представление PR в API
type PRView struct {
	PullRequestID     string   `json:"pull_request_id"`
	PullRequestName   string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
	CreatedAt         *string  `json:"createdAt,omitempty"`
	MergedAt          *string  `json:"mergedAt,omitempty"`
}

// ReassignResponse HTTP‑ответ при переназначении ревьювера
type ReassignResponse struct {
	PR         PRView `json:"pr"`
	ReplacedBy string `json:"replaced_by"`
}

// UserReviewsResponse HTTP‑ответ со списком PR пользователя
type UserReviewsResponse struct {
	UserID       string        `json:"user_id"`
	PullRequests []PRShortView `json:"pull_requests"`
}

// PRShortView краткое представление PR
type PRShortView struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

// ToDomainDTO преобразует HTTP‑запрос в DTO доменного слоя
func (r *CreatePRRequest) ToDomainDTO() *pullrequest.CreatePRDTO {
	return &pullrequest.CreatePRDTO{
		PullRequestID:   r.PullRequestID,
		PullRequestName: r.PullRequestName,
		AuthorID:        r.AuthorID,
	}
}

// FromDomainDTO преобразует DTO доменного слоя в HTTP‑ответ
func FromDomainDTO(dto *pullrequest.DTO) *PRResponse {
	var createdAt, mergedAt *string

	if dto.CreatedAt != nil {
		t := dto.CreatedAt.Format(time.RFC3339)
		createdAt = &t
	}

	if dto.MergedAt != nil {
		t := dto.MergedAt.Format(time.RFC3339)
		mergedAt = &t
	}

	return &PRResponse{
		PR: PRView{
			PullRequestID:     dto.PullRequestID,
			PullRequestName:   dto.PullRequestName,
			AuthorID:          dto.AuthorID,
			Status:            dto.Status,
			AssignedReviewers: dto.AssignedReviewers,
			CreatedAt:         createdAt,
			MergedAt:          mergedAt,
		},
	}
}

// FromReassignResultDTO преобразует результат переназначения в HTTP‑ответ
func FromReassignResultDTO(dto *pullrequest.ReassignResultDTO) *ReassignResponse {
	prResponse := FromDomainDTO(dto.PR)
	return &ReassignResponse{
		PR:         prResponse.PR,
		ReplacedBy: dto.ReplacedBy,
	}
}

// FromShortDTOs преобразует список кратких DTO в представления API
func FromShortDTOs(dtos []pullrequest.ShortDTO) []PRShortView {
	views := make([]PRShortView, len(dtos))
	for i, dto := range dtos {
		views[i] = PRShortView{
			PullRequestID:   dto.PullRequestID,
			PullRequestName: dto.PullRequestName,
			AuthorID:        dto.AuthorID,
			Status:          dto.Status,
		}
	}
	return views
}
