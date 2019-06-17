package main

import (
	"fmt"
	"os"

	log "github.com/inconshreveable/log15"

	"github.com/lemonade-command/lemonade/client"
	"github.com/lemonade-command/lemonade/lemon"
	"github.com/lemonade-command/lemonade/server"
)

var logLevelMap = map[int]log.Lvl{
	0: log.LvlDebug,
	1: log.LvlInfo,
	2: log.LvlWarn,
	3: log.LvlError,
	4: log.LvlCrit,
}

func main() {

	cli := &lemon.CLI{
		In:  os.Stdin,
		Out: os.Stdout,
		// default error message will be sent to user as stderr
		// help will be sent to user as stderr
		Err: os.Stderr,
	}
	os.Exit(Do(cli, os.Args))
}

func Do(c *lemon.CLI, args []string) int {
	logger := log.New()
	// set logger write log handler
	//const (
	//	LvlCrit Lvl = iota
	//	LvlError
	//	LvlWarn
	//	LvlInfo
	//	LvlDebug
	//)
	// only write less than special level log
	// LvlError is default level
	logger.SetHandler(log.LvlFilterHandler(log.LvlError, log.StdoutHandler))

	if err := c.FlagParse(args, false); err != nil {
		writeError(c, err)
		return lemon.FlagParseError
	}

	// override default log level
	logLevel := logLevelMap[c.LogLevel]
	logger.SetHandler(log.LvlFilterHandler(logLevel, log.StdoutHandler))

	// if user wanna help message
	// just use lemonade --help to do that
	if c.Help {
		fmt.Fprint(c.Err, lemon.Usage)
		return lemon.Help
	}

	// init a client
	lc := client.New(c, logger)
	var err error

	switch c.Type {
	case lemon.OPEN:
		logger.Debug("Opening URL")
		err = lc.Open(c.DataSource, c.TransLocalfile, c.TransLoopback)
	case lemon.COPY:
		logger.Debug("Copying text")
		err = lc.Copy(c.DataSource)
	case lemon.PASTE:
		logger.Debug("Pasting text")
		var text string
		text, err = lc.Paste()
		c.Out.Write([]byte(text))
	case lemon.SERVER:
		logger.Debug("Starting Server")
		err = server.Serve(c, logger)
	default:
		panic("Unreachable code")
	}

	if err != nil {
		writeError(c, err)
		return lemon.RPCError
	}
	return lemon.Success
}

func writeError(c *lemon.CLI, err error) {
	fmt.Fprintln(c.Err, err.Error())
}
