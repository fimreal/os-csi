## zap 日志模块的封装

旧版本每次都要 new 创建 logger 很麻烦，新版本参考大佬的用法，在包内 init 完，暴露需要的日志级别，用法接近 logrus。

缺点是每次引用都需要重新配置 logger，使用默认配置比较方便。

参考：https://github.com/blessmylovexy/log


#### 特点

- 时间格式使用 `grok` 可匹配的 `TIMESTAMP_ISO8601`， 也可以自定义时间格式，或者传入空字符串关闭时间显示。
- 默认不输出 json 格式，可使用 `ezap.EnableJSONFormat()` 启用 json 日志格式。
- 日志输出支持三种格式，默认的字符串输出，templ 输出，以及 kv 格式输出。

#### 用法🌰

```go
package main

import log "github.com/fimreal/goutils/ezap"

func main() {
	log.Info("info")
	log.Infof("%s", "info 2")
	log.Infow("birthday", "name", "bb", "time", "1996-11-06")

	log.Debug("debug 1")
	log.SetProjectName("[unified backend]")
	log.SetLevel("debug")
	log.Debug("debug 2")

	log.EnableJSONFormat()
	log.Info("info 3")
	log.Error("Undefined error occurred")
	// Output:
	// 2021-11-17T18:08:47.406+0800	INFO	info
	// 2021-11-17T18:08:47.406+0800	INFO	info 2
	// 2021-11-17T18:08:47.406+0800	INFO	birthday	{"name": "bb", "time": "1996-11-06"}
	// 2021-11-17T18:08:47.406+0800	DEBUG	debug 2
	// {"lv":"info","ts":"2021-11-17T18:08:47.406+0800","msg":"info 3"}
	// {"lv":"error","ts":"2021-11-17T18:08:47.406+0800","msg":"Undefined error occurred"}
}

```
