// Package domain is used for swagger auto doc
package domain

import (
	"time"
)

type ActorWithFilms struct {
	ID        int                 `json:"id"`
	Name      string              `json:"name"`
	Sex       Sex                 `json:"sex"`
	Birthdate time.Time           `json:"birthdate" format:"date"`
	Films     []FilmWithoutActors `json:"films"`
}

type ActorToAdd struct {
	Name      string    `json:"name"`
	Sex       Sex       `json:"sex"`
	Birthdate time.Time `json:"birthdate" format:"date"`
}

type ActorToFilmAdd struct {
	ID int `json:"id"`
}

type ActorWithoutFilms struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Sex       Sex       `json:"sex"`
	Birthdate time.Time `json:"birthdate" format:"date"`
}

type FilmWithoutActors struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ReleaseDate time.Time `json:"releaseDate" format:"date"`
	Rating      float64   `json:"rating"`
}

type FilmToAdd struct {
	Title       string           `json:"title"`
	Description string           `json:"description"`
	ReleaseDate time.Time        `json:"releaseDate" format:"date"`
	Rating      float64          `json:"rating"`
	Actors      []ActorToFilmAdd `json:"actors"`
}
