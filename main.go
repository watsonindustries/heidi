package main

import (
	"context"
	"fmt"
	"log"

	"github.com/watsonindustries/heidi/boetea/mongo"
	"github.com/watsonindustries/heidi/config"
)

func main() {
	fmt.Println("Hello world")

	env, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	store, err := mongo.New(ctx, env.MONGODB_URI, env.MONGODB_NAME)
	if err != nil {
		log.Fatal(err)
	}

	store.Init(ctx)
	store.Close(ctx)
}
