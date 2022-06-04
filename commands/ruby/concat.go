package ruby

import (
	"github.com/libyarp/yarpc/common"
	"strings"
)

func concatFiles(b *common.Builder, basePath []string, files []File) {
	if len(basePath) > 0 {
		b.Printfln("module %s", basePath[0])
		b.Indent(func(b *common.Builder) {
			concatFiles(b, basePath[1:], files)
		})
		b.Printfln("end")
		return
	}

	for _, f := range files {
		for _, l := range strings.Split(f.Data.String(), "\n") {
			b.Printfln(l)
		}
	}
}

func cleanup(b *common.Builder) string {
	lines := strings.Split(b.String(), "\n")
	for i := range lines {
		if strings.TrimSpace(lines[i]) == "" {
			lines[i] = ""
		}
	}
	return strings.Join(lines, "\n")
}
