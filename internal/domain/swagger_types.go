// Package domain is used for swagger auto doc
package domain

type UserWithoutId struct {
	Email      string `json:"email"`
	Password   []byte `json:"password"`
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	MiddleName string `json:"middleName"`
	Role       string `json:"role"`
}

type UserWithoutPassword struct {
	Email      string `json:"email"`
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	MiddleName string `json:"middleName"`
	Role       string `json:"role"`
}
