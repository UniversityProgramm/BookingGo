package usecase

import (
	"BookingGo/internal/entity"
	"BookingGo/internal/enum"
	"BookingGo/internal/repository"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken   = errors.New("user not found")
	ErrUserNotFound = errors.New("email is taken")
)

type UserRepository interface {
	GetAll() ([]entity.User, error)
	GetByID(id int) (*entity.User, error)
	GetByEmail(email string) (*entity.User, error)
	Create(user *entity.User) error
	Update(id int, requestUser map[string]interface{}) (*entity.User, error)
	Delete(id int) error
	EmailExists(email string) (bool, error)
}

type UserUsecase struct {
	userRepo UserRepository
}

func NewUserUsecase(userRepository UserRepository) *UserUsecase {
	return &UserUsecase{userRepo: userRepository}
}

func (u *UserUsecase) GetAllUsers() ([]entity.User, error) {
	return u.userRepo.GetAll()
}

func (u *UserUsecase) GetUserByID(id int) (*entity.User, error) {
	user, err := u.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (u *UserUsecase) GetUserByEmail(email string) (*entity.User, error) {
	user, err := u.userRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (u *UserUsecase) CreateUser(req *entity.CreateUserRequest) (*entity.User, error) {
	exists, err := u.userRepo.EmailExists(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailTaken
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
	}

	if err := u.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserUsecase) UpdateUser(id int, req *entity.UpdateUserRequest) (*entity.User, error) {
	_, err := u.userRepo.GetByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	updates := map[string]interface{}{}
	if req.Email != nil {
		exists, err := u.userRepo.EmailExists(*req.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrEmailTaken
		}
		updates["email"] = *req.Email
	}
	if req.Password != nil {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		updates["password_hash"] = string(passwordHash)
	}
	if req.FIO != nil {
		updates["fio"] = *req.FIO
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}

	user, err := u.userRepo.Update(id, updates)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserUsecase) DeleteUser(id int) error {
	err := u.userRepo.Delete(id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}
