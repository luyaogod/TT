package fep

// LaunchPath 是"这一次作业是怎么起来的"。
//
// 为什么它必须是一等身份（实测，见 .agents/plan/tt-agent-field-notes.md §9）：
// AUI 的对象编号由运行期按声明顺序分配，所以**编号是启动历史的函数** ——
// 同一套界面换一条启动路径，对象集几乎一样（1556 / 1558），而共同条目的编号
// 一条都不相同。录制与回放因此必须走**同一条**路径，而录制件必须记住自己是从
// 哪条路径来的：这是闸门唯一能提前发现"这份录像不是这条路径录的"的依据。
type LaunchPath string

const (
	// PathRunnerRR 是 T100 的标准启动器：`r.r <作业>` → `fglrun $AZZ/42r/azzp000.42r <作业>`。
	// 界面正常、但**不可调试**（azzp000 不带 -d）。
	PathRunnerRR LaunchPath = "r.r"
	// PathDebug 是我们自己的调试线：`fglrun -d <作业>.42r <启动参数>`，调试器是 fgldb。
	PathDebug LaunchPath = "debug"
	// PathRunnerRD 是 T100 自带的调试启动器：`r.d <作业>` → gendbg → **fgldeb**（T100 的调试器）。
	//
	// 它**已从命令面拿掉**（`launchCommand` 会拒）：调试器是 T100 的而不是我们的，
	// 画面上会多出一个 T100 原生调试器界面，而我们的断点台账 / fgldb 那套用不上。
	// 这个常量留着是为了**识别**以前录的 r.d 录制件（它们的身份里写着 r.d）——
	// 实测那种件里的作业界面与调试线的编号表相同，所以照样能在调试线上回放。
	PathRunnerRD LaunchPath = "r.d"
	// PathNone 是"没起作业"：只起了隔离实例与代理（--no-job 排障，或给调试线当画面接收端）。
	// 它没有自己的启动历史，所以闸门**不拿它做路径比对**，只报"路径无从而知"。
	PathNone LaunchPath = "none"
)

// Known 报告这条路径是不是一条真正起过作业的路径。
//
// `r.d` 也算：命令面已经不给它了，但**以前录的件**里写着它，闸门得认这个身份。
func (p LaunchPath) Known() bool {
	switch p {
	case PathRunnerRR, PathDebug, PathRunnerRD:
		return true
	}
	return false
}

// Label 给人读的名字（录制件身份里就写它）。
func (p LaunchPath) Label() string {
	switch p {
	case PathRunnerRR:
		return "r.r（T100 标准启动器，不可调试）"
	case PathDebug:
		return "debug（fglrun -d + fgldb）"
	case PathRunnerRD:
		return "r.d（gendbg + fgldeb）"
	case PathNone:
		return "none（只起了实例与代理，没起作业）"
	case "":
		return "（未知）"
	}
	return string(p)
}
