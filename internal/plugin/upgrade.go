package plugin

// 手动升级与回滚（P6-02～P6-06）。
//
// 这一份代码的全部重量都在**顺序**上：先装未激活版本 → 停旧服务 → 原子切换 → 起新服务
// 并健康检查 → 重算技能 → 收尾清理。任何一步失败都必须让"当前可用的那个版本"完好如初。
//
// 为什么每一步都这么排：
//   - 校验、解包、登记都发生在停旧服务**之前**：网络/哈希/清单问题不该先打断一个正在
//     工作的插件；
//   - 切换是注册表里的一次 WithLock 写入：要么整条记录换成新版本，要么还是旧版本；
//   - 切换之后**任何**失败（服务起不来、健康检查不过、UI 入口丢了、技能刷不进平台目录）
//     都走同一条回滚路径：切回旧版本、把旧服务重新拉起来、重算技能、报清原因；
//   - 旧实例的令牌与会话在切换前后都要撤（P6-05）：新旧版本绝不能共用凭证；
//   - 清理只保留"当前 + 上一个"两个程序版本：回滚要用的那个不能提前删，持久数据永远不动。

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"tt/internal/fetch"
	"tt/internal/winproc"
)

// UpdateRuntime 升级/回滚期间对**运行态**的操作。
//
// 为什么是注入点：常驻服务必须是长命进程（tt serve）的子进程，而 `tt plugin update` 是
// 一次性进程。tt serve 里它由 supervisor + 令牌表 + 连接服务实现；CLI 里没有，于是
// Manager 走"先确认没有别的进程正在跑这个插件的服务"这条路（见 ensureServiceIdle）。
type UpdateRuntime interface {
	// StopService 停掉某插件的常驻服务（没有在跑时返回 nil）。
	StopService(pluginID string) error
	// StartService 起服务并做健康检查（插件没有 service 贡献时返回 nil）。
	StartService(rec *Record) error
	// InvalidateInstances 撤销该插件的全部实例令牌并关闭它建立的会话。
	InvalidateInstances(pluginID string)
}

// UpdateOptions 一次手动升级的参数。
//
// 没有"自动检查/自动下载"这回事：索引地址、要不要升、升到哪一版都由用户这一次命令决定
// （plan §3：不自动更新插件）。
type UpdateOptions struct {
	// IndexURL 发布索引地址；空则用该插件上次安装时记下的索引地址。
	IndexURL string
	// ID 只升这一个插件；空 = 全部已装插件。
	ID string
	// To 固定版本（必须出现在索引里，且与本机 Host API 兼容）。
	To string
	// DryRun 只打印计划，一个字节都不写。
	DryRun bool
	// HTTP 测试缝；nil = 缺省客户端。
	HTTP *http.Client
	// AllowInsecureHTTP 只给测试用（httptest 的普通 HTTP 服务端）。
	AllowInsecureHTTP bool
	// Retries 下载重试次数；0 = 缺省，负数 = 不重试（测试用）。
	Retries int
	// Agent User-Agent。
	Agent string
}

// UpdateResult 一个插件的升级结果（批量里单个失败不拖垮整批）。
type UpdateResult struct {
	ID   string `json:"id"`
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
	// Updated 真的换了版本（成功）。
	Updated bool `json:"updated"`
	// UpToDate 没有可升的版本（不是错误）。
	UpToDate bool `json:"upToDate,omitempty"`
	// DryRun 这条结果只是计划。
	DryRun bool `json:"dryRun,omitempty"`
	// RolledBack 切换之后失败、已经自动退回旧版本。
	RolledBack bool     `json:"rolledBack,omitempty"`
	Reason     string   `json:"reason,omitempty"`
	Notes      []string `json:"notes,omitempty"`
	// Error 这个插件的失败原因（整批仍会继续跑下一个）。
	Error string `json:"error,omitempty"`
}

// Update 手动升级插件。
//
// 返回的 error 只表示**整批没法开始**（索引读不到、索引不合法、没有索引地址）；
// 单个插件的失败留在 UpdateResult.Error 里，调用方据此决定退出码，但不会因此漏掉其它插件。
func (m *Manager) Update(ctx context.Context, opt UpdateOptions) ([]UpdateResult, error) {
	reg, err := m.Store.Load()
	if err != nil {
		return nil, err
	}
	targets, err := m.updateTargets(reg, opt.ID)
	if err != nil {
		return nil, err
	}
	indexURL, err := m.indexURL(reg, opt)
	if err != nil {
		return nil, err
	}
	if err := fetch.RequireHTTPS(indexURL, opt.AllowInsecureHTTP); err != nil {
		return nil, fmt.Errorf("发布索引地址不可用: %w", err)
	}
	body, err := fetch.GetBytes(ctx, opt.HTTP, indexURL, opt.Agent, 0)
	if err != nil {
		return nil, fmt.Errorf("读发布索引失败（%s）: %w", indexURL, err)
	}
	idx, err := ParseReleaseIndex(body, opt.AllowInsecureHTTP)
	if err != nil {
		return nil, err
	}

	// 索引只读一次：批量升级时不该为每个插件各拉一份（还可能拉到不同版本的索引）。
	out := make([]UpdateResult, 0, len(targets))
	for _, rec := range targets {
		out = append(out, m.updateOne(ctx, rec, idx, indexURL, opt))
	}
	return out, nil
}

// updateTargets 决定这一批升哪些插件。
func (m *Manager) updateTargets(reg *Registry, id string) ([]*Record, error) {
	id = strings.TrimSpace(id)
	if id != "" {
		rec := reg.Find(id)
		if rec == nil {
			return nil, fmt.Errorf("没有插件 %q（tt plugin list --all 看已装的）", id)
		}
		return []*Record{rec}, nil
	}
	recs := append([]*Record(nil), reg.Plugins...)
	if len(recs) == 0 {
		return nil, errors.New("没有已安装的插件")
	}
	return recs, nil
}

// indexURL 决定这次用哪份发布索引。
//
// 显式 --index 优先；否则用"上次装这个插件时记下的索引地址"。**没有内置默认地址**：
// TT 不替用户选插件市场，装插件时用的是哪个源，升级就回哪个源（见 docs 13-release-index）。
func (m *Manager) indexURL(reg *Registry, opt UpdateOptions) (string, error) {
	if s := strings.TrimSpace(opt.IndexURL); s != "" {
		return s, nil
	}
	if strings.TrimSpace(opt.ID) == "" {
		return "", errors.New("必须给 --index <发布索引地址>（批量升级没有可以推断的源）")
	}
	rec := reg.Find(strings.TrimSpace(opt.ID))
	if rec == nil {
		return "", fmt.Errorf("没有插件 %q", opt.ID)
	}
	if rec.Source.Kind == SourceURL && strings.TrimSpace(rec.Source.Ref) != "" {
		return rec.Source.Ref, nil
	}
	return "", fmt.Errorf("插件 %s 是本机来源（%s）安装的，没有记录发布索引地址：加 --index <地址>",
		rec.ID, rec.Source)
}

// updateOne 升级一个插件。任何失败都只影响这一个插件。
func (m *Manager) updateOne(ctx context.Context, rec *Record, idx *ReleaseIndex, indexURL string, opt UpdateOptions) UpdateResult {
	res := UpdateResult{ID: rec.ID, From: rec.ActiveVersion, DryRun: opt.DryRun}
	ip := idx.Find(rec.ID)
	if ip == nil {
		res.UpToDate = true
		res.Reason = "发布索引里没有这个插件（索引可以只列部分插件）"
		return res
	}
	target, err := ip.Pick(rec.ActiveVersion, m.hostAPIVersion(), opt.To)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	if target == nil {
		res.UpToDate = true
		res.Reason = fmt.Sprintf("没有比 %s 更新、且与 Host API %s 兼容的版本", rec.ActiveVersion, m.hostAPIVersion())
		return res
	}
	res.To = target.Version
	asset, ok := target.AssetFor()
	if !ok {
		res.Error = fmt.Sprintf("%s %s 没有 %s 平台的资产（索引里只有: %s）",
			rec.ID, target.Version, AssetKey(), strings.Join(assetKeys(target), " / "))
		return res
	}

	// 前置条件放在动任何东西之前：这条不过就什么都还没变。
	if err := m.ensureServiceIdle(rec); err != nil {
		res.Error = err.Error()
		return res
	}
	if opt.DryRun {
		res.Notes = append(res.Notes,
			fmt.Sprintf("将下载 %s（%s，sha256 %s…）", asset.URL, fetch.HumanBytes(asset.Size), shortHash(asset.SHA256)),
			fmt.Sprintf("将解包校验后装成未激活版本 %s，再停服务、原子切换 %s → %s、起服务并健康检查、重算技能",
				target.Version, rec.ActiveVersion, target.Version),
			fmt.Sprintf("成功后清理旧版本，只保留 %s 与 %s", target.Version, rec.ActiveVersion),
		)
		return res
	}

	source := Source{Kind: SourceURL, Ref: indexURL, SHA256: asset.SHA256}
	if rec.HasVersion(target.Version) {
		// 上一轮升级失败留下的版本（或 --to 点了一个已装版本）：不再下载，直接校验后切换。
		// 这条路径让"失败后重试"不必重下几十 MB。
		dir := m.Store.VersionDir(rec.ID, target.Version)
		man, merr := LoadManifest(dir)
		if merr != nil {
			res.Error = fmt.Sprintf("已装版本 %s 的清单读不出来: %v（先 tt plugin uninstall 再装）", target.Version, merr)
			return res
		}
		if verr := m.validateForActivation(man, dir, rec); verr != nil {
			res.Error = fmt.Sprintf("已装版本 %s 不能启用: %v", target.Version, verr)
			return res
		}
		res.Notes = append(res.Notes, "该版本已在本机：直接切换，不重新下载")
	} else if ierr := m.installNewVersion(ctx, rec, target, asset, source, opt); ierr != nil {
		res.Error = ierr.Error()
		return res
	}

	// 2) 停旧服务。失败 = 一个新字节都没生效，旧版本好好的。
	if serr := m.stopService(rec.ID); serr != nil {
		res.Error = fmt.Sprintf("停止旧版本服务失败（活跃版本仍是 %s）: %v", rec.ActiveVersion, serr)
		return res
	}
	// 3) 撤旧实例令牌 + 关它的会话：新旧版本不能共用凭证（P6-05）。
	m.invalidateInstances(rec.ID)

	// 4) 原子切换。
	from := rec.ActiveVersion
	dir := m.Store.VersionDir(rec.ID, target.Version)
	man, merr := LoadManifest(dir)
	if merr != nil {
		res.Error = fmt.Sprintf("读新版本清单失败（活跃版本仍是 %s）: %v", from, merr)
		m.restartService(rec.ID)
		return res
	}
	if serr := m.switchActive(rec.ID, man, target.Version, from); serr != nil {
		res.Error = fmt.Sprintf("切换活跃版本失败（活跃版本仍是 %s）: %v", from, serr)
		m.restartService(rec.ID)
		return res
	}
	res.Notes = append(res.Notes, fmt.Sprintf("活跃版本 %s → %s", from, target.Version))

	// 5) 切换后的验收：UI 入口、服务启动 + 健康检查、技能重算。任一失败都回滚。
	if serr := m.postSwitchVerify(rec.ID, target.Version, from); serr != nil {
		if rbErr := m.revertTo(rec.ID, from, target.Version, serr); rbErr != nil {
			res.Error = fmt.Sprintf("升级到 %s 失败（%v），自动回滚也失败（%v）：先 tt plugin disable %s 再 tt plugin rollback %s",
				target.Version, serr, rbErr, rec.ID, rec.ID)
			return res
		}
		res.RolledBack = true
		res.Error = fmt.Sprintf("升级到 %s 失败，已回滚到 %s：%v", target.Version, from, serr)
		return res
	}

	res.Updated = true
	// 6) 收尾：只保留"当前 + 上一个"两个程序版本（持久数据从不在这里动）。
	if removed := m.pruneVersions(rec.ID, target.Version, from); len(removed) > 0 {
		res.Notes = append(res.Notes, "清理旧程序版本: "+strings.Join(removed, "、"))
	}
	return res
}

// ensureServiceIdle 确认"没有别的 tt 进程正在跑这个插件的服务"，或者本进程能把它停掉。
//
// 为什么要拦：命令行升级只能改注册表与载荷。若 tt serve 正跑着旧版本的服务，切换活跃版本
// 会留下「页面/CLI 是新版、常驻服务是旧版」的混用状态（旧令牌、旧会话都还活着）—— 那正是
// P6-05 要消灭的东西。tt serve 里 Runtime 非空，可以真停真起，这条检查因此不拦它。
func (m *Manager) ensureServiceIdle(rec *Record) error {
	if m.Runtime != nil {
		return nil
	}
	if rec.Contributions.Service == nil {
		return nil
	}
	st, err := m.Store.ReadInstanceState(rec.ID)
	if err != nil || st == nil {
		return nil
	}
	if st.State != "running" || st.PID <= 0 {
		return nil // 已经收工的状态：不是"正在跑"，不该挡住升级
	}
	// PID 存活不够：复用的 PID 会把"别人的进程"当成插件服务，从而永远拒绝升级（P3-14）。
	if reason := st.StaleReason(winproc.Alive(st.PID), processIdentity(st.PID)); reason != "" {
		return nil // 遗留/复用的运行状态：诊断项（tt plugin doctor）会指出来，不该挡住升级
	}
	return fmt.Errorf("插件 %s 的常驻服务正由另一个 tt 进程（tt serve）运行（pid=%d）：命令行升级会留下"+
		"「页面新、服务旧」的混用状态。先 tt serve --stop 再升级，升级完重新启动 tt serve", rec.ID, st.PID)
}

// installNewVersion 下载 → 校验 → 解包 → 校验清单 → 原子移入版本目录 → 登记为**未激活**。
//
// 这一段全程不碰活跃版本：装不上时旧版本连"被停一下"都没有发生。
func (m *Manager) installNewVersion(ctx context.Context, rec *Record, v *IndexVersion, asset IndexAsset, source Source, opt UpdateOptions) error {
	if err := m.Store.EnsureDirs(); err != nil {
		return err
	}
	if err := fetch.RequireHTTPS(asset.URL, opt.AllowInsecureHTTP); err != nil {
		return fmt.Errorf("资产地址不可用: %w", err)
	}
	// 下载物落在 plugins/cache（明确可清）；解包在 stage-* 临时目录里做。
	// 下载物留着：同样一份 zip 重跑时可以按摘要直接跳过。
	dl := filepath.Join(m.Store.CacheDir(), fmt.Sprintf("dl-%s-%s.zip", rec.ID, v.Version))
	if err := fetch.Download(ctx, opt.HTTP, fetch.Request{
		URL: asset.URL, Name: rec.ID + " " + v.Version, Dest: dl,
		SHA256: asset.SHA256, Size: asset.Size,
		Agent: opt.Agent, Via: "手动 tt plugin update", Retries: opt.Retries,
	}); err != nil {
		return fmt.Errorf("下载 %s %s 失败: %w", rec.ID, v.Version, err)
	}

	stage, err := m.newStageDir()
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage) // 移入版本目录后这里是空操作

	if err := ExtractZip(dl, stage, DefaultZipLimits); err != nil {
		return fmt.Errorf("解包 %s %s 失败: %w", rec.ID, v.Version, err)
	}
	root, err := manifestRoot(stage)
	if err != nil {
		return err
	}
	man, err := LoadManifest(root)
	if err != nil {
		return err
	}
	// 索引说是什么版本，包里的清单就必须是什么版本。不一致 = 有人拿别的包冒充这个版本。
	if man.ID != rec.ID {
		return fmt.Errorf("包内清单的 id 是 %q，与索引里的 %q 不一致", man.ID, rec.ID)
	}
	if man.Version != v.Version {
		return fmt.Errorf("包内清单的版本是 %s，与索引声明的 %s 不一致", man.Version, v.Version)
	}
	if err := m.validateForActivation(man, root, rec); err != nil {
		return err
	}
	target, err := m.placeVersion(man, root)
	if err != nil {
		return err
	}
	if err := m.addVersionInactive(man, source); err != nil {
		// 登记失败 → 撤掉刚移进来的载荷，不留"目录里有、注册表里没有"的孤儿。
		_ = os.RemoveAll(target)
		return fmt.Errorf("登记新版本失败（已撤掉载荷）: %w", err)
	}
	return nil
}

// validateForActivation 校验"这个版本能不能成为活跃版本"。
//
// 与安装时的校验同源，另外显式检查前端入口：升级后 UI 入口缺失是明确的回滚触发条件
// （P6-04），让它在这里失败比让用户看到一片白屏好。
func (m *Manager) validateForActivation(man *Manifest, dir string, rec *Record) error {
	reg, err := m.Store.Load()
	if err != nil {
		return err
	}
	vopt := ValidateOptions{
		HostAPIVersion:     m.hostAPIVersion(),
		ReservedNamespaces: m.ReservedNamespaces,
		OtherNamespaces:    reg.NamespacesInUse(rec.ID),
		OtherRoutes:        reg.RoutesInUse(rec.ID),
	}
	if err := ValidateManifest(man, dir, vopt); err != nil {
		return err
	}
	return checkUIEntry(man, dir)
}

// checkUIEntry 前端入口必须真实存在（没有前端贡献时什么也不查）。
func checkUIEntry(man *Manifest, dir string) error {
	if man.Contributions.Frontend == nil {
		return nil
	}
	rel, err := man.ResolveFrontendEntry()
	if err != nil {
		return err
	}
	st, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel)))
	if err != nil || !st.Mode().IsRegular() {
		return fmt.Errorf("前端入口 %q 不在版本目录里（页面会打不开）", rel)
	}
	return nil
}

// switchActive 把活跃版本换成 target（一次 WithLock 写入：要么换成功，要么还是旧的）。
func (m *Manager) switchActive(id string, man *Manifest, target, previous string) error {
	return m.Store.WithLock(func(reg *Registry) error {
		rec := reg.Find(id)
		if rec == nil {
			return fmt.Errorf("插件 %q 在切换过程中消失（注册表被并发修改）", id)
		}
		if !rec.HasVersion(target) {
			return fmt.Errorf("版本 %s 没有登记在注册表里（安装步骤没有完成）", target)
		}
		rec.ActiveVersion = target
		rec.PreviousVersion = previous
		if man != nil {
			rec.Name = man.Name
			rec.Contributions = man.Contributions
			rec.Permissions = append([]string(nil), man.Permissions...)
		}
		// 来源跟着活跃版本走，但**本机来源不覆盖已记录的发布源**：
		// 回滚到"当初从目录装的 0.1.0"之后，这个插件仍然是从那个发布索引来的，
		// 下次 `tt plugin update` 不该因为回滚了一次就要求重新给 --index。
		for _, iv := range rec.Versions {
			if iv.Version != target {
				continue
			}
			if iv.Source.Kind == SourceURL || rec.Source.Kind == "" {
				rec.Source = iv.Source
			}
			break
		}
		rec.UpdatedAt = m.Store.Stamp()
		rec.LastError = ""
		return nil
	})
}

// postSwitchVerify 切换之后的验收：UI 入口 → 新服务启动 + 健康检查 → 技能重算。
func (m *Manager) postSwitchVerify(id, version, from string) error {
	reg, err := m.Store.Load()
	if err != nil {
		return err
	}
	rec := reg.Find(id)
	if rec == nil {
		return fmt.Errorf("插件 %q 在切换后消失", id)
	}
	dir := m.Store.VersionDir(id, version)
	man, err := LoadManifest(dir)
	if err != nil {
		return err
	}
	if err := checkUIEntry(man, dir); err != nil {
		return err
	}
	if err := m.startService(rec); err != nil {
		return fmt.Errorf("新版本服务启动失败: %w", err)
	}
	if err := m.syncSkills(id, fmt.Sprintf("update %s %s→%s", id, from, version)); err != nil {
		return fmt.Errorf("技能刷新失败: %w", err)
	}
	return nil
}

// revertTo 把活跃版本退回到 from，并把 from 的服务与技能恢复到可用状态。
//
// cause 留在 LastError 上：回滚成功不等于"什么都没发生"，用户要能看到为什么退回来了。
func (m *Manager) revertTo(id, from, failed string, cause error) error {
	dir := m.Store.VersionDir(id, from)
	man, err := LoadManifest(dir)
	if err != nil {
		return fmt.Errorf("回滚目标版本 %s 的清单读不出来: %w", from, err)
	}
	if serr := m.stopService(id); serr != nil {
		return fmt.Errorf("停止失败的新版本服务出错: %w", serr)
	}
	m.invalidateInstances(id)
	// 回滚后再没有"上一个版本"：PreviousVersion 指向一个刚被证明不可用的版本只会误导人。
	if serr := m.switchActive(id, man, from, ""); serr != nil {
		return fmt.Errorf("切回 %s 失败: %w", from, serr)
	}
	m.recordError(id, fmt.Sprintf("升级到 %s 失败并已回滚到 %s: %v", failed, from, cause))

	reg, lerr := m.Store.Load()
	if lerr != nil {
		return lerr
	}
	rec := reg.Find(id)
	if rec != nil {
		if serr := m.startService(rec); serr != nil {
			return fmt.Errorf("回滚后 %s 的服务启动失败: %w", from, serr)
		}
	}
	if serr := m.syncSkills(id, fmt.Sprintf("rollback %s %s", id, from)); serr != nil {
		return fmt.Errorf("回滚后技能刷新失败: %w", serr)
	}
	return nil
}

// RollbackOptions 回滚参数。
type RollbackOptions struct {
	// DryRun 只说明会切到哪个版本，不改任何东西。
	DryRun bool
}

// Rollback 把插件切回上一个版本（`tt plugin rollback <id>`）。
//
// 与升级共用同一套安全动作：先校验目标版本、停旧服务、撤凭证、原子切换、起服务并健康
// 检查、重算技能。切换之后失败会把活跃版本切回原处并报告 —— 回滚本身也不能把插件弄坏。
func (m *Manager) Rollback(id string, opt RollbackOptions) (*UpdateResult, error) {
	reg, err := m.Store.Load()
	if err != nil {
		return nil, err
	}
	rec := reg.Find(id)
	if rec == nil {
		return nil, fmt.Errorf("没有插件 %q", id)
	}
	prev := strings.TrimSpace(rec.PreviousVersion)
	if prev == "" {
		return nil, fmt.Errorf("插件 %s 没有可回滚的版本（只有成功切换过版本的插件才记着上一个版本）", id)
	}
	if !rec.HasVersion(prev) {
		return nil, fmt.Errorf("记录的上一个版本 %s 不在已装版本里（可能已被清理）：tt plugin info %s 看现有版本", prev, id)
	}
	dir := m.Store.VersionDir(id, prev)
	man, merr := LoadManifest(dir)
	if merr != nil {
		return nil, fmt.Errorf("上一个版本 %s 的清单读不出来: %w", prev, merr)
	}
	// 那个版本当初能用，不代表现在还能用：Host API 可能已经升过大版本。
	if verr := m.validateForActivation(man, dir, rec); verr != nil {
		return nil, fmt.Errorf("上一个版本 %s 现在不能启用: %w", prev, verr)
	}
	if serr := m.ensureServiceIdle(rec); serr != nil {
		return nil, serr
	}

	res := &UpdateResult{ID: id, From: rec.ActiveVersion, To: prev, DryRun: opt.DryRun}
	if opt.DryRun {
		res.Notes = append(res.Notes, fmt.Sprintf("将把活跃版本从 %s 切回 %s", rec.ActiveVersion, prev))
		return res, nil
	}

	from := rec.ActiveVersion
	if serr := m.stopService(id); serr != nil {
		return nil, fmt.Errorf("停止当前版本服务失败（活跃版本仍是 %s）: %w", from, serr)
	}
	m.invalidateInstances(id)
	if serr := m.switchActive(id, man, prev, from); serr != nil {
		m.restartService(id)
		return nil, fmt.Errorf("切换活跃版本失败（活跃版本仍是 %s）: %w", from, serr)
	}
	// 切换后的验收：失败就用同一条路径切回去。
	if serr := m.postSwitchVerify(id, prev, from); serr != nil {
		if rbErr := m.revertTo(id, from, prev, serr); rbErr != nil {
			return nil, fmt.Errorf("回滚到 %s 失败（%v），自动退回 %s 也失败（%v）：先 tt plugin disable %s 再 tt plugin enable %s",
				prev, serr, from, rbErr, id, id)
		}
		return nil, fmt.Errorf("回滚到 %s 失败，已退回 %s：%v", prev, from, serr)
	}
	res.Updated = true
	res.Notes = append(res.Notes, fmt.Sprintf("活跃版本 %s → %s（%s 的程序留在本机，可再次切换）", from, prev, from))
	return res, nil
}

// pruneVersions 清理不再需要的旧程序版本，保留 keep 里的那些。
//
// 三条纪律：
//  1. 只删**程序**版本目录；持久数据在 plugins/data/<id>，这里一个字节都不碰；
//  2. 只删注册表里记过的版本（不按目录名猜）；
//  3. 删目录失败就继续在注册表里记着它 —— 否则磁盘上有、注册表里没有，永远清不掉。
func (m *Manager) pruneVersions(id string, keep ...string) []string {
	keepSet := map[string]bool{}
	for _, k := range keep {
		if k != "" {
			keepSet[k] = true
		}
	}
	reg, err := m.Store.Load()
	if err != nil {
		return nil
	}
	rec := reg.Find(id)
	if rec == nil {
		return nil
	}
	drop := map[string]bool{}
	var removed []string
	for _, iv := range rec.Versions {
		if keepSet[iv.Version] {
			continue
		}
		if dir := m.Store.VersionDir(id, iv.Version); dir != "" {
			if err := removeAllRetry(dir); err != nil {
				continue
			}
		}
		drop[iv.Version] = true
		removed = append(removed, iv.Version)
	}
	if len(drop) == 0 {
		return nil
	}
	_ = m.Store.WithLock(func(reg *Registry) error {
		r := reg.Find(id)
		if r == nil {
			return nil
		}
		kept := r.Versions[:0]
		for _, iv := range r.Versions {
			if drop[iv.Version] {
				continue
			}
			kept = append(kept, iv)
		}
		r.Versions = kept
		return nil
	})
	return removed
}

// stopService / startService / restartService / invalidateInstances 是 Runtime 的
// 空安全包装：CLI 进程没有运行态，这些调用必须变成"什么也不做"而不是 panic。
func (m *Manager) stopService(id string) error {
	if m.Runtime == nil {
		return nil
	}
	return m.Runtime.StopService(id)
}

func (m *Manager) startService(rec *Record) error {
	if m.Runtime == nil || rec == nil {
		return nil
	}
	return m.Runtime.StartService(rec)
}

// restartService 尽力把"当前活跃版本"的服务重新拉起来（中止路径用，错误只记不报）。
func (m *Manager) restartService(id string) {
	if m.Runtime == nil {
		return
	}
	reg, err := m.Store.Load()
	if err != nil {
		return
	}
	rec := reg.Find(id)
	if rec == nil || !rec.Enabled {
		return
	}
	_ = m.Runtime.StartService(rec)
}

func (m *Manager) invalidateInstances(id string) {
	if m.Runtime != nil {
		m.Runtime.InvalidateInstances(id)
	}
}

// shortHash 摘要前 12 位（展示用）。
func shortHash(s string) string {
	if len(s) > 12 {
		return s[:12]
	}
	return s
}
