package auth

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

// RDB is the Redis client; set by InitRedis.
var RDB *redis.Client

// InitRedis connects to Redis. Uses REDIS_ADDR (default localhost:6379).
func InitRedis() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	RDB = redis.NewClient(&redis.Options{Addr: addr})
}

func PingRedis() error {
	if RDB == nil {
		return fmt.Errorf("redis not initialized")
	}
	return RDB.Ping(Ctx).Err()
}

const sessionPrefix = "session:"

// SaveSession stores user id for this session id with TTL (same as token lifetime).
func SaveSession(sessionID string, userID int, ttl time.Duration) error {
	key := sessionPrefix + sessionID
	return RDB.Set(Ctx, key, strconv.Itoa(userID), ttl).Err()
}

// ValidateSession checks Redis has this session and the user id matches the JWT.
func ValidateSession(sessionID string, expectedUserID int) error {
	key := sessionPrefix + sessionID
	val, err := RDB.Get(Ctx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("session not found or expired")
	}
	if err != nil {
		return err
	}
	uid, err := strconv.Atoi(val)
	if err != nil {
		return fmt.Errorf("invalid session data")
	}
	if uid != expectedUserID {
		return fmt.Errorf("session user mismatch")
	}
	return nil
}

// DeleteSession removes a session (logout).
func DeleteSession(sessionID string) error {
	return RDB.Del(Ctx, sessionPrefix+sessionID).Err()
}
