//go:build !gui

package main

func startComputer(computer *Computer, debugger Debugger) {
	run(computer, debugger)
}
