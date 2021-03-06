package golang

import (
	"fmt"
	"github.com/libyarp/idl"
	"github.com/libyarp/yarpc/common"
	"github.com/urfave/cli"
	"go/format"
	"os"
	"strings"
)

func clientMethodReturnValue(m *idl.Method, resolver func(n string) string) string {
	if m.ReturnType == "void" {
		return "(yarp.Header, error)"
	}

	streaming := ""
	if m.ReturnStreaming {
		streaming = "<-chan "
	}
	return fmt.Sprintf("(%s*%s, yarp.Header, error)", streaming, resolver(m.ReturnType))
}

func clientMethodArguments(m *idl.Method, resolver func(n string) string) string {
	if m.ArgumentType == "void" {
		return "(ctx context.Context, optHeaders map[string]string)"
	}

	return fmt.Sprintf("(ctx context.Context, req *%s, optHeaders map[string]string)", resolver(m.ArgumentType))
}

func serverMethodArguments(m *idl.Method, resolver func(n string) string) string {
	fields := []string{
		"ctx context.Context",
		"headers yarp.Header",
	}
	if m.ArgumentType != "void" {
		fields = append(fields, "req *"+resolver(m.ArgumentType))
	}

	if m.ReturnStreaming {
		ret := resolver(m.ReturnType)
		if strings.ContainsRune(ret, '.') {
			c := strings.Split(ret, ".")
			ret = "Ext" + c[len(c)-1]
		}
		fields = append(fields, "out *"+ret+"Streamer")
	}
	return "(" + strings.Join(fields, ",") + ")"
}

func serverMethodReturnValues(m *idl.Method, resolver func(n string) string) string {
	var fields []string
	if !m.ReturnStreaming {
		fields = append(fields, "yarp.Header")
		if m.ReturnType != "void" {
			fields = append(fields, "*"+resolver(m.ReturnType))
		}
	}
	return "(" + strings.Join(append(fields, "error"), ",") + ")"
}

var Action = cli.Command{
	Name:  "go",
	Usage: "Compiles provided yarp files into Golang sources",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "package"},
		&cli.StringFlag{Name: "out", Required: true},
		cli.StringSliceFlag{
			Name:     "provided-by",
			Usage:    "PACKAGE=URL",
			Required: false,
		},
	},
	Action: func(c *cli.Context) error {
		importsProviders := map[string]string{}
		for _, v := range c.StringSlice("provided-by") {
			components := strings.SplitN(v, "=", 2)
			importsProviders[components[0]] = components[1]
		}
		usedImportsProviders := map[string]bool{}

		p := c.String("package")
		set, err := common.ProcessInputFiles(c)
		if err != nil {
			return err
		}
		if p == "" {
			c := strings.Split(set.Package(), ".")
			p = c[len(c)-1]
		}
		f := common.NewBuilder()
		f.Printfln("// Code generated by yarpc. DO NOT EDIT.\n")
		f.Printfln("package %s\n", p)

		imports := []string{
			`"context"`,
			`"reflect"`,
			`"github.com/libyarp/yarp"`,
		}
		var initializers []string

		b := common.NewBuilder()

		resolver := func(n string) string {
			r, s := processType(set, importsProviders, n)
			if r != "" {
				usedImportsProviders[r] = true
			}
			return s
		}

		for _, m := range set.Messages {

			for _, l := range m.Comments {
				b.Printfln("// %s", l)
			}
			msgName := common.Titleize(m.Name)
			initializers = append(initializers, msgName)
			b.Printfln("type %s struct {", msgName)
			b.Printfln("*yarp.Structure")

			for _, v := range m.Fields {
				switch f := v.(type) {
				case idl.Field:
					b.Printfln("%s %s `index:\"%d\"`", common.SnakeToCamel(f.Name), fieldType(f, resolver), f.Index)
				case idl.OneOfField:
				default:
					panic(fmt.Sprintf("unexpected %T", f))
				}
			}

			b.Printfln("}\n")
			b.Printfln("func (%s) YarpID() uint64 { return 0x%08x }", msgName, common.GenerateID(set, m))
			b.Printfln("func (%s) YarpPackage() string { return \"%s\" }", msgName, set.Package())
			b.Printfln("func (%s) YarpStructName() string { return \"%s\" }", msgName, m.Name)
		}

		wantedStreamers := map[string]bool{}

		for _, s := range set.Services {
			for _, l := range s.Comments {
				b.Printfln("// %s", l)
			}
			b.Printfln("type %sClient interface{", common.Titleize(s.Name))
			for _, m := range s.Methods {
				for _, l := range m.Comments {
					b.Printfln("// %s", l)
				}

				b.Printfln("%s%s %s", common.SnakeToCamel(m.Name), clientMethodArguments(&m, resolver), clientMethodReturnValue(&m, resolver))
			}
			b.Printfln("}\n")

			for _, l := range s.Comments {
				b.Printfln("// %s", l)
			}
			b.Printfln("type %sServer interface{", common.Titleize(s.Name))
			for _, m := range s.Methods {
				for _, l := range m.Comments {
					b.Printfln("// %s", l)
				}
				if m.ReturnStreaming {
					wantedStreamers[m.ReturnType] = true
				}

				b.Printfln("%s%s %s", common.SnakeToCamel(m.Name), serverMethodArguments(&m, resolver), serverMethodReturnValues(&m, resolver))
			}
			b.Printfln("}\n")
		}

		for k, v := range importsProviders {
			if _, ok := usedImportsProviders[k]; !ok {
				continue
			}
			ks := strings.Split(k, ".")
			imports = append(imports, fmt.Sprintf("%s \"%s\"", ks[len(ks)-1], v))
		}

		for _, i := range imports {
			f.Printfln("import %s", i)
		}

		f.Printf("\n")

		{
			var items []string
			for _, i := range initializers {
				items = append(items, i+"{}")
			}
			if len(items) > 0 {
				f.Printfln("func RegisterMessages() {")
				f.Printf("yarp.RegisterStructType(")
				f.Printf(strings.Join(items, ","))
				f.Printfln(")")
				f.Printfln("}")
			}
		}

		for _, s := range set.Services {
			name := common.Titleize(s.Name)
			b.Printfln("func New%sClient(addr string, opts ...yarp.Option) %sClient {", name, name)
			b.Printfln("return &_yarpClient%s{c: yarp.NewClient(addr, opts...)}", name)
			b.Printfln("}\n")
			b.Printfln("type _yarpClient%s struct {", common.Titleize(s.Name))
			b.Printfln("c *yarp.Client")
			b.Printfln("}")

			for _, m := range s.Methods {
				acceptsVoid := m.ArgumentType == "void"
				returnsVoid := m.ReturnType == "void"

				ret := resolver(m.ReturnType)
				b.Printfln("func (cli *_yarpClient%s) %s%s %s {", name, common.SnakeToCamel(m.Name), clientMethodArguments(&m, resolver), clientMethodReturnValue(&m, resolver))
				b.Printfln("request := yarp.Request{")
				b.Printfln("Method: 0x%08x,", common.GenerateMethodID(set, s, &m))
				b.Printfln("Headers: optHeaders,")
				b.Printfln("}\n")

				if m.ReturnStreaming && !returnsVoid {
					b.Printf("res, headers, err := cli.c.DoRequestStreamed(ctx, request")
					if acceptsVoid {
						b.Printf(", nil")
					} else {
						b.Printf(", req")
					}
					b.Printfln(")")
					b.Printfln("if err != nil { return nil, nil, err }")
					b.Printfln("ch := make(chan *%s, 10)", ret)
					b.Printfln("go func() {")
					b.Printfln("defer close(ch)")
					b.Printfln("for i := range res {")
					b.Printfln("if v, ok := i.(*%s); ok {", ret)
					b.Printfln("ch <- v")
					b.Printfln("}") // if
					b.Printfln("}") // for
					b.Printfln("}()")
					b.Printfln("return ch, headers, nil")
				} else {
					if returnsVoid {
						b.Printf("_")
					} else {
						b.Printf("res")
					}
					b.Printf(", headers, err := cli.c.DoRequest(ctx, request")
					if acceptsVoid {
						b.Printf(", nil")
					} else {
						b.Printf(", req")
					}
					b.Printfln(")")

					if returnsVoid {
						b.Printfln("if err != nil { return nil, err }")
					} else {
						b.Printfln("if err != nil { return nil, nil, err }")
					}
					if !returnsVoid {
						b.Printfln("if t, ok := res.(*%s); ok {", ret)
						b.Printfln("return t, headers, nil")
						b.Printfln("}")
						b.Printfln("return nil, nil, yarp.IncompatibleTypeError{")
						b.Printfln("Received: res,")
						b.Printfln("Wants: reflect.TypeOf(&%s{}),", ret)
						b.Printfln("}")
					} else {
						b.Printfln("return headers, nil")
					}
				}
				b.Printfln("}\n")
			}
		}

		for _, s := range set.Services {
			b.Printfln("func Register%s(s *yarp.Server, v %sServer) {", common.Titleize(s.Name), common.Titleize(s.Name))
			for _, v := range s.Methods {
				b.Printfln("s.RegisterHandler(0x%08x, \"%s.%s.%s\", v.%s)",
					common.GenerateMethodID(set, s, &v),
					set.Package(), s.Name, v.Name,
					common.SnakeToCamel(v.Name))
			}
			b.Printfln("}\n")
		}

		for w := range wantedStreamers {
			name := ""
			if strings.ContainsRune(w, '.') {
				name = "Ext"
				c := strings.Split(w, ".")
				name += c[len(c)-1]
			} else {
				name = w
			}
			targetType := resolver(w)

			b.Printfln("type %sStreamer struct{", name)
			b.Printfln("h yarp.Header")
			b.Printfln("ch chan<- *%s", targetType)
			b.Printfln("}")
			b.Printfln("func (i %sStreamer) Headers() yarp.Header { return i.h }", name)
			b.Printfln("func (i %sStreamer) Push(v *%s) { i.ch <- v }", name, targetType)
		}

		f.Absorb(b)

		for k := range importsProviders {
			if _, ok := usedImportsProviders[k]; !ok {
				fmt.Printf("WARNING: --provided-by flag set, but unused: %s\n", k)
			}
		}

		src, err := format.Source([]byte(f.String()))
		if err != nil {
			return fmt.Errorf("BUG: Error formatting source: %w\n\n%s", err, f.String())
		}

		path := c.String("out")
		outF, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("error opening %s: %w", path, err)
		}
		defer outF.Close()
		_, err = outF.Write(src)
		if err != nil {
			return fmt.Errorf("error writing %s: %w", path, err)
		}

		fmt.Printf("Wrote: %s\n", path)

		return nil
	},
}
