package fep

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDropMetaIfNoRecording 钉住台账 D28 的判据：
//
// 录制失败时缓存目录里会留一份 `<录制ID>.meta.json`，而**根本没有录制件**。
// 那份身份描述的是"一份不存在的录像" —— 留着它会让人以为"有一份录制可以回放"
// （真机冒烟里正是这样被误导的：日志说"录制身份已落"，缓存里却没有对应录制件）。
//
// 判据只有一条硬事实：**这一份有没有流量文件**。没有 ⇒ 该删；有 ⇒ 一个字节都不许动。
func TestDropMetaIfNoRecording(t *testing.T) {
	const tag = "fep-20261018-120000"

	t.Run("没有流量文件 ⇒ 该删", func(t *testing.T) {
		dir := t.TempDir()
		// 只有身份文件，没有 gdc2rt / rt2gdc 那些流量文件
		if err := os.WriteFile(filepath.Join(dir, tag+".meta.json"), []byte(`{"id":"x"}`), 0o644); err != nil {
			t.Fatal(err)
		}
		if !dropMetaIfNoRecording(dir, tag) {
			t.Fatal("没有录制件时该判「该删」—— 否则缓存里会留一份描述空录像的身份")
		}
	})

	t.Run("有流量文件 ⇒ 不许删", func(t *testing.T) {
		dir := t.TempDir()
		// 造一份真实出现过的流量文件形态（L1 实测：fep-…-1-gdc2rt.frames.jsonl）
		if err := os.WriteFile(filepath.Join(dir, tag+"-1-gdc2rt.frames.jsonl"), []byte("{}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if dropMetaIfNoRecording(dir, tag) {
			t.Fatal("有录制件时**绝不能**删身份 —— 那是唯一能说明这份录像是什么的来源")
		}
	})

	t.Run("别人的录制件不算这一份", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "fep-20261018-130000-1-gdc2rt.frames.jsonl"), []byte("{}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if !dropMetaIfNoRecording(dir, tag) {
			t.Fatal("只按这一份的 tag 找文件，别人的录制件不能当成本轮的")
		}
	})
}
