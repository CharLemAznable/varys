package gql

import "fmt"

type UnknownConnectionName struct {
    Name string
}

func (e *UnknownConnectionName) Error() string {
    return fmt.Sprintf("gql: Unknown connection named: %s", e.Name)
}
