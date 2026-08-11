package pdfgo

import (
	"strings"
	"testing"
)

// TestReadRealPDF 回归：ledongthuc/pdf 为惰性解析，OpenReader 不能提前关闭
// 文件，否则真实（FlateDecode 压缩）PDF 会被读成空文本。此测试确保对真实
// 生成的 PDF 能提取出非空正文。
//
// 运行前需准备真实 PDF，例如：
//   curl -sL -o /tmp/realdocs/real.pdf \
//     https://raw.githubusercontent.com/mozilla/pdf.js/master/web/compressed.tracemonkey-pldi-09.pdf
func TestReadRealPDF(t *testing.T) {
	const path = "/tmp/realdocs/real.pdf"
	r, err := OpenReader(path)
	if err != nil {
		t.Skipf("跳过：未找到真实 PDF %s（%v）", path, err)
	}
	defer r.Close()

	txt, err := r.Text()
	if err != nil {
		t.Fatalf("Text() 错误: %v", err)
	}
	if strings.TrimSpace(txt) == "" {
		t.Fatalf("真实 PDF 提取文本为空（惰性解析/文件提前关闭的回归）")
	}
	if !strings.Contains(txt, "Just-in-Time") && !strings.Contains(txt, "Trace") {
		t.Fatalf("真实 PDF 未提取出预期正文，preview=%q", txt[:min(80, len(txt))])
	}
}
