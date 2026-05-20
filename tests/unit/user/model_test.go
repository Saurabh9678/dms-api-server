package user_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"infiour.local/dms-api-server/internal/modules/user"
)

func TestUserTableName(t *testing.T) {
	u := &user.User{}
	assert.Equal(t, "users", u.TableName())
}

func TestUserRoleTableName(t *testing.T) {
	ur := &user.UserRole{}
	assert.Equal(t, "user_roles", ur.TableName())
}

func TestUserShowroomTableName(t *testing.T) {
	us := &user.UserShowroom{}
	assert.Equal(t, "user_showroom_relations", us.TableName())
}
