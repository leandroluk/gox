//go:build ignore

package main

import (
	"fmt"

	"github.com/leandroluk/gox/di"
)

// --- domain ---

type Logger interface {
	Log(msg string)
}

type StdLogger struct{}

func (l *StdLogger) Log(msg string) { fmt.Println(msg) }

type Repo interface {
	Find(id int) string
}

type MemRepo struct {
	data map[int]string
	log  Logger
}

func (r *MemRepo) Find(id int) string {
	r.log.Log(fmt.Sprintf("find %d", id))
	return r.data[id]
}

// --- registration ---

func main() {
	// 1. interface → concrete via New (singleton by default)
	di.Register[Logger](func(b di.Builder[Logger]) {
		b.New(func() (Logger, error) { return &StdLogger{}, nil })
	})

	// 2. pre-built instance
	di.Register[Repo](func(b di.Builder[Repo]) {
		b.Instance(&MemRepo{
			data: map[int]string{1: "Alice", 2: "Bob"},
			log:  di.Resolve[Logger](),
		})
	})

	// 3. named variants — resolved by token
	di.Register[Logger](func(b di.Builder[Logger]) {
		b.Named("stderr", func() (Logger, error) { return &StdLogger{}, nil }).
			Scope(di.ScopeTransient)
	})

	// 4. resolution
	log := di.Resolve[Logger]()
	log.Log("ready")

	repo := di.Resolve[Repo]()
	fmt.Println(repo.Find(1))

	stderr := di.ResolveNamed[Logger]("stderr")
	stderr.Log("via named")

	// 5. safe resolution — optional dependency
	if cache, ok := di.TryResolve[Repo](); ok {
		fmt.Println(cache.Find(2))
	}
}
