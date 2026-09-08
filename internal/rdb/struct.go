package rdb

// import (
// 	"context"
// 	"reflect"
// 	"time"

// 	"github.com/redis/go-redis/v9"
// )

// func (c *Client) ToRedisSet(ctx context.Context, s interface{}, key string) error {
// 	// Получаем элементы структуры
// 	val := reflect.ValueOf(s).Elem()

// 	// Создаем функцию для записи структуры в хранилище
// 	settter := func(p redis.Pipeliner) error {
// 		// Итерируемся по полям структуры
// 		for i := 0; i < val.NumField(); i++ {
// 			field := val.Type().Field(i)
// 			// Получаем содержимое тэга redis
// 			tag := field.Tag.Get("redis")
// 			// Записываем значение поля и содержимое тэга redis в хранилище
// 			if err := p.HSet(ctx, key, tag, val.Field(i).Interface()).Err(); err != nil {
// 				return err
// 			}
// 		}
// 		// Задаем время хранения 30 секунд
// 		if err := p.Expire(ctx, key, 30*time.Second).Err(); err != nil {
// 			return err
// 		}
// 		return nil
// 	}

// 	// Сохраняем структуру в хранилище
// 	if _, err := c.rdb.Pipelined(ctx, settter); err != nil {
// 		return err
// 	}

// 	return nil
// }
