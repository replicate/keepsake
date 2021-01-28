package errors

import (
	"fmt"
)

const (
	CodeDoesNotExist                  = "DOES_NOT_EXIST"
	CodeReadError                     = "READ_ERROR"
	CodeWriteError                    = "WRITE_ERROR"
	CodeRepositoryConfigurationError  = "REPOSITORY_CONFIGURATION_ERROR"
	CodeIncompatibleRepositoryVersion = "INCOMPATIBLE_REPOSITORY_VERSION"
	CodeCorruptedRepositorySpec       = "CORRUPTED_REPOSITORY_SPEC"
	CodeConfigNotFound                = "CONFIG_NOT_FOUND"
)

// TODO: support wrapping https://blog.golang.org/go1.13-errors
type CodedError interface {
	Code() string
}

type codedError struct {
	code string
	msg  string
}

func (e *codedError) Error() string {
	return e.msg
}

func (e *codedError) Code() string {
	return e.code
}

func IsDoesNotExist(err error) bool {
	return Code(err) == CodeDoesNotExist
}

func IsConfigNotFound(err error) bool {
	return Code(err) == CodeConfigNotFound
}

func DoesNotExist(msg string) error { return &codedError{code: CodeDoesNotExist, msg: msg} }
func ReadError(msg string) error    { return &codedError{code: CodeReadError, msg: msg} }
func WriteError(msg string) error   { return &codedError{code: CodeWriteError, msg: msg} }
func RepositoryConfigurationError(msg string) error {
	return &codedError{code: CodeRepositoryConfigurationError, msg: msg}
}

func ConfigNotFound(msg string) error {
	return &codedError{
		code: CodeConfigNotFound,
		msg: msg + `

You must either create a keepsake.yaml configuration file, or explicitly pass the arguments 'repository' and 'directory' to keepsake.Project().

For more information, see https://keepsake.ai/docs/reference/python"""
`,
	}
}

func IncompatibleRepositoryVersion(rootURL string) error {
	return &codedError{
		code: CodeIncompatibleRepositoryVersion,
		msg: `The repository at ` + rootURL + ` is using a newer storage mechanism which is incompatible with your version of Keepsake.

To upgrade, run:
pip install --upgrade keepsake`,
	}
}

func CorruptedRepositorySpec(rootURL string, specPath string, err error) error {
	return &codedError{
		code: CodeCorruptedRepositorySpec,
		msg: fmt.Sprintf(`The project spec file at %s/%s is corrupted (%v).

You can manually edit it with the format {"version": VERSION},
where VERSION is an integer.`, rootURL, specPath, err),
	}
}

func Code(err error) string {
	if cerr, ok := err.(CodedError); ok {
		return cerr.Code()
	}
	return ""
}
