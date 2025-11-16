package validator

import (
	"testing"

	"github.com/kreezerit/pr-assigner/internal/domain/team"
	"github.com/stretchr/testify/assert"
)

func TestValidateCreateTeamRequest_Success(t *testing.T) {
	members := []team.Member{
		{UserID: "u1", Username: "Alice", IsActive: true},
		{UserID: "u2", Username: "Bob", IsActive: true},
	}

	err := ValidateCreateTeamRequest("backend", members)
	assert.NoError(t, err)
}

func TestValidateCreateTeamRequest_EmptyTeamName(t *testing.T) {
	members := []team.Member{
		{UserID: "u1", Username: "Alice", IsActive: true},
	}

	err := ValidateCreateTeamRequest("", members)
	assert.Error(t, err)

	var valErrs ValidationErrors
	assert.ErrorAs(t, err, &valErrs)
}

func TestValidateCreateTeamRequest_NoMembers(t *testing.T) {
	err := ValidateCreateTeamRequest("backend", []team.Member{})
	assert.Error(t, err)
}

func TestValidateCreateTeamRequest_MemberMissingFields(t *testing.T) {
	members := []team.Member{
		{UserID: "", Username: "Alice", IsActive: true},
	}

	err := ValidateCreateTeamRequest("backend", members)
	assert.Error(t, err)
}

func TestValidateGetTeamRequest_Success(t *testing.T) {
	err := ValidateGetTeamRequest("backend")
	assert.NoError(t, err)
}

func TestValidateGetTeamRequest_Empty(t *testing.T) {
	err := ValidateGetTeamRequest("")
	assert.Error(t, err)
}

func TestValidateSetIsActiveRequest_Success(t *testing.T) {
	err := ValidateSetIsActiveRequest("u1")
	assert.NoError(t, err)
}

func TestValidateSetIsActiveRequest_Empty(t *testing.T) {
	err := ValidateSetIsActiveRequest("")
	assert.Error(t, err)
}

func TestValidateCreatePRRequest_Success(t *testing.T) {
	err := ValidateCreatePRRequest("pr-1", "Add feature", "u1")
	assert.NoError(t, err)
}

func TestValidateCreatePRRequest_MissingFields(t *testing.T) {
	err := ValidateCreatePRRequest("", "", "")
	assert.Error(t, err)

	var valErrs ValidationErrors
	assert.ErrorAs(t, err, &valErrs)
	assert.Len(t, valErrs, 3) // Все три поля отсутствуют
}

func TestValidateMergePRRequest_Success(t *testing.T) {
	err := ValidateMergePRRequest("pr-1")
	assert.NoError(t, err)
}

func TestValidateMergePRRequest_Empty(t *testing.T) {
	err := ValidateMergePRRequest("")
	assert.Error(t, err)
}

func TestValidateReassignRequest_Success(t *testing.T) {
	err := ValidateReassignRequest("pr-1", "u2")
	assert.NoError(t, err)
}

func TestValidateReassignRequest_MissingFields(t *testing.T) {
	err := ValidateReassignRequest("", "")
	assert.Error(t, err)

	var valErrs ValidationErrors
	assert.ErrorAs(t, err, &valErrs)
	assert.Len(t, valErrs, 2)
}

func TestValidateGetUserReviewsRequest_Success(t *testing.T) {
	err := ValidateGetUserReviewsRequest("u1")
	assert.NoError(t, err)
}

func TestValidateGetUserReviewsRequest_Empty(t *testing.T) {
	err := ValidateGetUserReviewsRequest("")
	assert.Error(t, err)
}
