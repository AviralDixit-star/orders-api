package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/AviralDixit-star/orders-api/model"
	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	Client *redis.Client
}

func orderIDKey(id uint64) string {
	return fmt.Sprintf("order:%d", id)
}

func (r *RedisRepo) Insert(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to encode: %w", err)
	}
	key := orderIDKey(order.OrderID)

	txn := r.Client.TxPipeline()

	res := txn.SetNX(ctx, key, string(data), 0)
	if err := res.Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to set: %w", err)
	}

	err = txn.SAdd(ctx, "orders", key).Err()
	if err != nil {
		txn.Discard()
		return fmt.Errorf("failed to add to order set: %w", err)
	}

	_, err = txn.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to exce %w", err)
	}

	return nil
}

var ErrNotExist = errors.New("order does not exist")

func (r *RedisRepo) FindByID(ctx context.Context, id uint64) (model.Order, error) {
	key := orderIDKey(id)

	val, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return model.Order{}, ErrNotExist
	} else if err != nil {
		return model.Order{}, fmt.Errorf("get error %w", err)
	}

	var order model.Order
	err = json.Unmarshal([]byte(val), &order)
	if err != nil {
		return model.Order{}, fmt.Errorf("fail to decode %w", err)
	}

	return order, nil
}

func (r *RedisRepo) DeleteByID(ctx context.Context, id uint64) error {
	key := orderIDKey(id)

	txn := r.Client.TxPipeline()

	err := txn.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		txn.Discard()
		return ErrNotExist
	} else if err != nil {
		txn.Discard()
		return fmt.Errorf("get error %w", err)
	}

	err = txn.SRem(ctx, "orders", key).Err()
	if err != nil {
		return fmt.Errorf("failed to remove from order set: %w", err)
	}
	_, err = txn.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to exce %w", err)
	}

	return nil
}

func (r *RedisRepo) Update(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to encode: %w", err)
	}
	key := orderIDKey(order.OrderID)

	err = r.Client.SetXX(ctx, key, string(data), 0).Err()
	if errors.Is(err, redis.Nil) {
		return fmt.Errorf("failed to set: %w", err)
	} else if err != nil {
		return fmt.Errorf("set order %w", err)
	}

	return nil
}

type FindAllPage struct {
	Size   uint64
	OffSet uint64
}

type FindResult struct {
	Order  []model.Order
	Cursor uint64
}

func (r *RedisRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	res := r.Client.SScan(ctx, "orders", page.OffSet, "*", int64(page.Size))

	keys, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get order ids: %w", err)
	}

	if keys == nil {
		return FindResult{}, fmt.Errorf("No key is found: %w", err)
	}

	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get orders: %w", err)
	}

	orders := make([]model.Order, len(xs))

	for _, va := range xs {
		val := va.(string)

		var order model.Order
		err := json.Unmarshal([]byte(val), &order)
		if err != nil {
			return FindResult{}, fmt.Errorf("failed to decode order json: %w", err)
		}

		orders = append(orders, order)
	}
	return FindResult{Order: orders, Cursor: cursor}, nil
}
