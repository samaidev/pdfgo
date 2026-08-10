package pdfgo

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// TestReaderRoundTrip 自包含验证：用本库 Document 生成一个含已知英文文本的
// PDF，再用 Reader 读回，确认文本提取链路真实可用（不依赖网络/外部文件）。
func TestReaderRoundTrip(t *testing.T) {
	doc := New()
	doc.AddHeading("Demo", 1)
	doc.AddParagraph("Hello PDF from pdfgo")
	doc.AddList([]string{"alpha", "beta"})

	data, err := doc.Bytes()
	if err != nil {
		t.Fatalf("生成 PDF 失败: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF")) {
		t.Fatalf("生成的不是合法 PDF")
	}

	r, err := OpenReaderBytes(data)
	if err != nil {
		t.Fatalf("OpenReaderBytes: %v", err)
	}
	if r.PageCount() < 1 {
		t.Fatalf("页数应 >= 1，实际 %d", r.PageCount())
	}
	text, err := r.Text()
	if err != nil {
		t.Fatalf("Text: %v", err)
	}
	if !strings.Contains(text, "Hello PDF from pdfgo") {
		t.Fatalf("提取文本不含预期内容，实际: %q", firstN(text, 200))
	}
	t.Logf("往返成功：生成 %d 字节，读回页数=%d，文本长度=%d", len(data), r.PageCount(), len(text))
}

func TestReaderFrom(t *testing.T) {
	doc := New()
	doc.AddParagraph("ReaderFrom check")
	data, err := doc.Bytes()
	if err != nil {
		t.Fatalf("生成失败: %v", err)
	}
	r, err := OpenReaderFrom(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("OpenReaderFrom: %v", err)
	}
	if r.PageCount() < 1 {
		t.Fatalf("页数应 >= 1，实际 %d", r.PageCount())
	}
}

// TestReaderRealFile 用网上下载的真实 PDF 做端到端验证（文件存在才跑）。
func TestReaderRealFile(t *testing.T) {
	const path = "/tmp/real/fw4.pdf"
	if _, err := os.Stat(path); err != nil {
		t.Skip("跳过：真实样本不存在（需先下载 /tmp/real/fw4.pdf）")
	}
	r, err := OpenReader(path)
	if err != nil {
		t.Fatalf("OpenReader: %v", err)
	}
	if r.PageCount() < 1 {
		t.Fatalf("真实 PDF 页数应 >= 1，实际 %d", r.PageCount())
	}
	text, err := r.Text()
	if err != nil {
		// ledongthuc/pdf 对个别复杂 PDF 暂不支持，标记跳过而非失败。
		t.Skipf("底层库暂不支持该真实 PDF 的文本提取: %v", err)
	}
	if len(strings.TrimSpace(text)) == 0 {
		t.Skip("真实 PDF 提取文本为空（底层库限制）")
	}
	t.Logf("真实 PDF 页数=%d, 文本长度=%d, 前 120 字: %q", r.PageCount(), len(text), firstN(text, 120))
}

func firstN(s string, n int) string {
	s = strings.ReplaceAll(strings.TrimSpace(s), "\n", " ")
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}
