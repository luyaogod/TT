package devdata

// 长跑任务的状态机：单飞（同一 (env, kind) 只允许一个）+ 可取消 + 进度可见。
//
// 为什么在领域层而不是 HTTP 层：CLI 与 Web 都要"同环境同类任务不许并发"这条语义，
// 而进程内只有一个 Manager —— 放在 HTTP 层意味着 CLI 路径没有这层保护（两次并发
// `tt gen mirror` 会同时往同一个临时目录写）。
//
// 注意 Web 进程与 CLI 进程是**两个进程**，各有自己的 Manager：跨进程的单飞靠
// "临时目录 + 原子改名"兜底（同一个目标目录的并发 rename 只会有一个成功）。

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Progress 一次进度回报（字段与 plan §5.4 的载荷对齐）。
type Progress struct {
	// Phase connect | table | md l | fetch | index | write | done
	Phase string
	// Message 给界面的一句话（"拉取 DZFK_T"、"下载 code_a05.template"）。
	Message string
	// Done/Total 当前阶段已完成/总数（Total=0 表示"不知道总数"）。
	Done  int
	Total int
	// CurrentFile 当前处理的表名/文件名。
	CurrentFile string
}

// ProgressFunc 进度回调（可能被并发调用，实现需自行加锁）。
type ProgressFunc func(Progress)

// progressReporter 把 nil 回调变成一个可直接调用的函数。
func progressReporter(f ProgressFunc) func(Progress) {
	if f == nil {
		return func(Progress) {}
	}
	return func(p Progress) { f(p) }
}

// Job 一次（或最近一次）任务的状态。前端轮询它画进度条。
type Job struct {
	Running     bool   `json:"running"`
	Kind        string `json:"kind"` // mirror | fetch
	Env         string `json:"env"`
	Prog        string `json:"prog,omitempty"`
	Phase       string `json:"phase"`
	Message     string `json:"message"`
	Done        int    `json:"done"`
	Total       int    `json:"total"`
	CurrentFile string `json:"currentFile,omitempty"`
	StartedAt   string `json:"startedAt,omitempty"`
	FinishedAt  string `json:"finishedAt,omitempty"`
	Elapsed     string `json:"elapsed,omitempty"`
	// Error 失败原因（空 = 没失败）。Cancelled 与 Error 互斥：取消不是失败。
	Error     string `json:"error,omitempty"`
	Cancelled bool   `json:"cancelled,omitempty"`
	// Result 成功时的结果（*MirrorResult / *FetchResult），前端按 kind 解析。
	Result any `json:"result,omitempty"`
}

// Key 单飞键。
func jobKey(kind, env, prog string) string {
	if kind == KindFetch {
		// fetch 按**程序**单飞：同一环境可以同时拉两个程序（写的是两个目录），
		// 但不能同时拉同一个程序两次。
		return kind + "\x00" + env + "\x00" + prog
	}
	return kind + "\x00" + env
}

// Manager 进程内的任务登记处。
type Manager struct {
	mu   sync.Mutex
	jobs map[string]*jobState
	// now 供测试替换时间源（生产 = time.Now）
	now func() time.Time
}

type jobState struct {
	job    Job
	cancel context.CancelFunc
	start  time.Time
}

// NewManager 造一个 Manager。
func NewManager() *Manager {
	return &Manager{jobs: map[string]*jobState{}, now: time.Now}
}

// Start 启动一个任务。run 在后台跑；返回启动瞬间的快照。
//
// 单飞语义：同一键有任务在跑 → ErrBusy（HTTP 映射 409）。**不排队**：排队会让
// "点了没反应"与"在跑"混在一起，而用户要的是立刻知道"已经有一个在跑"。
func (m *Manager) Start(kind, env, prog string, run func(ctx context.Context, report ProgressFunc) (any, error)) (Job, error) {
	key := jobKey(kind, env, prog)
	m.mu.Lock()
	if st, ok := m.jobs[key]; ok && st.job.Running {
		m.mu.Unlock()
		return st.job, ErrBusy
	}
	ctx, cancel := context.WithCancel(context.Background())
	now := m.now()
	st := &jobState{
		job: Job{
			Running:   true,
			Kind:      kind,
			Env:       env,
			Prog:      prog,
			Phase:     "start",
			Message:   "准备中…",
			StartedAt: now.Format(time.RFC3339),
		},
		cancel: cancel,
		start:  now,
	}
	m.jobs[key] = st
	snapshot := st.job
	m.mu.Unlock()

	go func() {
		report := func(p Progress) {
			m.mu.Lock()
			defer m.mu.Unlock()
			st := m.jobs[key]
			if st == nil {
				return
			}
			if p.Phase != "" {
				st.job.Phase = p.Phase
			}
			if p.Message != "" {
				st.job.Message = p.Message
			}
			if p.Total > 0 {
				st.job.Total = p.Total
			}
			st.job.Done = p.Done
			if p.CurrentFile != "" {
				st.job.CurrentFile = p.CurrentFile
			}
			st.job.Elapsed = m.now().Sub(st.start).Round(time.Second).String()
		}
		res, err := run(ctx, report)

		m.mu.Lock()
		defer m.mu.Unlock()
		st := m.jobs[key]
		if st == nil {
			return
		}
		st.job.Running = false
		st.job.FinishedAt = m.now().Format(time.RFC3339)
		st.job.Elapsed = m.now().Sub(st.start).Round(10 * time.Millisecond).String()
		st.cancel = nil
		switch {
		case err == nil:
			st.job.Phase = "done"
			st.job.Done = st.job.Total
			st.job.Result = res
			if st.job.Message == "" {
				st.job.Message = "完成"
			}
		case errors.Is(err, context.Canceled):
			// 用户取消 ≠ 失败（与 dict 的镜像/同步同一套口径）
			st.job.Phase = "cancelled"
			st.job.Cancelled = true
			st.job.Message = "已停止；未完成的材料未落盘"
		default:
			st.job.Phase = "error"
			st.job.Error = err.Error()
			st.job.Message = "失败：" + err.Error()
		}
	}()
	return snapshot, nil
}

// Stop 取消某键正在跑的任务。没有在跑 → ErrNoJob（HTTP 映射 409）。
func (m *Manager) Stop(kind, env, prog string) error {
	key := jobKey(kind, env, prog)
	m.mu.Lock()
	st, ok := m.jobs[key]
	if !ok || !st.job.Running || st.cancel == nil {
		m.mu.Unlock()
		return ErrNoJob
	}
	cancel := st.cancel
	st.job.Message = "正在停止…"
	m.mu.Unlock()
	cancel()
	return nil
}

// Snapshot 取某键的任务快照（没跑过返回零值 Job）。
func (m *Manager) Snapshot(kind, env, prog string) (Job, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	st, ok := m.jobs[jobKey(kind, env, prog)]
	if !ok {
		return Job{}, false
	}
	return st.job, true
}

// Running 某环境当前在跑的任务（status 一次给出，前端不用逐个问）。
func (m *Manager) Running(env string) []Job {
	var out []Job
	for _, j := range m.JobsFor(env) {
		if j.Running {
			out = append(out, j)
		}
	}
	return out
}

// JobsFor 某环境相关的任务快照：**跑着的在前，其次是最近结束的**（同一键只留最近一次）。
//
// 为什么要带"已结束"的：界面靠 status 里这一项判断"这次拉取是成功、取消还是失败"，
// 并据此弹提示。只回 running 的话，任务一结束它就从列表里消失，用户永远看不到结果
// （实测：点停止之后界面什么也没说）。
func (m *Manager) JobsFor(env string) []Job {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []Job
	for _, st := range m.jobs {
		if st.job.Env != env {
			continue
		}
		out = append(out, st.job)
	}
	// 顺序稳定：运行中优先，其次 kind，再次 prog，最后按开始时间倒序
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Running != out[j].Running {
			return out[i].Running
		}
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		if out[i].Prog != out[j].Prog {
			return out[i].Prog < out[j].Prog
		}
		return out[i].StartedAt > out[j].StartedAt
	})
	return out
}

// All 全部任务快照（测试与排障用）。
func (m *Manager) All() []Job {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Job, 0, len(m.jobs))
	for _, st := range m.jobs {
		out = append(out, st.job)
	}
	return out
}

// Wait 等某键的任务结束（CLI 前台跑时用；测试也用）。
//
// 不用轮询：任务结束时把 Running 置 false，这里按 20ms 间隔看它 —— 比引入 channel
// 简单，而 CLI 本来就不是性能敏感路径。
func (m *Manager) Wait(kind, env, prog string) Job {
	for {
		j, ok := m.Snapshot(kind, env, prog)
		if !ok {
			return Job{}
		}
		if !j.Running {
			return j
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// Describe 给 CLI/日志用的一行描述。
func (j Job) Describe() string {
	s := fmt.Sprintf("%s %s", j.Kind, j.Env)
	if j.Prog != "" {
		s += "/" + j.Prog
	}
	if j.Running {
		return s + " 运行中"
	}
	if j.Cancelled {
		return s + " 已停止"
	}
	if j.Error != "" {
		return s + " 失败"
	}
	return s + " 完成"
}
