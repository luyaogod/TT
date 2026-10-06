package drawio

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tt/internal/drawio"
	"tt/internal/testkit"
)

// TestSkillListsEveryShape 钉住技能里的组件清单与形状源一致。
//
// 为什么值得一条：技能里那张表是**手抄**的（读者要一眼看全，不能让它去跑命令），
// 而形状源在 Go 那边。加了控件忘了同步技能，错的不是编译，是**执行者照着技能画图时
// 找不到那个控件** —— 那种错没有任何东西会响。
//
// 只查一个方向（形状源有的，技能里都得提到）。反方向不查：技能里合法地出现一堆
// 非控件名的反引号词（`label` `field` `datainfo` …），拿它们去比库只会误报。
func TestSkillListsEveryShape(t *testing.T) {
	root := testkit.RepoRoot(t)
	p := filepath.Join(root, "skills", "tt-drawio", "SKILL.md")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("读不到技能 %s：%v（技能是对外契约，缺了就是仓库坏了）", p, err)
	}
	skill := string(b)

	src, err := drawio.Load()
	if err != nil {
		t.Fatalf("加载形状源失败：%v", err)
	}
	for _, c := range src.Catalogs {
		for _, s := range c.Shapes {
			if !strings.Contains(skill, "`"+s.ID+"`") {
				t.Errorf("技能里没提到控件 `%s`（库 %s）—— 形状源加了控件就要同步这张表", s.ID, c.Name)
			}
		}
	}
}
