package validator

import (
	"strconv"

	"github.com/kreezerit/pr-assigner/internal/domain/team"
)

// ValidateCreateTeamRequest валидация запроса создания команды
func ValidateCreateTeamRequest(teamName string, members []team.Member) error {
	cv := NewCompositeValidator()

	cv.Add(Required("team_name", teamName))
	cv.Add(MinLength("team_name", teamName, 1))
	cv.Add(MaxLength("team_name", teamName, 255))
	cv.Add(NotEmptySlice("members", len(members)))

	// Валидация каждого участника команды
	for i, member := range members {
		prefix := "members[" + strconv.Itoa(i) + "]."
		cv.Add(Required(prefix+"user_id", member.UserID))
		cv.Add(Required(prefix+"username", member.Username))
		cv.Add(MaxLength(prefix+"user_id", member.UserID, 255))
		cv.Add(MaxLength(prefix+"username", member.Username, 255))
	}

	return cv.Validate()
}

// ValidateGetTeamRequest валидация запроса получения команды
func ValidateGetTeamRequest(teamName string) error {
	cv := NewCompositeValidator()
	cv.Add(Required("team_name", teamName))
	return cv.Validate()
}

// ValidateSetIsActiveRequest валидация запроса изменения активности
func ValidateSetIsActiveRequest(userID string) error {
	cv := NewCompositeValidator()
	cv.Add(Required("user_id", userID))
	cv.Add(MaxLength("user_id", userID, 255))
	return cv.Validate()
}

// ValidateCreatePRRequest валидация запроса создания PR
func ValidateCreatePRRequest(prID, prName, authorID string) error {
	cv := NewCompositeValidator()

	cv.Add(Required("pull_request_id", prID))
	cv.Add(Required("pull_request_name", prName))
	cv.Add(Required("author_id", authorID))
	cv.Add(MaxLength("pull_request_id", prID, 255))
	cv.Add(MaxLength("pull_request_name", prName, 500))
	cv.Add(MaxLength("author_id", authorID, 255))

	return cv.Validate()
}

// ValidateMergePRRequest валидация запроса merge PR
func ValidateMergePRRequest(prID string) error {
	cv := NewCompositeValidator()
	cv.Add(Required("pull_request_id", prID))
	cv.Add(MaxLength("pull_request_id", prID, 255))
	return cv.Validate()
}

// ValidateReassignRequest валидация запроса переназначения
func ValidateReassignRequest(prID, oldUserID string) error {
	cv := NewCompositeValidator()

	cv.Add(Required("pull_request_id", prID))
	cv.Add(Required("old_user_id", oldUserID))
	cv.Add(MaxLength("pull_request_id", prID, 255))
	cv.Add(MaxLength("old_user_id", oldUserID, 255))

	return cv.Validate()
}

// ValidateGetUserReviewsRequest валидация запроса получения PR пользователя
func ValidateGetUserReviewsRequest(userID string) error {
	cv := NewCompositeValidator()
	cv.Add(Required("user_id", userID))
	cv.Add(MaxLength("user_id", userID, 255))
	return cv.Validate()
}
