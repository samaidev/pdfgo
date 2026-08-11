package pdfgo

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/ledongthuc/pdf"
)

// Reader 提供 PDF 的只读文本提取能力。
//
// 设计约束：PDF 为二进制排版格式，本库仅做读取（提取文本），
// 不提供保真编辑（与 .doc/.xls/.ppt 旧格式的只读策略一致）。
// 生成新 PDF 请使用本包的 Document（基于 gopdf）。
//
// 文本提取底层使用 ledongthuc/pdf（纯 Go、轻量）。对于内容流为
// FlateDecode 压缩、含嵌入字体的真实 PDF（由办公套件/浏览器生成），
// 会自动回退到 GetPlainText 路径以解压缩并解析字体编码，从而提取出文字。
//
// ledongthuc/pdf 采用惰性解析：页面内容流在调用 Page/Text 时才真正从底层
// 源读取，因此 Reader 持有底层源并在 Close 时才释放，OpenReader 不会提前关闭文件。
type Reader struct {
	r     *pdf.Reader
	pages []*Page
	src   io.Closer // 底层源（如 *os.File），需 Close 释放
}

// Page 表示 PDF 中的一页，支持按页提取文本。
type Page struct {
	reader *Reader
	num    int // 1-based，对应 ledongthuc/pdf 的页码
}

// OpenReader 打开一个 PDF 文件并返回只读 Reader。
// 返回的 Reader 持有底层文件句柄，调用方应在使用完毕后调用 Close 释放。
func OpenReader(path string) (*Reader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("pdfgo: 打开 %s 失败: %w", path, err)
	}
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("pdfgo: 读取文件信息失败: %w", err)
	}
	r, err := openFrom(f, fi.Size())
	if err != nil {
		f.Close()
		return nil, err
	}
	r.src = f
	return r, nil
}

// OpenReaderBytes 从字节切片打开 PDF。
func OpenReaderBytes(b []byte) (*Reader, error) {
	return openFrom(bytes.NewReader(b), int64(len(b)))
}

// OpenReaderFrom 从实现了 io.ReaderAt 的源（如 *os.File、*bytes.Reader）打开 PDF。
// size 为该源的字节总长度。
func OpenReaderFrom(r io.ReaderAt, size int64) (*Reader, error) {
	return openFrom(r, size)
}

func openFrom(r io.ReaderAt, size int64) (*Reader, error) {
	reader, err := pdf.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("pdfgo: 解析失败: %w", err)
	}
	count := reader.NumPage()
	pages := make([]*Page, count)
	for i := 0; i < count; i++ {
		pages[i] = &Page{reader: &Reader{r: reader}, num: i + 1}
	}
	return &Reader{r: reader, pages: pages}, nil
}

// PageCount 返回 PDF 总页数。
func (r *Reader) PageCount() int {
	return len(r.pages)
}

// Close 释放底层源（如文件句柄）。仅当 OpenReader 打开文件时需要调用；
// OpenReaderBytes/OpenReaderFrom 传入的内存源无需（也不会）被关闭。
func (r *Reader) Close() error {
	if r.src != nil {
		return r.src.Close()
	}
	return nil
}

// Pages 返回所有页的切片。
func (r *Reader) Pages() []*Page {
	return r.pages
}

// Page 返回指定页（0-based）。越界返回 nil。
func (r *Reader) Page(i int) *Page {
	if i < 0 || i >= len(r.pages) {
		return nil
	}
	return r.pages[i]
}

// Text 提取整篇 PDF 的全部文本（按页拼接，页间以空行分隔）。
func (r *Reader) Text() (string, error) {
	var buf bytes.Buffer
	for i, p := range r.pages {
		t, err := p.Text()
		if err != nil {
			return "", err
		}
		if i > 0 {
			buf.WriteString("\n\n")
		}
		buf.WriteString(t)
	}
	return buf.String(), nil
}

// Text 提取该页纯文本。
//
// 实现：遍历页面内容流中的文本片段（ledongthuc/pdf 的 Content.Text），
// 按基线 Y 坐标从上到下、同行按 X 从左到右排序后拼接；当相邻片段的 Y
// 落差超过字号一半时视为换行。直接读取片段的 .S 字段，避免 GetPlainText
// 在个别复杂 PDF 上的兼容性问题。
func (p *Page) Text() (string, error) {
	if p.reader == nil || p.reader.r == nil {
		return "", fmt.Errorf("pdfgo: 页未关联有效 Reader")
	}

	var (
		out string
		err error
	)
	// ledongthuc/pdf 在遇到个别复杂/损坏的 PDF 时会在内部 panic，
	// 这里用 recover 兜底，将其转换为普通错误，避免拖垮调用方进程。
	defer func() {
		if r := recover(); r != nil {
			if out == "" {
				err = fmt.Errorf("pdfgo: 该 PDF 结构暂不支持提取（底层解析异常: %v）", r)
			}
		}
	}()

	page := p.reader.r.Page(p.num)

	// 优先走 GetPlainText：它会在内部解压缩 FlateDecode 内容流并解析
	// 字体编码/ToUnicode CMap，对由办公套件/浏览器生成的真实 PDF 兼容性最好。
	fontMap := map[string]*pdf.Font{}
	for _, fn := range page.Fonts() {
		f := page.Font(fn)
		fontMap[fn] = &f
	}
	if plain, perr := page.GetPlainText(fontMap); perr == nil && strings.TrimSpace(plain) != "" {
		return plain, nil
	}

	// 回退：读取内容流中的文本片段（适合未压缩/结构简单的 PDF）。
	content := page.Content()
	texts := content.Text
	if len(texts) == 0 {
		return "", nil
	}

	sorted := make([]pdf.Text, len(texts))
	copy(sorted, texts)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Y != sorted[j].Y {
			return sorted[i].Y > sorted[j].Y // PDF Y 向上，大者在上
		}
		return sorted[i].X < sorted[j].X
	})

	var buf bytes.Buffer
	var prevY float64
	first := true
	for _, t := range sorted {
		if t.S == "" {
			continue
		}
		if !first && prevY-t.Y > t.FontSize/2 {
			buf.WriteByte('\n')
		}
		buf.WriteString(t.S)
		prevY = t.Y
		first = false
	}
	out = buf.String()
	return out, err
}

// Index 返回页序号（0-based）。
func (p *Page) Index() int {
	return p.num - 1
}
