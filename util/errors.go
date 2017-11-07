package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/facebookgo/stack"
)

var (
	ERR_NO_LOGIN = errors.New("No server configuration found. Did you login?")
)

type Error struct {
	inner      error
	fields     map[string]string
	multiStack *stack.Multi
}

func CheckErr(err error, check error) bool {
	if Er, ok := err.(*Error); ok {
		if tmp := Er.Inner(); tmp != nil {
			err = tmp
		}
	}
	return err == check
}

func CheckFieldErr(err error, field, val string) bool {
	if Er, ok := err.(*Error); ok {
		if tmp, ok := Er.fields[field]; !ok {
			return false
		} else if tmp != val {
			return false
		}
		return true
	}
	return false
}

func GetStack(err error) string {
	me, ok := err.(*Error)
	if !ok {
		return fmt.Sprintf("Error [%s] is not an stackable error", err)
	}
	return fmt.Sprintf("%s:\n%s", me.Error(), me.multiStack)
}

func (e Error) Error() string {
	if len(e.fields) > 0 {
		return fmt.Sprintf("Error in fields: %s", e.fields)
	}
	return e.inner.Error()
}

func (e Error) ErrorWithStack() string {
	return fmt.Sprintf("%s\n%s", e.Error(), e.multiStack)
}

func (e *Error) MultiStack() *stack.Multi {
	return e.multiStack
}

func (e *Error) Match(o error) bool {
	return e.inner == o
}

func (e *Error) Inner() error {
	return e.inner
}

func NewErrorFrom(err error) error {
	if err == nil {
		return nil
	}
	if c, ok := err.(*Error); ok {
		return c
	}
	return &Error{
		err,
		make(map[string]string),
		stack.CallersMulti(1),
	}
}

func NewErrorFields(s ...string) error {
	e := &Error{
		nil,
		make(map[string]string),
		stack.CallersMulti(1),
	}
	for i := 0; i+1 < len(s); i++ {
		e.fields[s[i]] = s[i+1]
	}
	e.multiStack = stack.CallersMulti(1)
	return e
}

func NewError(inner string) error {
	return &Error{
		errors.New(inner),
		make(map[string]string),
		stack.CallersMulti(1),
	}
}

func NewErrorf(f string, v ...interface{}) error {
	return &Error{
		fmt.Errorf(f, v...),
		make(map[string]string),
		stack.CallersMulti(1),
	}
}

func (e *Error) SetFieldError(f, er string) {
	e.fields[f] = er
}

func (e *Error) Camo() error {
	if e.Empty() {
		return nil
	}
	return e
}

func (e Error) Empty() bool {
	return e.inner == nil && len(e.fields) == 0
}

func (r Error) MarshalJSON() ([]byte, error) {
	f := make([]string, 0)
	if r.inner != nil {
		m, err := json.Marshal(r.inner.Error())
		if err != nil {
			return nil, err
		}
		f = append(f, `"error":`+string(m))
	}
	if len(r.fields) > 0 {
		m, err := json.Marshal(r.fields)
		if err != nil {
			return nil, err
		}
		f = append(f, `"error_fields":`+string(m))
	}
	return []byte("{" + strings.Join(f, ",") + "}"), nil
}

type hasMultiStack interface {
	MultiStack() *stack.Multi
}

func WrapSkip(err error, skip int) error {
	// nil errors are returned back as nil.
	if err == nil {
		return nil
	}

	// we're adding another Stack to an already wrapped error.
	if se, ok := err.(hasMultiStack); ok {
		se.MultiStack().AddCallers(skip + 1)
		return err
	}

	// we're create a freshly wrapped error.
	return NewErrorFrom(err)
}
