package entity

import "time"

type User struct {
	ID        int       `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"password" db:"password"`
	FIO       string    `json:"fio" db:"fio"`
	Phone     string    `json:"phone" db:"phone"`
	Role      string    `json:"role" db:"role"`
	CreatedAt time.Time `json:"createdAt" db:"createdAt"`
}
