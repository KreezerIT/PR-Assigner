package statistics

// UserStatsDTO DTO для статистики пользователя
type UserStatsDTO struct {
	UserID           string `json:"user_id"`
	Username         string `json:"username"`
	TeamName         string `json:"team_name"`
	TotalAssignments int    `json:"total_assignments"`
	ActiveReviews    int    `json:"active_reviews"`
	CompletedReviews int    `json:"completed_reviews"`
}

// TeamStatsDTO DTO для статистики команды
type TeamStatsDTO struct {
	TeamName      string `json:"team_name"`
	TotalMembers  int    `json:"total_members"`
	ActiveMembers int    `json:"active_members"`
	TotalPRs      int    `json:"total_prs"`
	OpenPRs       int    `json:"open_prs"`
	MergedPRs     int    `json:"merged_prs"`
}

// GlobalStatsDTO DTO для общей статистики
type GlobalStatsDTO struct {
	TotalUsers       int `json:"total_users"`
	ActiveUsers      int `json:"active_users"`
	TotalTeams       int `json:"total_teams"`
	TotalPRs         int `json:"total_prs"`
	OpenPRs          int `json:"open_prs"`
	MergedPRs        int `json:"merged_prs"`
	TotalAssignments int `json:"total_assignments"`
}

// TopReviewerDTO DTO для топ ревьювера
type TopReviewerDTO struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	TeamName    string `json:"team_name"`
	ReviewCount int    `json:"review_count"`
}

// Преобразования

func (u *UserStats) ToDTO() *UserStatsDTO {
	return &UserStatsDTO{
		UserID:           u.UserID,
		Username:         u.Username,
		TeamName:         u.TeamName,
		TotalAssignments: u.TotalAssignments,
		ActiveReviews:    u.ActiveReviews,
		CompletedReviews: u.CompletedReviews,
	}
}

func (t *TeamStats) ToDTO() *TeamStatsDTO {
	return &TeamStatsDTO{
		TeamName:      t.TeamName,
		TotalMembers:  t.TotalMembers,
		ActiveMembers: t.ActiveMembers,
		TotalPRs:      t.TotalPRs,
		OpenPRs:       t.OpenPRs,
		MergedPRs:     t.MergedPRs,
	}
}

func (g *GlobalStats) ToDTO() *GlobalStatsDTO {
	return &GlobalStatsDTO{
		TotalUsers:       g.TotalUsers,
		ActiveUsers:      g.ActiveUsers,
		TotalTeams:       g.TotalTeams,
		TotalPRs:         g.TotalPRs,
		OpenPRs:          g.OpenPRs,
		MergedPRs:        g.MergedPRs,
		TotalAssignments: g.TotalAssignments,
	}
}

func (t *TopReviewer) ToDTO() *TopReviewerDTO {
	return &TopReviewerDTO{
		UserID:      t.UserID,
		Username:    t.Username,
		TeamName:    t.TeamName,
		ReviewCount: t.ReviewCount,
	}
}

func TopReviewersToDTO(reviewers []TopReviewer) []TopReviewerDTO {
	dtos := make([]TopReviewerDTO, len(reviewers))
	for i, r := range reviewers {
		dtos[i] = *r.ToDTO()
	}
	return dtos
}
