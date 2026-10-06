package drawio

// mxlibrary 的合成：catalog → 一个可以被 drawio 加载的 <mxlibrary> 文件。

// MxEntries 把一个库的条目转成 mxlibrary 里的条目。
//
// 只带 drawio 认的那几个字段：xml / w / h / title / aspect，非空的 tags 才写。
// 形状源里的其它东西（slots、pfill/pdeco、multiTab）是**我们自己的**合成期元数据，
// 塞进库文件只会让它变大且无意义。
func MxEntries(c Catalog) []MxEntry {
	out := make([]MxEntry, 0, len(c.Shapes))
	for _, s := range c.Shapes {
		e := MxEntry{XML: s.XML, W: s.W, H: s.H, Title: s.Title, Aspect: s.Aspect}
		if e.Aspect == "" {
			e.Aspect = "variable"
		}
		if len(s.Tags) > 0 {
			e.Tags = s.Tags
		}
		out = append(out, e)
	}
	return out
}

// MxLibraryOf 是一个库的完整库文件内容。
func MxLibraryOf(c Catalog) (string, error) {
	return MxLibrary(MxEntries(c))
}
