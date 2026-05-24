package usecase

import (
	"BookingGo/internal/domain"
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	GetAll() ([]entity.User, error)
	GetByID(id int) (*entity.User, error)
	GetByEmail(email string) (*entity.User, error)
	Create(user *entity.User) error
	Update(id int, requestUser map[string]any) (*entity.User, error)
	Delete(id int) error
	EmailExists(email string) (bool, error)
}

type UserUseCase struct {
	userRepo UserRepository
}

func NewUserUseCase(userRepository UserRepository) *UserUseCase {
	return &UserUseCase{userRepo: userRepository}
}

func (u *UserUseCase) GetAllUsers() ([]entity.User, error) {
	return u.userRepo.GetAll()
}

func (u *UserUseCase) GetUserByID(id int) (*entity.User, error) {
	return u.userRepo.GetByID(id)
}

func (u *UserUseCase) GetUserByEmail(email string) (*entity.User, error) {
	return u.userRepo.GetByEmail(email)
}

func (u *UserUseCase) CreateUser(req *entity.CreateUserRequest) (*entity.User, error) {
	exists, err := u.userRepo.EmailExists(req.Email)
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

	user := &entity.User{
		Email:        req.Email,
		PasswordHash: string(passwordHash),
		FIO:          req.FIO,
		Phone:        req.Phone,
		Role:         enum.RoleClient,
		IsActive:     true,
		UserNotificationSettings: entity.NotificationSettings{
			IsEmailSend: true,
			IsPhoneSend: false,
		},
	}
	if err := u.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserUseCase) UpdateUserProfile(id int, req *entity.UpdateUserRequest) (*entity.User, error) {
	_, err := u.userRepo.GetByID(id)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	updates := map[string]any{}
	if req.FIO != nil {
		updates["fio"] = *req.FIO
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}

	return u.userRepo.Update(id, updates)
}

func (u *UserUseCase) UpdateUserData(id int, updates map[string]any) (*entity.User, error) {
	_, err := u.userRepo.GetByID(id)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	return u.userRepo.Update(id, updates)
}

func (u *UserUseCase) DeleteUser(id int) error {
	return u.userRepo.Delete(id)
}

func (u *UserUseCase) EmailExists(email string) (bool, error) {
	return u.userRepo.EmailExists(email)
}
