package usecase

import "knot-api/internal/domain"

// ErrNotFound は対象のEntityが見つからなかったことを表す。
// controllerがdomainパッケージへ直接依存しなくて済むよう、domain.ErrNotFoundを再公開する。
var ErrNotFound = domain.ErrNotFound

// ErrForbidden は操作を行う権限が無いことを表す。
// controllerがdomainパッケージへ直接依存しなくて済むよう、domain.ErrForbiddenを再公開する。
var ErrForbidden = domain.ErrForbidden
