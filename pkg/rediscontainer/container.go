package rediscontainer

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	rediscnt "github.com/testcontainers/testcontainers-go/modules/redis"
)

func New(ctx context.Context) (*redis.Client, func(context.Context) error, error) {
	rcnt, err := rediscnt.RunContainer(ctx, testcontainers.WithImage("redis:7.0.11-alpine3.17"))
	if err != nil {
		return nil, nil, err
	}
	err = rcnt.Start(ctx)
	if err != nil {
		return nil, nil, err
	}
	uri, err := rcnt.ConnectionString(ctx)
	if err != nil {
		return nil, nil, err
	}
	opts, err := redis.ParseURL(uri)
	if err != nil {
		return nil, nil, err
	}
	rc := redis.NewClient(opts)
	term := func(ctx context.Context) error {
		err := rc.Close()
		if err != nil {
			return fmt.Errorf("failed to close redis connection, %s", err)
		}
		err = rcnt.Terminate(ctx)
		if err != nil {
			return fmt.Errorf("failed to terminate redis container, %s", err)
		}
		return nil
	}
	return rc, term, nil
}
