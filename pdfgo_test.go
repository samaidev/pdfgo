package pdfgo

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

// makePNG 生成一个最小的合法 PNG 字节，供 AddImageBytes 测试使用。
func makePNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {
			img.Set(x, y, color.RGBA{uint8(x * 32), uint8(y * 32), 128, 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png encode: %v", err)
	}
	return buf.Bytes()
}

func TestNewAndBytes(t *testing.T) {
	doc := New()
	doc.SetTitle("测试文档").SetAuthor("pdfgo").SetPageNumber(true)
	doc.AddHeading("标题一", 1)
	doc.AddParagraph("中文正文内容，验证 UTF-8 编码。").AddRun("").Bold(true)
	doc.AddList([]string{"项目甲", "项目乙"})
	doc.AddOrderedList([]string{"第一步", "第二步"})
	doc.AddTable([][]string{
		{"名称", "说明"},
		{"pdfgo", "PDF 生成库"},
	})
	doc.AddCodeBlock("go", "fmt.Println(\"hi\")")
	doc.AddImageBytes(makePNG(t), 40, 40)
	doc.AddDivider()
	doc.AddPageBreak()

	data, err := doc.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF")) {
		t.Fatalf("输出不是合法 PDF（缺少 %%PDF 头），前 16 字节: % x", data[:min(16, len(data))])
	}
	if !bytes.Contains(data, []byte("%%EOF")) {
		t.Fatalf("PDF 缺少 %%EOF 结束标记")
	}
	if len(data) < 1000 {
		t.Fatalf("生成的 PDF 过小 (%d 字节)，可能渲染异常", len(data))
	}
}

func TestElements(t *testing.T) {
	doc := New()
	doc.AddHeading("H", 1)
	doc.AddParagraph("P")
	doc.AddList([]string{"a"})
	els := doc.Elements()
	want := map[string]int{"heading": 0, "paragraph": 0, "list": 0}
	for _, e := range els {
		switch e.elementType() {
		case "heading", "paragraph", "list":
			want[e.elementType()]++
		}
	}
	if want["heading"] != 1 || want["paragraph"] != 1 || want["list"] != 1 {
		t.Fatalf("元素类型计数不符: %+v", want)
	}
}

func TestChineseText(t *testing.T) {
	doc := New()
	doc.AddParagraph("中文字体渲染测试：中华人民共和国")
	data, err := doc.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	// gopdf 以字体字形方式嵌入中文，文本不会以明文形式出现，
	// 这里只验证能成功生成且体积合理（含 Noto Serif SC 字形子集）。
	if len(data) < 2000 {
		t.Fatalf("中文 PDF 体积异常 (%d 字节)，可能字体未嵌入", len(data))
	}
}

func TestTableAndCode(t *testing.T) {
	doc := New()
	tbl := doc.AddTable([][]string{{"A", "B"}, {"1", "2"}})
	if !tbl.HasHeader() {
		t.Fatalf("AddTable 应默认含表头")
	}
	cb := doc.AddCodeBlock("python", "print(1)")
	if !strings.Contains(cb.code, "print") {
		t.Fatalf("CodeBlock 内容丢失")
	}
	if _, err := doc.Bytes(); err != nil {
		t.Fatalf("Bytes: %v", err)
	}
}

func TestSaveAndMetadata(t *testing.T) {
	doc := New()
	doc.SetTitle("元信息测试").SetAuthor("pdfgo").SetSubject("subject")
	if doc.Title() != "元信息测试" || doc.Author() != "pdfgo" {
		t.Fatalf("元数据 getter 不符")
	}
	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF")) {
		t.Fatalf("WriteTo 输出不是合法 PDF")
	}
}
