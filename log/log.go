package log

import (
	"fmt"

	"github.com/jwalton/gchalk"
)

var prefixERR = gchalk.BrightRed("[ERR] ")
var prefixINF = gchalk.BrightBlue("[INF] ")
var prefixTX = gchalk.BrightGreen("[<--] ")
var prefixRX = gchalk.BrightYellow("[-->] ")
var prefixCMD = gchalk.BrightBlue("[•••] ")
var prefixAPI = gchalk.BrightMagenta("[API] ")
var prefixDLY = gchalk.BrightCyan("[ 💤] ")

func ERR(msg ...any) {
	fmt.Println(gchalk.BrightRed(prefixERR), msg)
}

func INF(msg ...any) {
	fmt.Println(prefixINF, msg)
}

func API(msg ...any) {
	fmt.Println(prefixAPI, msg)
}

func DLY(msg ...any) {
	fmt.Println(prefixDLY, msg)
}

func printWithPrefix(prefix, str string, hex []byte) {
	h := fmt.Sprintf("[% x]", hex)
	fmt.Println(prefix, gchalk.BrightWhite(str), gchalk.Gray(h))
}

func CMD(str ...any) {
	fmt.Println(prefixCMD, str)
}

func TX(str string, hex []byte) {
	printWithPrefix(prefixTX, str, hex)
}

func RX(str string, hex []byte) {
	printWithPrefix(prefixRX, str, hex)
}
