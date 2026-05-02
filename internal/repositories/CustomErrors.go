package repositories

import "errors"

var ErrEmailTaken = errors.New("email is taken")
var ErrUserNotFound = errors.New("user not found")
