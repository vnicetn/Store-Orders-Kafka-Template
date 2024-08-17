package main

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type OrdersDAO struct {
	c *mongo.Collection
}

func NewOrdersDAO(ctx context.Context, client *mongo.Client) (*OrdersDAO, error) {
	return &OrdersDAO{
		c: client.Database("core").Collection("store_orders"),
	}, nil
}
