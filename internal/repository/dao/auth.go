package dao

import (
	"database/sql"
	"time"

	"github.com/Olegsandrik/Exponenta/internal/usecase/models"
)

type User struct {
	ID           uint          `db:"id"`
	VKID         sql.NullInt64 `db:"vk_id"`
	Name         string        `dn:"name"`
	SurName      string        `db:"sur_name"`
	Login        string        `db:"login"`
	PasswordHash string        `db:"password_hash"`
	CreatedAt    time.Time     `db:"created_at"`
	UpdatedAt    time.Time     `db:"updated_at"`
}

func ConvertUserTableToModel(ut []User) []models.User {
	users := make([]models.User, len(ut))
	for i, u := range ut {
		users[i] = models.User{
			ID:           u.ID,
			Name:         u.Name,
			SurName:      u.SurName,
			Login:        u.Login,
			PasswordHash: u.PasswordHash,
			CreatedAt:    u.CreatedAt,
			UpdatedAt:    u.UpdatedAt,
		}
	}
	return users
}
