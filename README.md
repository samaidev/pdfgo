# pdfgo

`samaidev` 旗下的 PDF 生成库，从 `samaidev/samoffice` 的 `pkg/pdfgo` 中抽取并独立封装。

设计与 `docgo` / `xlsgo` / `pptgo` 保持一致的链式 API 风格，让开发者一行代码生成结构化的 PDF 文档。

## 特性

- 多级标题（1–6 级）
- 段落（粗体 / 斜体 / 下划线 / 删除线 / 颜色 / 字体 / 字号 / 缩进 / 行距 / 对齐）
- 富文本 Run（段落内多段不同样式的文字）
- 有序 / 无序列表
- 表格（自动列宽 + 表头）
- 图片（路径或字节）
- 代码块（等宽字体 + 灰底）
- 分隔线 / 分页符
- 页面设置（A4 / Letter、页边距）
- **中文字体**：内置 Noto Serif SC，无需外部字体文件即可生成含中文的 PDF
- 元信息（标题 / 作者 / 主题 / 创建者）与页脚页码

## 安装

```bash
go get github.com/samaidev/pdfgo
```

## 快速开始

```go
package main

import (
	"log"

	"github.com/samaidev/pdfgo"
)

func main() {
	doc := pdfgo.New()
	doc.SetTitle("报告标题").SetAuthor("SamAI").SetPageNumber(true)
	doc.SetPageSize(pdfgo.PageSizeA4)
	doc.SetMargins(pdfgo.Margins{Top: 60, Bottom: 60, Left: 60, Right: 60})

	doc.AddHeading("第一章 概述", 1)
	p := doc.AddParagraph("")
	p.AddRun("pdfgo 提供与 docgo 一致的链式 API。").Bold(true)
	doc.AddParagraph("正文内容，支持中文。").SetIndent(24).SetLineSpacing(1.6)

	doc.AddHeading("核心 API", 2)
	doc.AddList([]string{
		"AddHeading(text, level) — 多级标题",
		"AddParagraph(text).AddRun(text) — 富文本段落",
		"AddTable(rows) — 表格（含表头）",
		"AddImage(path, w, h) — 图片",
		"AddCodeBlock(lang, code) — 代码块",
	})

	doc.AddTable([][]string{
		{"元素", "API"},
		{"标题", "AddHeading"},
		{"段落", "AddParagraph"},
	})

	if err := doc.Save("report.pdf"); err != nil {
		log.Fatal(err)
	}
}
```

## 输出

- `doc.Save(path)` 写入文件
- `doc.Bytes() ([]byte, error)` 返回字节，便于在内存中处理（如 HTTP 响应）
- `doc.WriteTo(w io.Writer)` 直接流式写出

## 读取（提取文本）

`pdfgo` 同样提供 **PDF 只读** 能力，与 `.doc` / `.xls` / `.ppt` 旧格式的策略一致：
只能读取，不支持编辑。底层使用 `github.com/ledongthuc/pdf`（纯 Go、轻量）。

```go
r, _ := pdfgo.OpenReader("report.pdf")
// 或：pdfgo.OpenReaderBytes(b) / pdfgo.OpenReaderFrom(readerAt, size)
fmt.Println("页数:", r.PageCount())

// 整篇文本
text, _ := r.Text()
fmt.Println(text)

// 按页
for _, p := range r.Pages() {
    t, _ := p.Text()
    fmt.Printf("第 %d 页: %s\n", p.Index()+1, t)
}
```

> 注意：文本提取依赖 PDF 内部结构。对于由本库（`Document.Bytes()`）或常见工具
> 生成的 PDF，提取效果良好；个别结构特殊的 PDF 可能无法提取，此时 `Text()` 会
> 返回错误而非崩溃（底层 panic 已被兜底）。

## 依赖

- 生成：`github.com/signintech/gopdf`
- 读取：`github.com/ledongthuc/pdf`

## 许可证

MIT
