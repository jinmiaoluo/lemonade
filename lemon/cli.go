package lemon

import "io"

type CommandType int

// Commands type const to represent command type
const (
	OPEN CommandType = iota + 1
	COPY
	PASTE
	SERVER
)

// return 0 10 20 30 as special err codes
const (
	Success        = 0
	FlagParseError = iota + 10
	RPCError
	Help
)

type CommandStyle int

const (
	ALIAS CommandStyle = iota + 1
	SUBCOMMAND
)

// cli struct is used for store info pased from cli
type CLI struct {
	In       io.Reader // todo
	Out, Err io.Writer // todo

	Type       CommandType
	DataSource string

	// options
	Port           int // server port
	Allow          string
	Host           string // server address
	TransLoopback  bool
	TransLocalfile bool
	LineEnding     string
	LogLevel       int

	Help bool

	NoFallbackMessages bool
}
