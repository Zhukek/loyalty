package errs

import "errors"

var ErrUsernameTaken = errors.New("username already taken")
