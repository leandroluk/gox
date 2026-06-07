//go:build ignore

// Demonstrates the init() + selector pattern combined with Multi for aggregation.
// Each provider package self-registers by name in init().
// A selector promotes the configured provider to the unnamed default.
// Connectable aggregates all providers for startup health checks.
package main

import (
	"fmt"
	"os"

	"github.com/leandroluk/gox/di"
)

// --- domain ---

type Connectable interface {
	Connect() error
	Ping() error
	Close() error
}

type Broker interface{ Connectable }
type Database interface{ Connectable }

// --- broker/nats ---

type NatsBroker struct{}

func (b *NatsBroker) Connect() error { fmt.Println("nats: connected"); return nil }
func (b *NatsBroker) Ping() error    { fmt.Println("nats: ping ok"); return nil }
func (b *NatsBroker) Close() error   { fmt.Println("nats: closed"); return nil }

func init() {
	// self-registers by name — always present regardless of selector
	di.Register[Broker](func(b di.Builder[Broker]) {
		b.Named("[broker/nats]", func() (Broker, error) { return &NatsBroker{}, nil }).
			OnStart(func(br Broker) error { return br.Connect() }).
			OnStop(func(br Broker) error { return br.Close() })
	})

	// opts in to Connectable aggregation
	di.Register[Connectable](func(b di.Builder[Connectable]) {
		var broker Broker
		b.Extend(&broker).Multi()
	})
}

// --- db/postgres ---

type PostgresDB struct{}

func (d *PostgresDB) Connect() error { fmt.Println("postgres: connected"); return nil }
func (d *PostgresDB) Ping() error    { fmt.Println("postgres: ping ok"); return nil }
func (d *PostgresDB) Close() error   { fmt.Println("postgres: closed"); return nil }

func init() {
	di.Register[Database](func(b di.Builder[Database]) {
		b.Named("[db/postgres]", func() (Database, error) { return &PostgresDB{}, nil }).
			OnStart(func(db Database) error { return db.Connect() }).
			OnStop(func(db Database) error { return db.Close() })
	})

	di.Register[Connectable](func(b di.Builder[Connectable]) {
		var db Database
		b.Extend(&db).Multi()
	})
}

// --- selectors (driven by config/env) ---

func registerBroker(provider string) {
	switch provider {
	case "nats":
		di.Register[Broker](func(b di.Builder[Broker]) {
			b.New(func() (Broker, error) {
				return di.ResolveNamed[Broker]("[broker/nats]"), nil
			})
		})
	default:
		panic("unknown broker provider: " + provider)
	}
}

func registerDatabase(provider string) {
	switch provider {
	case "postgres":
		di.Register[Database](func(b di.Builder[Database]) {
			b.New(func() (Database, error) {
				return di.ResolveNamed[Database]("[db/postgres]"), nil
			})
		})
	default:
		panic("unknown database provider: " + provider)
	}
}

// --- main ---

func main() {
	registerBroker("nats")
	registerDatabase("postgres")

	// start all — OnStart hooks run in registration order
	if err := di.StartAll(); err != nil {
		fmt.Fprintln(os.Stderr, "start:", err)
		os.Exit(1)
	}

	// health check all Connectables — only Multi-marked entries
	fmt.Println("--- health check ---")
	for _, c := range di.ResolveAll[Connectable]() {
		if err := c.Ping(); err != nil {
			fmt.Fprintln(os.Stderr, "ping failed:", err)
		}
	}

	// stop all — OnStop hooks run in reverse order
	if err := di.StopAll(); err != nil {
		fmt.Fprintln(os.Stderr, "stop:", err)
	}
}
