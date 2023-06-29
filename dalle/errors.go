package dalle

import "errors"

var (
	ErrResponsePromptReview    = errors.New(responsePromptReview)
	ErrResponsePromptBlocked   = errors.New(responsePromptBlocked)
	ErrResponseUnsupportedLang = errors.New(responseUnsupportedLang)
)
