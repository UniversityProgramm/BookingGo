package enum

type Role string

const (
	RoleClient Role = "client"
	RoleStaff  Role = "staff"
	RoleAdmin  Role = "admin"
)
