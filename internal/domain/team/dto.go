package team

// CreateTeamDTO DTO для создания команды
type CreateTeamDTO struct {
	TeamName string
	Members  []MemberDTO
}

// MemberDTO представляет участника команды в DTO
type MemberDTO struct {
	UserID   string
	Username string
	IsActive bool
}

// DTO содержит информацию о команде
type DTO struct {
	TeamName string
	Members  []MemberDTO
}

// Преобразования

// ToDTO конвертирует доменную модель Team в DTO
func (t *Team) ToDTO() *DTO {
	members := make([]MemberDTO, len(t.Members))
	for i, m := range t.Members {
		members[i] = m.ToDTO()
	}

	return &DTO{
		TeamName: t.TeamName,
		Members:  members,
	}
}

// ToDTO конвертирует доменную модель Member в MemberDTO
func (m *Member) ToDTO() MemberDTO {
	return MemberDTO{
		UserID:   m.UserID,
		Username: m.Username,
		IsActive: m.IsActive,
	}
}

// FromCreateDTO создаёт доменную модель Team из CreateTeamDTO
func FromCreateDTO(dto *CreateTeamDTO) *Team {
	members := make([]Member, len(dto.Members))
	for i, m := range dto.Members {
		members[i] = FromMemberDTO(m)
	}

	return &Team{
		TeamName: dto.TeamName,
		Members:  members,
	}
}

// FromMemberDTO создаёт доменную модель Member из MemberDTO
func FromMemberDTO(dto MemberDTO) Member {
	return Member(dto)
}
