package pullrequest

import "time"

// CreatePRDTO используется для создания PR
type CreatePRDTO struct {
	PullRequestID   string
	PullRequestName string
	AuthorID        string
}

// DTO содержит полную информацию о PR
type DTO struct {
	PullRequestID     string
	PullRequestName   string
	AuthorID          string
	Status            string
	AssignedReviewers []string
	CreatedAt         *time.Time
	MergedAt          *time.Time
}

// ShortDTO содержит краткую информацию о PR
type ShortDTO struct {
	PullRequestID   string
	PullRequestName string
	AuthorID        string
	Status          string
}

// ReassignResultDTO результат переназначения ревьювера
type ReassignResultDTO struct {
	PR         *DTO
	ReplacedBy string
}

// ToDTO преобразует доменную модель PullRequest в DTO
func (pr *PullRequest) ToDTO() *DTO {
	return &DTO{
		PullRequestID:     pr.PullRequestID,
		PullRequestName:   pr.PullRequestName,
		AuthorID:          pr.AuthorID,
		Status:            pr.Status,
		AssignedReviewers: pr.AssignedReviewers,
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}

// ToShortDTO преобразует доменную модель PullRequestShort в ShortDTO
func (pr *PullRequestShort) ToShortDTO() *ShortDTO {
	return &ShortDTO{
		PullRequestID:   pr.PullRequestID,
		PullRequestName: pr.PullRequestName,
		AuthorID:        pr.AuthorID,
		Status:          pr.Status,
	}
}

// ShortDTOsToDTO преобразует срез моделей PullRequestShort в срез ShortDTO
func ShortDTOsToDTO(prs []PullRequestShort) []ShortDTO {
	dtos := make([]ShortDTO, len(prs))
	for i, pr := range prs {
		dtos[i] = *pr.ToShortDTO()
	}
	return dtos
}
