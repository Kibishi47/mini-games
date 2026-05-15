package redis

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

func NewClient(ctx context.Context, url, username, password string) (*goredis.Client, error) {
	opts, err := goredis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	if username != "" {
		opts.Username = username
	}
	if password != "" {
		opts.Password = password
	}

	client := goredis.NewClient(opts)

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
