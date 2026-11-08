package usecase

import (
	"BookingGo/internal/domain"
	"BookingGo/internal/dto"
	"BookingGo/internal/enum"
	"context"

	"golang.org/x/crypto/bcrypt"
)

//go:generate mockgen -source=user_usecase.go -destination=mocks/user_mocks.go -package=mocks
type UserPostgres interface {
	GetAll() ([]domain.User, error)
	GetByID(id int) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	Create(user *domain.User) error
	Update(id int, requestUser map[string]any) (*domain.User, error)
	Delete(id int) error
	EmailExists(email string) (bool, error)
	GetByEmailContext(ctx context.Context, email string) (*domain.User, error)
	GetByIDContext(ctx context.Context, id int) (*domain.User, error)
}

type UserUseCase struct {
	userPostgres UserPostgres
}

func NewUserUseCase(userPostgressitory UserPostgres) *UserUseCase {
	return &UserUseCase{userPostgres: userPostgressitory}
}

func (u *UserUseCase) GetAllUsers() ([]domain.User, error) {
	return u.userPostgres.GetAll()
}

func (u *UserUseCase) GetUserByID(id int) (*domain.User, error) {
	return u.userPostgres.GetByID(id)
}

func (u *UserUseCase) GetUserByEmail(email string) (*domain.User, error) {
	return u.userPostgres.GetByEmail(email)
}

func (u *UserUseCase) GetUserByEmailContext(ctx context.Context, email string) (*domain.User, error) {
	return u.userPostgres.GetByEmailContext(ctx, email)
}

func (u *UserUseCase) CreateUser(req *dto.CreateUserRequest) (*domain.User, error) {
	exists, err := u.userPostgres.EmailExists(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrEmailTaken
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:        req.Email,
		PasswordHash: string(passwordHash),
		FIO:          req.FIO,
		Phone:        req.Phone,
		Role:         enum.RoleClient,
		IsActive:     true,
		UserNotificationSettings: domain.NotificationSettings{
			IsEmailSend: true,
			IsPhoneSend: false,
		},
	}
	if err := u.userPostgres.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserUseCase) UpdateUserProfile(id int, req *dto.UpdateUserProfileRequest) (*domain.User, error) {
	updates := map[string]any{}
	if req.FIO != nil {
		updates["fio"] = *req.FIO
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}

	return u.userPostgres.Update(id, updates)
}

func (u *UserUseCase) UpdateUserData(id int, updates map[string]any) (*domain.User, error) {
	return u.userPostgres.Update(id, updates)
}

func (u *UserUseCase) DeleteUser(id int) error {
	return u.userPostgres.Delete(id)
}

func (u *UserUseCase) EmailExists(email string) (bool, error) {
	return u.userPostgres.EmailExists(email)
}
