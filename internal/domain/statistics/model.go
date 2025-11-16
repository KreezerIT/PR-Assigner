package statistics

// UserStats статистика по пользователю
type UserStats struct {
	UserID           string
	Username         string
	TeamName         string
	TotalAssignments int
	ActiveReviews    int
	CompletedReviews int
}

// TeamStats статистика по команде
type TeamStats struct {
	TeamName      string
	TotalMembers  int
	ActiveMembers int
	TotalPRs      int
	OpenPRs       int
	MergedPRs     int
}

// GlobalStats общая статистика
type GlobalStats struct {
	TotalUsers       int
	ActiveUsers      int
	TotalTeams       int
	TotalPRs         int
	OpenPRs          int
	MergedPRs        int
	TotalAssignments int
}

// TopReviewer топ ревьювер
type TopReviewer struct {
	UserID      string
	Username    string
	TeamName    string
	ReviewCount int
}
