package threading

import (
	"runtime/debug"

	"github.com/rs/zerolog/log"
)

func GoSafe(fn func(), panicMsg string, panicHandler func(r any)) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error().
					Str("panic_msg", panicMsg).
					Str("stack", string(debug.Stack())).
					Msgf("recovered from panic: %+v", r)

				if panicHandler != nil {
					panicHandler(r)
				}
			}
		}()
		fn()
	}()
}
