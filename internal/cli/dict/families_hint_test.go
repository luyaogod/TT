package dict

import (
	"testing"

	"tt/internal/dict/dbsync"
)

// TestFamilyCommandsResolve 盯着 dbsync.Families[].Commands 里的名字在命令树里真的存在。
//
// 为什么值得一条测试：这些名字今天只有两个用途 —— `tt dict db status` 的「依赖命令」列，
// 以及各命令 `--help` 末尾那句「本地数据齐不齐」。后者是靠
// dbsync.FamilyKeysForCommand(c.Name()) 反查的：**名字对不上，那一行就什么都不打**。
// 提示静默消失，而 `--help` 其余部分一切正常，没人会想到去查它。改名（r.t → rt）
// 或打错字都走这条路。
//
// 不能只断言「Find 没报错」：cobra 的 Find 找不到子命令时会**原样返回父命令**，
// 所以必须比对 Name()。
func TestFamilyCommandsResolve(t *testing.T) {
	// 防空转：Families 被清空或 Commands 全被抹掉时，下面的循环一次都不进，测试会假绿。
	const minFamilies, minCommandEntries = 10, 15

	if len(dbsync.Families) < minFamilies {
		t.Fatalf("只读到 %d 个数据族（应有至少 %d 个）—— 本测试扫的是 Families，先确认它没被清空",
			len(dbsync.Families), minFamilies)
	}
	entries := 0
	for _, f := range dbsync.Families {
		if len(f.Commands) == 0 {
			t.Errorf("数据族 %s 没有登记任何依赖命令 —— 它的表就不会出现在任何一条提示里", f.Key)
		}
		entries += len(f.Commands)
	}
	if entries < minCommandEntries {
		t.Fatalf("全部数据族加起来只有 %d 条依赖命令（应有至少 %d 条）—— 同上，先确认 Families 没被改坏",
			entries, minCommandEntries)
	}

	for _, f := range dbsync.Families {
		for _, name := range f.Commands {
			cmd, _, err := Group.Find([]string{name})
			got := "(nil)"
			if cmd != nil {
				got = cmd.Name()
			}
			if err != nil || got != name {
				t.Errorf("数据族 %s 登记的依赖命令 %q 在 tt dict 命令树里找不到（Find 得到 %q, err=%v）\n"+
					"  后果：这条命令的 --help 末尾不会显示「本地数据」那一行，且 db status 的依赖命令列会指错。\n"+
					"  改法：把 %q 改成命令的真实名字（看 Group.Commands() 里那个 Name()），或从这一族删掉它。",
					f.Key, name, got, err, name)
			}
		}
	}
}
