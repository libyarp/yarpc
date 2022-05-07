package common

import (
	"github.com/OneOfOne/xxhash"
	"github.com/libyarp/idl"
	"io"
	"strings"
)

func GenerateID(set *idl.FileSet, message *idl.Message) uint64 {
	h := xxhash.New64()
	r := strings.NewReader(set.Package() + "." + message.Name)
	_, _ = io.Copy(h, r)
	return h.Sum64()
}

func GenerateMethodID(set *idl.FileSet, service *idl.Service, method *idl.Method) uint64 {
	h := xxhash.New64()
	r := strings.NewReader(set.Package() + "." + service.Name + "." + method.Name)
	_, _ = io.Copy(h, r)
	return h.Sum64()
}
