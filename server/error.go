package server

import (
	"errors"
	"fmt"
)

var errInvalidNumberOfArguments = errors.New("invalid number of arguments")

func errInvalidArgument(arg string) error {
	return fmt.Errorf("invalid argument '%s'", arg)
}
