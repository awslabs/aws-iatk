package publicrpc

import "fmt"

type ErrNoMethodFound struct {
	err string
}

func (e *ErrNoMethodFound) Error() string {
	return fmt.Sprintf("No Method found: %s", e.err)
}

type ErrParameter struct {
	ParentErr error
}

func (e *ErrParameter) Error() string {
	if e.ParentErr != nil {
		return fmt.Sprintf("Invalid Parameter Serialization: %v", e.ParentErr)
	}
	return "no parameters found"
}
