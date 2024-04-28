package utils

import "strings"

func GetFileExt(path string) string {
	lstIndex := strings.LastIndex(path, ".")
	if lstIndex == -1 {
		return ""
	}
	return path[lstIndex:]
}
