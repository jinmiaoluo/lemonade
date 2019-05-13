package lemon

import (
	"flag"
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/mitchellh/go-homedir"
	"github.com/monochromegane/conflag"
)

// parse args from commandline
func (c *CLI) FlagParse(args []string, skip bool) error {

	style, err := c.getCommandType(args)
	if err != nil {
		return err
	}
	if style == SUBCOMMAND {
		args = args[:len(args)-1]
	}

	return c.parse(args, skip)
}

// figure out what command name is using
// lemonade could be used as a alias or
// used as a simple command with subcommand
func (c *CLI) getCommandType(args []string) (s CommandStyle, err error) {
	s = ALIAS
	switch {
	case regexp.MustCompile(`/?xdg-open$`).MatchString(args[0]): // 判断是否使用了 alias. 并判断alias的类型.
		c.Type = OPEN
		return
	case regexp.MustCompile(`/?pbpaste$`).MatchString(args[0]):
		c.Type = PASTE
		return
	case regexp.MustCompile(`/?pbcopy$`).MatchString(args[0]):
		c.Type = COPY
		return
	}

	del := func(i int) {
		// delete specified index argument
		copy(args[i+1:], args[i+2:])
		args[len(args)-1] = ""
	}

	s = SUBCOMMAND
	// delete subcommand when we know what subcommand it is
	for i, v := range args[1:] {
		switch v {
		case "open":
			c.Type = OPEN
			del(i)
			return
		case "paste":
			c.Type = PASTE
			del(i)
			return
		case "copy":
			c.Type = COPY
			del(i)
			return
		case "server":
			c.Type = SERVER
			del(i)
			return
		}
	}

	return s, fmt.Errorf("Unknown SubCommand\n\n" + Usage)
}

// create a FlagSet
// create flag with default value and description
// bind flag with struct CLI's field
func (c *CLI) flags() *flag.FlagSet {
	flags := flag.NewFlagSet("lemonade", flag.ContinueOnError)
	flags.IntVar(&c.Port, "port", 2489, "TCP port number")
	flags.StringVar(&c.Allow, "allow", "0.0.0.0/0,::/0", "Allow IP range")
	flags.StringVar(&c.Host, "host", "localhost", "Destination host name.")
	flags.BoolVar(&c.Help, "help", false, "Show this message")
	flags.BoolVar(&c.TransLoopback, "trans-loopback", true, "Translate loopback address")
	flags.BoolVar(&c.TransLocalfile, "trans-localfile", true, "Translate local file")
	flags.StringVar(&c.LineEnding, "line-ending", "", "Convert Line Endings (CR/CRLF)")
	flags.BoolVar(&c.NoFallbackMessages, "no-fallback-messages", false, "Do not show fallback messages")
	flags.IntVar(&c.LogLevel, "log-level", 1, "Log level")
	return flags
}

// parse args without subcommand
func (c *CLI) parse(args []string, skip bool) error {
	flags := c.flags() //初始化一个 FlagSet

	confPath, err := homedir.Expand("~/.config/lemonade.toml")
	if err == nil && !skip {
		// 如果配置文件存在的话
		// 解析配置文件中对应的配置项
		// 从 配置文件中解析出对应的 args 字符串
		if confArgs, err := conflag.ArgsFrom(confPath); err == nil {
			flags.Parse(confArgs)
		}
	}

	// if do not have config file
	// just parse args from command line
	var arg string
	err = flags.Parse(args[1:])
	if err != nil {
		return err
	}

	// PASTE and SERVER do not have args
	if c.Type == PASTE || c.Type == SERVER {
		return nil
	}

	// NArg is the number of arguments remaining after flags have been processed.
	for 0 < flags.NArg() {
		/* Arg returns the i'th argument. Arg(0) is the first remaining argument after
		flags have been processed. Arg returns an empty string if the requested element
		does not exist.
		*/
		arg = flags.Arg(0)
		err := flags.Parse(flags.Args()[1:])
		if err != nil {
			return err
		}

	}

	if c.Help {
		return nil
	}

	if arg != "" {
		c.DataSource = arg
	} else {
		b, err := ioutil.ReadAll(c.In)
		if err != nil {
			return err
		}
		c.DataSource = string(b)
	}

	return nil
}
