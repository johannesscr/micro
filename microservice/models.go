package microservice

import "github.com/google/uuid"

type User struct {
	UUID      uuid.UUID `json:"uuid"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
}
