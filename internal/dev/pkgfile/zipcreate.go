package pkgfile

// ---------------------------------------------------------------------------
// 从零造一个 .tzc：按 Info-ZIP 3.0（Unix）的形态写
//
// 为什么不用 archive/zip.Writer：
//
//	见 zipraw.go 顶部注释 —— 它**无条件**置 general purpose flag bit 3 并补数据描述符，
//	而设计器（.NET）读的是"CRC / 大小直接写在局部头里"那种形态，遇到描述符就报
//	「Data descriptor signature not found」。所以这里手写局部头与中央目录。
//
// 形态是从真包（服务器 sadzp050_zip_compress_to_code_zip 的产物）逐字节量出来的：
//
//	局部头：ver_need=20（deflate）/ 10（store），flags=0，crc+csize+usize 全部内联
//	        extra = UT(0x5455, 9 字节: flags=3 + mtime + atime) + ux(0x7875, uid/gid)
//	中央目录：version_made_by=0x031e（Info-ZIP 3.0 / Unix），extra 里 UT 只带 mtime
//	         external_attrs = unix mode << 16（0644 / 0775）
//	EOCD：无注释
//
// 内容本身不追求与 Info-ZIP 逐字节相同（deflate 实现不同），但**结构**必须一样：
// 设计器只认结构。
// ---------------------------------------------------------------------------

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"path"
	"strings"
	"time"
)

// 真包里量到的常量。
const (
	zipVersionMadeBy = 0x031e // Info-ZIP 3.0 on Unix
	zipVerDeflate    = 20     // 2.0：deflate
	zipVerStore      = 10     // 1.0：store
	zipUTTag         = 0x5455
	zipUxTag         = 0x7875
	zipUID           = 1002
	zipGID           = 1003
)

// WriteEntry 是一个待写入条目的内容。
type WriteEntry struct {
	// Name 是包内条目名（正斜杠、不允许绝对路径或 ..）。
	Name string
	// Data 是条目原文。
	Data []byte
	// Store=true 表示不压缩（真包里的 ver 条目就是 store）。
	Store bool
	// Mode 是 unix 权限位（如 0o664、0o775）。0 取 0o664（真包就是 664）。
	Mode uint32
	// ModTime 写进 DOS 时间与 UT extra。零值取当前时间。
	ModTime time.Time
}

// WriteZip 按设计器可打开的形态写出一个 zip。
//
// 顺序即 entries 顺序（真包里是 4gl / tap / tgl / ver，按扩展名字典序）。
func WriteZip(w io.Writer, entries []WriteEntry) error {
	if len(entries) == 0 {
		return fmt.Errorf("zip: 没有条目可写")
	}

	type rec struct {
		name    []byte
		method  uint16
		mtime   uint16
		mdate   uint16
		crc     uint32
		csize   uint32
		usize   uint32
		offset  uint32
		mode    uint32
		comp    []byte
		lExtra  []byte
		cExtra  []byte
		verNeed uint16
	}

	recs := make([]rec, 0, len(entries))
	var body bytes.Buffer

	for _, e := range entries {
		name := strings.TrimPrefix(strings.ReplaceAll(e.Name, "\\", "/"), "/")
		if name == "" || name != path.Clean(name) || strings.HasPrefix(name, "../") {
			return fmt.Errorf("zip: 条目名非法 %q", e.Name)
		}
		mt := e.ModTime
		if mt.IsZero() {
			mt = time.Now()
		}
		mtime, mdate := dosDateTime(mt)
		secs := uint32(mt.Unix())

		method := uint16(8)
		verNeed := uint16(zipVerDeflate)
		comp := e.Data
		if e.Store {
			method, verNeed = 0, zipVerStore
		} else {
			var err error
			comp, err = deflate(e.Data)
			if err != nil {
				return fmt.Errorf("zip: 压缩 %s 失败: %w", name, err)
			}
		}

		mode := e.Mode
		if mode == 0 {
			mode = 0o664
		}

		r := rec{
			name:    []byte(name),
			method:  method,
			mtime:   mtime,
			mdate:   mdate,
			crc:     crc32.ChecksumIEEE(e.Data),
			csize:   uint32(len(comp)),
			usize:   uint32(len(e.Data)),
			offset:  uint32(body.Len()),
			mode:    mode,
			comp:    comp,
			lExtra:  localExtra(secs),
			cExtra:  centralExtra(secs),
			verNeed: verNeed,
		}
		recs = append(recs, r)

		var hdr [30]byte
		binary.LittleEndian.PutUint32(hdr[0:], 0x04034b50)
		binary.LittleEndian.PutUint16(hdr[4:], r.verNeed)
		binary.LittleEndian.PutUint16(hdr[6:], 0) // flags：**不置 bit 3**
		binary.LittleEndian.PutUint16(hdr[8:], r.method)
		binary.LittleEndian.PutUint16(hdr[10:], r.mtime)
		binary.LittleEndian.PutUint16(hdr[12:], r.mdate)
		binary.LittleEndian.PutUint32(hdr[14:], r.crc)
		binary.LittleEndian.PutUint32(hdr[18:], r.csize)
		binary.LittleEndian.PutUint32(hdr[22:], r.usize)
		binary.LittleEndian.PutUint16(hdr[26:], uint16(len(r.name)))
		binary.LittleEndian.PutUint16(hdr[28:], uint16(len(r.lExtra)))
		body.Write(hdr[:])
		body.Write(r.name)
		body.Write(r.lExtra)
		body.Write(r.comp)
	}

	cdStart := body.Len()
	for _, r := range recs {
		var h [46]byte
		binary.LittleEndian.PutUint32(h[0:], 0x02014b50)
		binary.LittleEndian.PutUint16(h[4:], zipVersionMadeBy)
		binary.LittleEndian.PutUint16(h[6:], r.verNeed)
		binary.LittleEndian.PutUint16(h[8:], 0) // flags
		binary.LittleEndian.PutUint16(h[10:], r.method)
		binary.LittleEndian.PutUint16(h[12:], r.mtime)
		binary.LittleEndian.PutUint16(h[14:], r.mdate)
		binary.LittleEndian.PutUint32(h[16:], r.crc)
		binary.LittleEndian.PutUint32(h[20:], r.csize)
		binary.LittleEndian.PutUint32(h[24:], r.usize)
		binary.LittleEndian.PutUint16(h[28:], uint16(len(r.name)))
		binary.LittleEndian.PutUint16(h[30:], uint16(len(r.cExtra)))
		binary.LittleEndian.PutUint16(h[32:], 0) // comment len
		binary.LittleEndian.PutUint16(h[34:], 0) // disk
		binary.LittleEndian.PutUint16(h[36:], 1) // internal attrs：文本
		binary.LittleEndian.PutUint32(h[38:], (0o100000|r.mode)<<16)
		binary.LittleEndian.PutUint32(h[42:], r.offset)
		body.Write(h[:])
		body.Write(r.name)
		body.Write(r.cExtra)
	}
	cdSize := body.Len() - cdStart

	var eocd [22]byte
	binary.LittleEndian.PutUint32(eocd[0:], 0x06054b50)
	binary.LittleEndian.PutUint16(eocd[4:], 0)
	binary.LittleEndian.PutUint16(eocd[6:], 0)
	binary.LittleEndian.PutUint16(eocd[8:], uint16(len(recs)))
	binary.LittleEndian.PutUint16(eocd[10:], uint16(len(recs)))
	binary.LittleEndian.PutUint32(eocd[12:], uint32(cdSize))
	binary.LittleEndian.PutUint32(eocd[16:], uint32(cdStart))
	binary.LittleEndian.PutUint16(eocd[20:], 0)
	body.Write(eocd[:])

	_, err := w.Write(body.Bytes())
	return err
}

// localExtra 局部头的 extra：UT（flags=3 + mtime + atime）+ ux（uid/gid）。
func localExtra(secs uint32) []byte {
	var b bytes.Buffer
	var h [4]byte
	binary.LittleEndian.PutUint16(h[0:], zipUTTag)
	binary.LittleEndian.PutUint16(h[2:], 9)
	b.Write(h[:])
	b.WriteByte(3) // flags: mtime + atime
	var t [4]byte
	binary.LittleEndian.PutUint32(t[:], secs)
	b.Write(t[:])
	b.Write(t[:]) // atime 同 mtime

	binary.LittleEndian.PutUint16(h[0:], zipUxTag)
	binary.LittleEndian.PutUint16(h[2:], 11)
	b.Write(h[:])
	b.WriteByte(1) // version
	b.WriteByte(4) // uid 长度
	binary.LittleEndian.PutUint32(t[:], zipUID)
	b.Write(t[:])
	b.WriteByte(4) // gid 长度
	binary.LittleEndian.PutUint32(t[:], zipGID)
	b.Write(t[:])
	return b.Bytes()
}

// centralExtra 中央目录的 extra：UT 只带 mtime，ux 与局部头相同。
func centralExtra(secs uint32) []byte {
	var b bytes.Buffer
	var h [4]byte
	binary.LittleEndian.PutUint16(h[0:], zipUTTag)
	binary.LittleEndian.PutUint16(h[2:], 5)
	b.Write(h[:])
	b.WriteByte(3)
	var t [4]byte
	binary.LittleEndian.PutUint32(t[:], secs)
	b.Write(t[:])

	binary.LittleEndian.PutUint16(h[0:], zipUxTag)
	binary.LittleEndian.PutUint16(h[2:], 11)
	b.Write(h[:])
	b.WriteByte(1)
	b.WriteByte(4)
	binary.LittleEndian.PutUint32(t[:], zipUID)
	b.Write(t[:])
	b.WriteByte(4)
	binary.LittleEndian.PutUint32(t[:], zipGID)
	b.Write(t[:])
	return b.Bytes()
}

func deflate(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	fw, err := flate.NewWriter(&buf, flate.DefaultCompression)
	if err != nil {
		return nil, err
	}
	if _, err := fw.Write(data); err != nil {
		return nil, err
	}
	if err := fw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// dosDateTime 把时间转成 DOS 的 (time, date)。
func dosDateTime(t time.Time) (uint16, uint16) {
	if t.Year() < 1980 {
		t = time.Date(1980, 1, 1, 0, 0, 0, 0, t.Location())
	}
	tm := uint16(t.Hour()<<11 | t.Minute()<<5 | t.Second()/2)
	dt := uint16((t.Year()-1980)<<9 | int(t.Month())<<5 | t.Day())
	return tm, dt
}
