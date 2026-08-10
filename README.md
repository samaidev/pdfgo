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

## 关于「读取」

本库当前聚焦 **PDF 生成**。原 `samoffice` 仓库中并不存在 PDF 读取/解析实现，因此
`samaidev/pdfgo` 暂不提供读功能。若后续需要读取（提取文本、合并、拆分等），
可基于 `pdfcpu` 或 `unipdf` 等扩展，欢迎提 issue 讨论。

## 许可证

MIT
