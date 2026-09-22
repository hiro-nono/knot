package domain

import "errors"

// ErrNotFound は対象のEntityが見つからなかったことを表す。
var ErrNotFound = errors.New("entity not found")

// ErrForbidden は操作を行う権限が無いことを表す。
var ErrForbidden = errors.New("forbidden")
