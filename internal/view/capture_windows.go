//go:build windows

package view

import (
	"fmt"
	"image"
	"image/color"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

// DPI 感知必须在**任何 GDI 调用之前**声明：
//
//	没声明时进程看到的是虚拟桌面坐标（实测 1920×1200 变成 1536×960），窗口截图的
//	右侧被剪掉 8~14px、左侧出现 7px 黑边 —— 而图看起来仍然"像一张正常的截图"。
//
// 顺序用程序保证：所有抓图入口都先调 ensureDPI()（sync.Once），它内部只调一次
// SetProcessDpiAwarenessContext(-4)（回退 SetProcessDPIAware）。
var (
	dpiOnce   sync.Once
	modUser32 = syscall.NewLazyDLL("user32.dll")
	modGdi32  = syscall.NewLazyDLL("gdi32.dll")

	procSetProcessDpiAwarenessContext = modUser32.NewProc("SetProcessDpiAwarenessContext")
	procSetProcessDPIAware            = modUser32.NewProc("SetProcessDPIAware")
	procEnumWindows                   = modUser32.NewProc("EnumWindows")
	procIsWindowVisible               = modUser32.NewProc("IsWindowVisible")
	procGetWindowTextW                = modUser32.NewProc("GetWindowTextW")
	procGetClassNameW                 = modUser32.NewProc("GetClassNameW")
	procGetWindowRect                 = modUser32.NewProc("GetWindowRect")
	procGetWindowThreadProcessID      = modUser32.NewProc("GetWindowThreadProcessId")
	procGetDC                         = modUser32.NewProc("GetDC")
	procReleaseDC                     = modUser32.NewProc("ReleaseDC")
	procPrintWindow                   = modUser32.NewProc("PrintWindow")

	procCreateCompatibleDC     = modGdi32.NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = modGdi32.NewProc("CreateCompatibleBitmap")
	procSelectObject           = modGdi32.NewProc("SelectObject")
	procDeleteObject           = modGdi32.NewProc("DeleteObject")
	procDeleteDC               = modGdi32.NewProc("DeleteDC")
	procGetDIBits              = modGdi32.NewProc("GetDIBits")
)

const pwRenderFullContent = 0x00000002

// ensureDPI declares per-monitor DPI awareness once, before any GDI call.
func ensureDPI() {
	dpiOnce.Do(func() {
		const dpiAwarenessPerMonitorV2 = ^uintptr(3) // -4
		if ok, _, _ := procSetProcessDpiAwarenessContext.Call(dpiAwarenessPerMonitorV2); ok != 0 {
			return
		}
		procSetProcessDPIAware.Call()
	})
}

type winRect struct{ Left, Top, Right, Bottom int32 }

// List enumerates top-level windows. pid == 0 lists every process (diagnostics only).
//
// 枚举是同步跨进程调用，目标线程忙时会卡住：所以枚举本身也带超时（见 ListTimeout）。
func List(pid uint32) ([]Window, error) {
	ensureDPI()
	var out []Window
	cb := syscall.NewCallback(func(hwnd uintptr, _ uintptr) uintptr {
		var r winRect
		ok, _, _ := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&r)))
		if ok == 0 {
			return 1
		}
		var wpid uint32
		procGetWindowThreadProcessID.Call(hwnd, uintptr(unsafe.Pointer(&wpid)))
		if pid != 0 && wpid != pid {
			return 1
		}
		vis, _, _ := procIsWindowVisible.Call(hwnd)
		title := windowText(hwnd)
		class := className(hwnd)
		out = append(out, Window{
			HWND: hwnd, PID: wpid, Title: title, Class: class, Visible: vis != 0,
			Rect: Rect{X: int(r.Left), Y: int(r.Top), W: int(r.Right - r.Left), H: int(r.Bottom - r.Top)},
		})
		return 1
	})
	procEnumWindows.Call(cb, 0)
	return out, nil
}

// ListTimeout runs List in a goroutine and gives up after d: 同步枚举卡住时不能让
// 整条命令一起卡住（真机上出现过目标线程忙导致的阻塞）。
func ListTimeout(pid uint32, d time.Duration) ([]Window, error) {
	type res struct {
		wins []Window
		err  error
	}
	ch := make(chan res, 1)
	go func() {
		w, err := List(pid)
		ch <- res{w, err}
	}()
	select {
	case r := <-ch:
		return r.wins, r.err
	case <-time.After(d):
		return nil, fmt.Errorf("view: 枚举窗口超过 %s 没回来（目标线程忙？）", d)
	}
}

func windowText(hwnd uintptr) string {
	buf := make([]uint16, 1024)
	n, _, _ := procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf[:n])
}

func className(hwnd uintptr) string {
	buf := make([]uint16, 256)
	n, _, _ := procGetClassNameW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf[:n])
}

// Capture grabs one window with PrintWindow(+PW_RENDERFULLCONTENT).
//
// 这是首选后端：不受遮挡、含标题栏/滚动条、不渗透背后的窗口。返回的图**未裁阴影**，
// 交给 TrimShadow（分开是为了让"抓"和"裁"各自可测）。
func Capture(hwnd uintptr) (*image.RGBA, error) {
	ensureDPI()
	var r winRect
	if ok, _, _ := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&r))); ok == 0 {
		return nil, fmt.Errorf("view: 拿不到窗口矩形")
	}
	w, h := int(r.Right-r.Left), int(r.Bottom-r.Top)
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("view: 窗口尺寸不合法 %dx%d", w, h)
	}
	hdc, _, _ := procGetDC.Call(hwnd)
	if hdc == 0 {
		return nil, fmt.Errorf("view: GetDC 失败")
	}
	defer procReleaseDC.Call(hwnd, hdc)
	memDC, _, _ := procCreateCompatibleDC.Call(hdc)
	if memDC == 0 {
		return nil, fmt.Errorf("view: CreateCompatibleDC 失败")
	}
	defer procDeleteDC.Call(memDC)
	bitmap, _, _ := procCreateCompatibleBitmap.Call(hdc, uintptr(w), uintptr(h))
	if bitmap == 0 {
		return nil, fmt.Errorf("view: CreateCompatibleBitmap 失败")
	}
	defer procDeleteObject.Call(bitmap)
	old, _, _ := procSelectObject.Call(memDC, bitmap)
	defer procSelectObject.Call(memDC, old)

	if ok, _, _ := procPrintWindow.Call(hwnd, memDC, pwRenderFullContent); ok == 0 {
		return nil, fmt.Errorf("view: PrintWindow 失败")
	}

	// BITMAPINFOHEADER + 32bpp top-down DIB
	type bitmapInfoHeader struct {
		Size          uint32
		Width         int32
		Height        int32
		Planes        uint16
		BitCount      uint16
		Compression   uint32
		SizeImage     uint32
		XPelsPerMeter int32
		YPelsPerMeter int32
		ClrUsed       uint32
		ClrImportant  uint32
	}
	bi := bitmapInfoHeader{
		Size: uint32(unsafe.Sizeof(bitmapInfoHeader{})), Width: int32(w), Height: -int32(h),
		Planes: 1, BitCount: 32, Compression: 0, // BI_RGB
	}
	buf := make([]byte, w*h*4)
	ret, _, _ := procGetDIBits.Call(memDC, bitmap, 0, uintptr(h),
		uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&bi)), 0)
	if ret == 0 {
		return nil, fmt.Errorf("view: GetDIBits 失败")
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := (y*w + x) * 4
			img.SetRGBA(x, y, color.RGBA{R: buf[i+2], G: buf[i+1], B: buf[i], A: 255})
		}
	}
	return img, nil
}
