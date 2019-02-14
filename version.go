package w32

import (
	"unsafe"
	"syscall"
	"fmt"
)

var (
	version = syscall.NewLazyDLL("version.dll")

	getFileVersionInfoSize = version.NewProc("GetFileVersionInfoSizeW")
	getFileVersionInfo     = version.NewProc("GetFileVersionInfoW")
	verQueryValue = version.NewProc("VerQueryValueW")
)

type VS_FIXEDFILEINFO struct {
	Signature        uint32
	StrucVersion     uint32
	FileVersionMS    uint32
	FileVersionLS    uint32
	ProductVersionMS uint32
	ProductVersionLS uint32
	FileFlagsMask    uint32
	FileFlags        uint32
	FileOS           uint32
	FileType         uint32
	FileSubtype      uint32
	FileDateMS       uint32
	FileDateLS       uint32
}


func GetFileVersionInfoSize(path string) uint32 {
	ret, _, _ := getFileVersionInfoSize.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(path))),
		0,
	)
	return uint32(ret)
}

func GetFileVersionInfo(path string, data []byte) bool {
	ret, _, _ := getFileVersionInfo.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(path))),
		0,
		uintptr(len(data)),
		uintptr(unsafe.Pointer(&data[0])),
	)
	return ret != 0
}

// VerQueryValueRoot calls VerQueryValue
// (https://msdn.microsoft.com/en-us/library/windows/desktop/ms647464(v=vs.85).aspx)
// with `\` (root) to retieve the VS_FIXEDFILEINFO.
func VerQueryValueRoot(block []byte) (VS_FIXEDFILEINFO, bool) {
	var offset uintptr
	var length uint
	blockStart := uintptr(unsafe.Pointer(&block[0]))
	ret, _, _ := verQueryValue.Call(
		blockStart,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(`\`))),
		uintptr(unsafe.Pointer(&offset)),
		uintptr(unsafe.Pointer(&length)),
	)
	if ret == 0 {
		return VS_FIXEDFILEINFO{}, false
	}
	start := int(offset) - int(blockStart)
	end := start + int(length)
	if start < 0 || start >= len(block) || end < start || end > len(block) {
		return VS_FIXEDFILEINFO{}, false
	}
	data := block[start:end]
	info := *((*VS_FIXEDFILEINFO)(unsafe.Pointer(&data[0])))
	return info, true
}

// VerQueryValueTranslations calls VerQueryValue
// (https://msdn.microsoft.com/en-us/library/windows/desktop/ms647464(v=vs.85).aspx)
// with `\VarFileInfo\Translation` to retrieve a list of 4-character translation
// strings as required by VerQueryValueString.
func VerQueryValueTranslations(block []byte) ([]string, bool) {
	var offset uintptr
	var length uint
	blockStart := uintptr(unsafe.Pointer(&block[0]))
	ret, _, _ := verQueryValue.Call(
		blockStart,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(`\VarFileInfo\Translation`))),
		uintptr(unsafe.Pointer(&offset)),
		uintptr(unsafe.Pointer(&length)),
	)
	if ret == 0 {
		return nil, false
	}
	start := int(offset) - int(blockStart)
	end := start + int(length)
	if start < 0 || start >= len(block) || end < start || end > len(block) {
		return nil, false
	}
	data := block[start:end]
	// each translation consists of a 16-bit language ID and a 16-bit code page
	// ID, so each entry has 4 bytes
	if len(data)%4 != 0 {
		return nil, false
	}
	trans := make([]string, len(data)/4)
	for i := range trans {
		t := data[i*4 : (i+1)*4]
		// handle endianness of the 16-bit values
		t[0], t[1] = t[1], t[0]
		t[2], t[3] = t[3], t[2]
		trans[i] = fmt.Sprintf("%x", t)
	}
	return trans, true
}

// these constants can be passed to VerQueryValueString as the item
const (
	CompanyName      = "CompanyName"
	FileDescription  = "FileDescription"
	FileVersion      = "FileVersion"
	LegalCopyright   = "LegalCopyright"
	LegalTrademarks  = "LegalTrademarks"
	OriginalFilename = "OriginalFilename"
	ProductVersion   = "ProductVersion"
	PrivateBuild     = "PrivateBuild"
	SpecialBuild     = "SpecialBuild"
)

// VerQueryValueString calls VerQueryValue
// (https://msdn.microsoft.com/en-us/library/windows/desktop/ms647464(v=vs.85).aspx)
// with `\StringFileInfo\...` to retrieve a specific piece of information as
// string in a specific translation.
func VerQueryValueString(block []byte, translation, item string) (string, bool) {
	var offset uintptr
	var utf16Length uint
	blockStart := uintptr(unsafe.Pointer(&block[0]))
	id := `\StringFileInfo\` + translation + `\` + item
	ret, _, _ := verQueryValue.Call(
		blockStart,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(id))),
		uintptr(unsafe.Pointer(&offset)),
		uintptr(unsafe.Pointer(&utf16Length)),
	)
	if ret == 0 {
		return "", false
	}
	start := int(offset) - int(blockStart)
	end := start + int(2*utf16Length)
	if start < 0 || start >= len(block) || end < start || end > len(block) {
		return "", false
	}
	data := block[start:end]
	u16 := make([]uint16, utf16Length)
	for i := range u16 {
		u16[i] = uint16(data[i*2+1])<<8 | uint16(data[i*2+0])
	}
	return syscall.UTF16ToString(u16), true
}