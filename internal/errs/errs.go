package errs

import "errors"

var ErrUsernameTaken = errors.New("username already taken")
var ErrNotValidToken = errors.New("token is not valid")
var ErrSigningMethod = errors.New("unexpected signing method")
