package db

import (
	"fmt"
	"golang.org/x/sys/windows"
	"os"
)

// PostgreSQL's Windows tools re-execute through the active ANSI code page.
// Resolve the existing directory to its Windows short path before invoking
// them, preserving the same persistent data location for Unicode profiles.
func prepareEmbeddedRoot(root string) (string, error) {
	if err := os.MkdirAll(root, 0o700); err != nil {
		return "", err
	}
	p, err := windows.UTF16PtrFromString(root)
	if err != nil {
		return "", err
	}
	buf := make([]uint16, 32768)
	n, err := windows.GetShortPathName(p, &buf[0], uint32(len(buf)))
	if err != nil {
		return "", fmt.Errorf("resolve database path: %w", err)
	}
	if n == 0 || int(n) >= len(buf) {
		return "", fmt.Errorf("invalid database short path")
	}
	short := windows.UTF16ToString(buf[:n])
	for _, r := range short {
		if r > 127 {
			return "", fmt.Errorf("数据库目录不支持当前 Windows 字符编码，请使用不含中文的 -data 路径：%s", root)
		}
	}
	return short, nil
}
