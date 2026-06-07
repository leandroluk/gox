//go:build ignore

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/leandroluk/gox/di"
)

// --- domain ---

type DB interface {
	Query(q string) string
	Ping() error
}

type Cache interface {
	Get(key string) string
	Ping() error
}

type MemDB struct{ connected bool }

func (d *MemDB) Query(q string) string { return "result:" + q }
func (d *MemDB) Ping() error           { return nil }
func (d *MemDB) Connect() error        { d.connected = true; fmt.Println("db connected"); return nil }
func (d *MemDB) Close() error          { d.connected = false; fmt.Println("db closed"); return nil }

type MemCache struct{ connected bool }

func (c *MemCache) Get(key string) string { return "cached:" + key }
func (c *MemCache) Ping() error           { return nil }
func (c *MemCache) Connect() error        { c.connected = true; fmt.Println("cache connected"); return nil }
func (c *MemCache) Close() error          { c.connected = false; fmt.Println("cache closed"); return nil }

func main() {
	// DB registered first — starts first, stops last
	di.Register[DB](func(b di.Builder[DB]) {
		b.New(func() (DB, error) { return &MemDB{}, nil }).
			OnStart(func(d DB) error { return d.(*MemDB).Connect() }).
			OnStop(func(d DB) error { return d.(*MemDB).Close() })
	})

	// Cache registered second — starts second, stops before DB
	di.Register[Cache](func(b di.Builder[Cache]) {
		b.New(func() (Cache, error) { return &MemCache{}, nil }).
			OnStart(func(c Cache) error { return c.(*MemCache).Connect() }).
			OnStop(func(c Cache) error { return c.(*MemCache).Close() })
	})

	if err := di.StartAll(); err != nil {
		fmt.Fprintln(os.Stderr, "start:", err)
		os.Exit(1)
	}

	fmt.Println(di.Resolve[DB]().Query("select 1"))

	// graceful shutdown on SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if err := di.StopAll(); err != nil {
		fmt.Fprintln(os.Stderr, "stop:", err)
	}
}
