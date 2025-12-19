package db

import (
	"github.com/prisma/prisma-client-go"
)

func New() (prisma.Client, error) {
	client := prisma.NewClient()
	if err := client.Prisma.Connect(); err != nil {
		return nil, err
	}
	return client, nil
}
