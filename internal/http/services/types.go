package services

import (
	"github.com/ahmedbaig/go-backend-db"
	"github.com/ahmedbaig/go-backend-db-handler/types/demo"
)

type PostsService struct {
	posts *db.Service[demo.Post]
}
