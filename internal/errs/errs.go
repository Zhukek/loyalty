package errs

import "errors"

var ErrUsernameTaken = errors.New("username already taken")
var ErrNotValidToken = errors.New("token is not valid")
var ErrSigningMethod = errors.New("unexpected signing method")
var ErrNoOrderFound = errors.New("no order found")
var ErrLowBalance = errors.New("low user's balance")
