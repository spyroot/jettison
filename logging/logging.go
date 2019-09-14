package logging

import (
	"github.com/fatih/color"
	"log"
	"runtime"
)

/**
  Output critical message to console in red
*/
func CriticalMessage(msg ...string) {
	color.Set(color.FgRed)
	log.Println(msg)
	color.Unset()
}

/**
  Used to print additional meta information about error.
  function name and line.
*/
func ErrorLogging(err error) (b bool) {
	if err != nil {
		_, fn, line, _ := runtime.Caller(1)
		log.Printf("[error] %s:%d %v", fn, line, err)
		b = true
	}
	return
}

/**
  Used to print additional meta information about error.
  function name and line.
*/
func FmtErrorLogging(err error) (b bool) {
	if err != nil {
		pc, fn, line, _ := runtime.Caller(1)
		log.Printf("[error] in %s[%s:%d] %v", runtime.FuncForPC(pc).Name(), fn, line, err)
		b = true
	}
	return
}
