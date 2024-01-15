package log

import (
	"fmt"

	"github.com/jwalton/gchalk"
)

var (
	prefixERR = gchalk.BrightRed("[ERR] ")
	prefixINF = gchalk.BrightBlue("[INF] ")
	prefixTX  = gchalk.BrightGreen("[<--] ")
	prefixRX  = gchalk.BrightYellow("[-->] ")
	prefixCMD = gchalk.BrightBlue("[•••] ")
	prefixAPI = gchalk.BrightMagenta("[API] ")
	prefixDLY = gchalk.BrightCyan("[ 💤] ")
	prefixMSM = gchalk.BrightRed("[MSM] ")
)

func ERR(msg ...any) {
	fmt.Println(gchalk.BrightRed(prefixERR), msg)
}

func INF(msg ...any) {
	fmt.Println(prefixINF, msg)
}

func API(msg ...any) {
	fmt.Println(prefixAPI, msg)
}

func MSM(msg ...any) {
	fmt.Println(prefixMSM, "[returning mismatch]", msg)
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
