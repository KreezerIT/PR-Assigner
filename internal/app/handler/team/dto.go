package team

import "github.com/kreezerit/pr-assigner/internal/domain/team"

// CreateTeamRequest HTTP‑запрос для создания команды
type CreateTeamRequest struct {
	TeamName string          `json:"team_name"`
	Members  []MemberRequest `json:"members"`
}

// MemberRequest участник команды в HTTP‑запросе
type MemberRequest struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

// TeamResponse HTTP‑ответ с данными команды
type TeamResponse struct {
	Team TeamView `json:"team"`
}

// TeamView представление команды в API
type TeamView struct {
	TeamName string       `json:"team_name"`
	Members  []MemberView `json:"members"`
}

// MemberView представление участника команды в API
type MemberView struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

// ToDomainDTO преобразует HTTP‑запрос в DTO доменного слоя
func (r *CreateTeamRequest) ToDomainDTO() *team.CreateTeamDTO {
	members := make([]team.MemberDTO, len(r.Members))
	for i, m := range r.Members {
		members[i] = team.MemberDTO{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}

	return &team.CreateTeamDTO{
		TeamName: r.TeamName,
		Members:  members,
	}
}

// FromDomainDTO преобразует DTO доменного слоя в HTTP‑ответ
func FromDomainDTO(dto *team.DTO) *TeamResponse {
	members := make([]MemberView, len(dto.Members))
	for i, m := range dto.Members {
		members[i] = MemberView{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}

	return &TeamResponse{
		Team: TeamView{
			TeamName: dto.TeamName,
			Members:  members,
		},
	}
}
