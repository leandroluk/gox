// internal/ast/presence.go
package ast

type Presence uint8

const (
	Missing Presence = iota
	Null
	Present
)

func (presence Presence) String() string {
	switch presence {
	case Missing:
		return "missing"
	case Null:
		return "null"
	case Present:
		return "present"
	default:
		return "unknown"
	}
}
