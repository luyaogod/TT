package fep

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/fep"
	"tt/internal/winproc"
)



var recActiveCmd = &cobra.Command{
	Use:   "active",
	Short: "列出正在进行的录制（登记里还活着的那些：含 GDC 是否还在、权限是否对得上）",
	Args:  noArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		reg, _, err := registry()
		if err != nil {
			return err
		}
		list, err := reg.List()
		if err != nil {
			return fail(exitFailed, "%v", err)
		}
		if len(list) == 0 {
			return fail(exitNoSession, "没有登记中的会话（%s 不存在）", reg.Path)
		}
		// 多条之后 `ps` 是**列表**：每条都带自己的 tag，`close --tag` 就是照它点名。
		views := make([]map[string]any, 0, len(list))
		for _, s := range list {
			alive := winproc.Alive(s.GDCPID)
			v := map[string]any{"session": s, "tag": s.Tag, "gdcAlive": alive, "registry": reg.Path}
			if e, eerr := fep.CheckElevation(s.GDCPID); eerr == nil {
				v["elevation"] = e
				if !e.OK {
					v["warning"] = e.Advice
				}
			}
			views = append(views, v)
		}
		return printJSONOrText(map[string]any{"sessions": views, "registry": reg.Path}, func() {
			for i, s := range list {
				if i > 0 {
					fmt.Println()
				}
				fmt.Printf("会话    : %s (环境 %s / 区域 %s / 启动路径 %s / %s)\n", s.Prog, s.Env, s.Zone, s.Path, entText(s.Ent))
				fmt.Printf("  登记ID: %s\n", s.Tag)
				if s.Ref != "" {
					// 调试线那边的地址（C2.0 的 `环境#序号`）：两条线从此对得上，不用人脑补。
					fmt.Printf("  会话  : %s\n", s.Ref)
				}
				fmt.Printf("  GDC   : pid %d 端口 %d  %s\n", s.GDCPID, s.GDCPort, aliveText(winproc.Alive(s.GDCPID)))
				fmt.Printf("  代理  : 监听 %d → 127.0.0.1:%d（注入口 %d）\n", s.ProxyPort, s.GDCPort, s.InjectPort)
				fmt.Printf("  起始  : %s（启动者 pid %d）\n", s.Started.Format(time.RFC3339), s.OwnerPID)
				if e, eerr := fep.CheckElevation(s.GDCPID); eerr == nil && !e.OK {
					fmt.Printf("  ⚠ %s\n", e.Advice)
				}
			}
		})
	},
}

// closeRegisteredSession 收掉一份登记（tag 空 = 唯一那条；多条则要求点名）。
//
// 抽出来是因为**两个入口共用它**：`tt recording stop <录制ID>`（旧）与 `tt recording stop`（新）。
// 收尾这件事里有一串容易漏的细节（先请常驻进程自退、只杀登记的 pid、调试侧按 ref 收口、
// 最后才摘登记），复制两份必然漂 —— 而漂的那一次会留下一个跑着的代理或一条僵着的调试会话。
func closeRegisteredSession(cmd *cobra.Command, tag string) error {
		reg, _, err := registry()
		if err != nil {
			return err
		}
		list, err := reg.List()
		if err != nil {
			return fail(exitFailed, "%v", err)
		}
		if len(list) == 0 {
			return fail(exitNoSession, "没有登记中的会话（%s 不存在）", reg.Path)
		}
		// 多条登记必须点名：与调试线同一条规矩 —— 多候选不自己挑。
		// 从前这里恒为"唯一那条"，多会话之后"随手收一条"就是收别人的界面。
		tag = strings.TrimSpace(tag)
		if tag == "" {
			if len(list) > 1 {
				tags := make([]string, 0, len(list))
				for _, s := range list {
					tags = append(tags, s.Tag)
				}
				return fail(exitNoSession, "登记里有多条会话，请用 --tag 点名要收哪一条：%s", strings.Join(tags, "、"))
			}
			tag = list[0].Tag
		}
		s, lerr := reg.LoadTag(tag)
		if errors.Is(lerr, fep.ErrSessionNotFound) {
			return fail(exitNoSession, "没有 tag 为 %q 的登记（用 tt recording active 看全部）", tag)
		}
		if lerr != nil {
			return fail(exitFailed, "%v", lerr)
		}
		alive := winproc.Alive(s.GDCPID)
		// **先结束调试，再收代理** —— 顺序反了会把挂靠的那条会话一起带走。
		//
		// 代理是作业画面唯一的出口：先杀代理，作业那一侧立刻断连，会话以 exit 收场，
		// 于是"收工后宿主保留、下一轮免登录"这条承诺当场失效（真机实测：挂在 `regA` 上录一轮，
		// `recording stop` 之后 `session ls` 空）。
		// 先 quit 则是"结束本轮、宿主留着"，代理随后关掉时已经没有活作业可连累了。
		preQuit := s.Path == fep.PathDebug && strings.TrimSpace(s.Ref) != ""
		if preQuit {
			if cl, cerr := debugClient(); cerr == nil {
				if b, rerr := cl.Resolve(s.Ref); rerr == nil {
					if qerr := cl.Quit(b.ID); qerr != nil {
						fmt.Fprintf(cmd.ErrOrStderr(), "· 结束调试时失败（继续收代理）：%v\n", qerr)
					}
				}
			}
		}
		stopped, serr := reg.StopTag(tag)
		if serr != nil {
			return fail(exitFailed, "%v", serr)
		}
		// 老登记没有 `ref` 时只能按别名找回那条会话 —— 那一步必须放在收代理**之后**
		// （别名是服务端的寻址事实，代理死了也还在），且只在没做过 preQuit 时走。
		if stopped.Path == fep.PathDebug && !preQuit {
			target := recSessionAlias(stopped.Tag)
			if cl, cerr := debugClient(); cerr != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "· 连不上调试面，跳过调试侧收尾：%v\n", cerr)
			} else if target != "" {
				if b, rerr := cl.Resolve(target); rerr != nil {
					fmt.Fprintln(cmd.ErrOrStderr(), "· 调试会话已不在（无需收）")
				} else if qerr := cl.Quit(b.ID); qerr != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "· 调试会话没收干净（可手工 tt debug quit）：%v\n", qerr)
				}
			}
		}
		return printJSONOrText(map[string]any{"stopped": stopped.Prog, "tag": stopped.Tag, "gdcPid": stopped.GDCPID, "wasAlive": alive, "registryRemoved": true}, func() {
			if alive {
				fmt.Printf("已收会话 %s（%s）：GDC pid %d 已结束，登记已清\n", stopped.Prog, stopped.Tag, stopped.GDCPID)
			} else {
				fmt.Printf("已收会话 %s（%s）：GDC pid %d 早已不在，登记已清\n", stopped.Prog, stopped.Tag, stopped.GDCPID)
			}
		})
}


func aliveText(alive bool) string {
	if alive {
		return "运行中"
	}
	return "已不在（登记过期）"
}

// entText 这个会话跑在哪个企业上。
//
// 没给 --ent 时要**如实说**"沿用登录 shell 的"，不能留空或写个 `-`：
// 留空读起来像"这个会话没有企业"，而"我没指定"与"没有"是两件事。
func entText(ent string) string {
	if ent == "" {
		return "ENT 沿用登录 shell 的 $TOPENT"
	}
	return "ENT " + ent
}
