package lemon

import (
	"fmt"
	"regexp"
	"strings"
)

// declare lemonade version info
// exported variable for visiting
// use `go -ldflags` to override
var Version string

// 初始化用法的输出信息
// 存储了用法信息的变量可以在其他包内访问
var Usage = fmt.Sprintf(`Usage: lemonade [options]... SUB_COMMAND [arg]
Sub Commands:
  open [URL]                  Open URL by browser
  copy [text]                 Copy text.
  paste                       Paste text.
  server                      Start lemonade server.

Options:
  --port=2489                 TCP port number
  --line-ending               Convert Line Ending (CR/CRLF)
  --allow="0.0.0.0/0,::/0"    Allow IP Range                [Server only]
  --host="localhost"          Destination hostname          [Client only]
  --no-fallback-messages      Do not show fallback messages [Client only]
  --trans-loopback=true       Translate loopback address    [open subcommand only]
  --trans-localfile=true      Translate local file path     [open subcommand only]
  --log-level=1               Log level                     [4 = Critical, 0 = Debug]
  --help                      Show this message


Version:
  %s`, Version)

// 转换末尾行的结束符
func ConvertLineEnding(text, option string) string {
	switch option {
	case "lf", "LF":
		text = strings.Replace(text, "\r\n", "\n", -1) // 匹配所有的字符串. 将所有的 `\r\n` 替换为 `\n`
		return strings.Replace(text, "\r", "\n", -1)   // 匹配所有的字符串. 将所有的 `\r` 替换为 `\n`
	case "crlf", "CRLF":
		text = regexp.MustCompile(`\r(.)|\r$`).ReplaceAllString(text, "\r\n$1")     //构建匹配非 `\r\n` 的正则, 正则表达式必须正确, 否则报错. 通过正则表达式匹配字符串, 并进行替换. 组内的数据作为 $1 传递到新的字符串内.
		text = regexp.MustCompile(`([^\r])\n|^\n`).ReplaceAllString(text, "$1\r\n") //替换末尾的 `\n` 为 `\r\n`
		return text
	default:
		return text // 将处理过的 text 作为默认的输出
	}
}
