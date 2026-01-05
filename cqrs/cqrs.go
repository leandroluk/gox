package cqrs

import "context"

// --- Queries ---

var queryRegistry = newRegistry("query")

type IQueryHandler[TQuery any, TResult any] interface {
	Handle(ctx context.Context, query TQuery) (TResult, error)
}

func RegisterQueryHandler[TQuery any, TResult any, THandler IQueryHandler[TQuery, TResult]](factoryFN any) {
	register[TQuery, TResult, THandler](queryRegistry, factoryFN)
}

func ExecuteQuery[TResult any](ctx context.Context, query any) (TResult, error) {
	return execute[TResult](queryRegistry, ctx, query)
}

// --- Commands ---

var commandRegistry = newRegistry("command")

type ICommandHandler[TCommand any, TResult any] interface {
	Handle(ctx context.Context, command TCommand) (TResult, error)
}

func RegisterCommandHandler[TCommand any, TResult any, THandler ICommandHandler[TCommand, TResult]](factoryFN any) {
	register[TCommand, TResult, THandler](commandRegistry, factoryFN)
}

func ExecuteCommand[TResult any](ctx context.Context, command any) (TResult, error) {
	return execute[TResult](commandRegistry, ctx, command)
}
