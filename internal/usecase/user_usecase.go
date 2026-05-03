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
	userRepository UserRepository
}

func NewUserUsecase(userRepository UserRepository) *UserUsecase {
	return &UserUsecase{userRepository: userRepository}
}

func (u *UserUsecase) GetAllUsers() ([]entity.User, error) {
	return u.userRepository.GetAll()
}

func (u *UserUsecase) GetUserByID(id int) (*entity.User, error) {
	return u.userRepository.GetByID(id)
}

func (u *UserUsecase) GetUserByEmail(email string) (*entity.User, error) {
	user, err := u.userRepository.GetByEmail(email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (u *UserUsecase) CreateUser(req *entity.CreateUserRequest) (*entity.User, error) {
	exists, err := u.userRepository.EmailExists(req.Email)
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

	if err := u.userRepository.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserUsecase) UpdateUser(id int, req *entity.UpdateUserRequest) (*entity.User, error) {
	_, err := u.userRepository.GetByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	updates := map[string]interface{}{}
	if req.Email != nil {
		exists, err := u.userRepository.EmailExists(*req.Email)
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

	user, err := u.userRepository.Update(id, updates)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserUsecase) DeleteUser(id int) error {
	err := u.userRepository.Delete(id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}
