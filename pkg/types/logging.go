package types

type Logger interface {
	Println(a ...any)
	RePrintf(format string, a ...any)
	IsNop() bool
}
