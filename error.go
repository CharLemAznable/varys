package main

import (
    "fmt"
)

type UnexpectedError struct {
    Message string
}

func (e *UnexpectedError) Error() string {
    return fmt.Sprintf("UnexpectedError: %s", e.Message)
}