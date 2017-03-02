package main

type (
	InvalidJobError struct {
		msg string
	}
)

func (e InvalidJobError) Error() string {
	return e.msg
}

