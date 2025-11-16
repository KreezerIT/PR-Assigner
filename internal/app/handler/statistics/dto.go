package statistics

import "github.com/kreezerit/pr-assigner/internal/domain/statistics"

// UserStatsResponse HTTP-ответ для статистики пользователя
type UserStatsResponse struct {
	UserID           string `json:"user_id"`
	Username         string `json:"username"`
	TeamName         string `json:"team_name"`
	TotalAssignments int    `json:"total_assignments"`
	ActiveReviews    int    `json:"active_reviews"`
	CompletedReviews int    `json:"completed_reviews"`
}

// TeamStatsResponse HTTP-ответ для статистики команды
type TeamStatsResponse struct {
	TeamName      string `json:"team_name"`
	TotalMembers  int    `json:"total_members"`
	ActiveMembers int    `json:"active_members"`
	TotalPRs      int    `json:"total_prs"`
	OpenPRs       int    `json:"open_prs"`
	MergedPRs     int    `json:"merged_prs"`
}

// GlobalStatsResponse HTTP-ответ для общей статистики
type GlobalStatsResponse struct {
	TotalUsers       int `json:"total_users"`
	ActiveUsers      int `json:"active_users"`
	TotalTeams       int `json:"total_teams"`
	TotalPRs         int `json:"total_prs"`
	OpenPRs          int `json:"open_prs"`
	MergedPRs        int `json:"merged_prs"`
	TotalAssignments int `json:"total_assignments"`
}

// TopReviewersResponse HTTP-ответ для топ ревьюверов
type TopReviewersResponse struct {
	TopReviewers []TopReviewerView `json:"top_reviewers"`
}

type TopReviewerView struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	TeamName    string `json:"team_name"`
	ReviewCount int    `json:"review_count"`
}

// Преобразования из domain DTO в HTTP-ответы

func FromUserStatsDTO(dto *statistics.UserStatsDTO) *UserStatsResponse {
	return &UserStatsResponse{
		UserID:           dto.UserID,
		Username:         dto.Username,
		TeamName:         dto.TeamName,
		TotalAssignments: dto.TotalAssignments,
		ActiveReviews:    dto.ActiveReviews,
		CompletedReviews: dto.CompletedReviews,
	}
}

func FromTeamStatsDTO(dto *statistics.TeamStatsDTO) *TeamStatsResponse {
	return &TeamStatsResponse{
		TeamName:      dto.TeamName,
		TotalMembers:  dto.TotalMembers,
		ActiveMembers: dto.ActiveMembers,
		TotalPRs:      dto.TotalPRs,
		OpenPRs:       dto.OpenPRs,
		MergedPRs:     dto.MergedPRs,
	}
}

func FromGlobalStatsDTO(dto *statistics.GlobalStatsDTO) *GlobalStatsResponse {
	return &GlobalStatsResponse{
		TotalUsers:       dto.TotalUsers,
		ActiveUsers:      dto.ActiveUsers,
		TotalTeams:       dto.TotalTeams,
		TotalPRs:         dto.TotalPRs,
		OpenPRs:          dto.OpenPRs,
		MergedPRs:        dto.MergedPRs,
		TotalAssignments: dto.TotalAssignments,
	}
}

func FromTopReviewersDTOs(dtos []statistics.TopReviewerDTO) *TopReviewersResponse {
	views := make([]TopReviewerView, len(dtos))
	for i, dto := range dtos {
		views[i] = TopReviewerView{
			UserID:      dto.UserID,
			Username:    dto.Username,
			TeamName:    dto.TeamName,
			ReviewCount: dto.ReviewCount,
		}
	}
	return &TopReviewersResponse{TopReviewers: views}
}
