package arangodb

import (
	"context"
	common "github.com/Nubes3/common/models/arangodb"
	arangoDriver "github.com/arangodb/go-driver"
	"time"
)

const ContextExpiredTime = 30

var (
	userCol arangoDriver.Collection
)

func InitArangoRepo() {
	ctx, cancel := context.WithTimeout(context.Background(), ContextExpiredTime*time.Second)
	defer cancel()

	exist, err := common.ArangoDb.CollectionExists(ctx, "users")
	if err != nil {
		panic(err)
	}

	if !exist {
		userCol, _ = common.ArangoDb.CreateCollection(ctx, "users", &arangoDriver.CreateCollectionOptions{})
	} else {
		userCol, _ = common.ArangoDb.Collection(ctx, "users")
	}
}
