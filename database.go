package main

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	Collection *mongo.Collection
	MongoDB *mongo.Client
}

func (d *Database) Connect(ctx context.Context) error {
	Env, err := GetEnv()
	if err != nil {
		return err
	}
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)

	d.MongoDB, err = mongo.Connect(ctx, options.Client().
		ApplyURI(Env.MongoURI).SetServerAPIOptions(serverAPIOptions))
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) Disconnect(ctx context.Context) error {
	if err := d.MongoDB.Disconnect(ctx); err != nil {
			return err
	}
	return nil
}

func (d *Database) SwitchTo(Database string, Collection string) {
	d.Collection = d.MongoDB.Database(Database).Collection(Collection)
}
