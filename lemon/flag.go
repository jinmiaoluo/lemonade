package lemon

import (
	"flag"
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/mitchellh/go-homedir"
	"github.com/monochromegane/conflag"
)

// parse os.Args []string
func (c *CLI) FlagParse(args []string, skip bool) error {

	style, err := c.getCommandType(args)
	if err != nil {
		return err
	}
	// if lemonade used ad subcommand
	// the subcommand arguments will be delete
	// and shift 1 argument from right to left
	// the last one will be duplicated
	// so we remove the last argument
	if style == SUBCOMMAND {
		args = args[:len(args)-1]
	}

	return c.parse(args, skip)
}

// figure out what command name is using
// lemonade could be used as a alias or
// used as a simple command with subcommand
// reture comamnd type and error
func (c *CLI) getCommandType(args []string) (s CommandStyle, err error) {
	s = ALIAS
	// return if lemonade is used as a alias
	switch {
	case regexp.MustCompile(`/?xdg-open$`).MatchString(args[0]): // if use lemonade as a alias. what alias it is
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

	// if subcommand dont match any case just print error with usage string
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

// args 参数中有不包含文件名的 arguments 构成的字符串数组
// 解析 args []string. 返回错误
// 将 arguments 通过 flag 解析到 CLI 结构体的实例里面
func (c *CLI) parse(args []string, skip bool) error {
	//init a FlagSet, whose structure is:
	//type FlagSet struct {
	//	// Usage is the function called when an error occurs while parsing flags.
	//	// The field is a function (not a method) that may be changed to point to
	//	// a custom error handler. What happens after Usage is called depends
	//	// on the ErrorHandling setting; for the command line, this defaults
	//	// to ExitOnError, which exits the program after calling Usage.
	//	Usage func()
	//
	//	name          string
	//	parsed        bool
	//	actual        map[string]*Flag
	//	formal        map[string]*Flag
	//	args          []string // arguments after flags
	//	errorHandling ErrorHandling
	//	output        io.Writer // nil means stderr; use out() accessor
	//}
	// parse flag with default value at the same time
	flags := c.flags()

	// 加载目录中的配置文件
	// 通过配置文件初始化flag
	confPath, err := homedir.Expand("~/.config/lemonade.toml")
	if err == nil && !skip {
		// 如果配置文件存在的话
		// 解析配置文件中对应的配置项
		// 从 配置文件中解析出对应的 arguments []string
		// if success. use it as default values
		if confArgs, err := conflag.ArgsFrom(confPath); err == nil {
			flags.Parse(confArgs)
		}
	}

	// parse args from command line
	// arguments from CLI will have more higher priority
	var arg string
	// If there are some arguments is passed from CLI. Parse it
	err = flags.Parse(args[1:])
	if err != nil {
		return err
	}

	// PASTE and SERVER do not have args. so just return nil if type satisfied
	if c.Type == PASTE || c.Type == SERVER {
		return nil
	}

	// figure out whether non-flag exists or not
	for 0 < flags.NArg() {
		// arg is the first non-flag
		// arg will be used as c.DataSource
		// flags.Arg(i) will return FlagSet.args[i]
		arg = flags.Arg(0)
		// continue parse the last non-flag arguments
		err := flags.Parse(flags.Args()[1:])
		if err != nil {
			return err
		}

	}

	// 如果命令行调用了 help
	if c.Help {
		return nil
	}

	// use first non-flag as c.DataSource
	if arg != "" {
		c.DataSource = arg
	} else {
		b, err := ioutil.ReadAll(c.In)
		if err != nil {
			return err
		}
		// if non-flag is not set use CLI.In as DataSource
		c.DataSource = string(b)
	}

	return nil
}
