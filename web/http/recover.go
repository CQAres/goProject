package http

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

func Recover() HandlerFunc {
	return func(context *Context) {
		defer func() {
			if r := recover(); r != nil {
				message := fmt.Sprintf("%s", r)
				log.Printf("%s\n\n", trace(message))

				// 确保 ResponseWriter 可用
				if context.Writer != nil {
					context.Writer.Header().Set("Content-Type", "text/plain")
					context.Writer.WriteHeader(http.StatusInternalServerError)
					_, _ = fmt.Fprintf(context.Writer, "Internal Server Error")
				}
			}
		}()

		context.Next()
	}
}

func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:])

	var str strings.Builder
	str.WriteString(message + "\nTraceback: ")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}
