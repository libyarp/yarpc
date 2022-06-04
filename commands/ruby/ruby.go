package ruby

import (
	"fmt"
	"github.com/libyarp/idl"
	"github.com/libyarp/yarpc/common"
	"github.com/urfave/cli"
	"os"
	"strings"
)

func processType(s *idl.FileSet, p map[string]string, t string) (provider string, typeName string) {
	if s.FromSamePackage(t) {
		return "", t
	}
	pkg, name := idl.SplitComponents(t)
	if pr, ok := p[pkg]; ok {
		url := strings.Split(pr, "/")
		return pkg, url[len(url)-1] + "." + name
	}
	if pkg != "" {
		pkc := strings.Split(pkg, ".")
		return "", pkc[len(pkc)-1] + "." + name
	}
	return "", name
}

var Action = cli.Command{
	Name:  "ruby",
	Usage: "Compiles provided yarp files into Ruby sources",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "class-path"},
		&cli.BoolFlag{Name: "single-file"},
		&cli.StringFlag{Name: "out", Required: true},
		&cli.StringSliceFlag{
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
		set, err := common.ProcessInputFiles(c)
		if err != nil {
			return err
		}

		var classPathComponents []string
		if cpArgs := c.String("class-path"); cpArgs != "" {
			if strings.ContainsRune(cpArgs, '.') {
				return fmt.Errorf("invalid value for --class-path: Expected path in format Module[::Module...]")
			}
			classPathComponents = strings.Split(cpArgs, "::")
		} else {
			classPathComponents = strings.Split(set.Package(), ".")
			for i := range classPathComponents {
				classPathComponents[i] = common.SnakeToCamel(classPathComponents[i])
			}
		}
		usedImportsProviders := map[string]bool{}
		resolver := func(n string) string {
			r, s := processType(set, importsProviders, n)
			if r != "" {
				usedImportsProviders[r] = true
			}
			return s
		}

		var files []File
		for _, m := range set.Messages {
			msg := File{
				Location: append(classPathComponents, common.CamelToSnake(m.Name)+".rb"),
				Data:     *common.NewBuilder(),
			}
			b := msg.Data
			for _, l := range m.Comments {
				b.Printfln("# %s", l)
			}
			msgName := common.Titleize(m.Name)
			b.Printfln("class %s < Yarp::Structure", msgName)
			b.Indent(func(b *common.Builder) {
				b.Printfln("yarp_meta id: 0x%08x, package: \"%s\", name: :%s", common.GenerateID(set, m), set.Package(), m.Name)
				for _, v := range m.Fields {
					switch f := v.(type) {
					case idl.Field:
						b.Printfln(composeClassField(f, resolver))
					case idl.OneOfField:
						composeOneofField(b, f, resolver)
					default:
						panic(fmt.Sprintf("unexpected %T", f))
					}
				}
			})
			b.Printfln("end")
			msg.Data = b
			files = append(files, msg)
		}

		for _, s := range set.Services {
			name := common.Titleize(s.Name)
			msg := common.NewBuilder()

			msg.Printfln("class %sClient < ")
		}

		if c.Bool("single-file") {
			f, err := os.OpenFile(c.String("out"), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0655)
			if err != nil {
				return err
			}
			defer f.Close()
			b := common.NewBuilder()
			concatFiles(b, classPathComponents, files)
			if _, err = f.WriteString(cleanup(b)); err != nil {
				return err
			}
		}
		return nil
	},
}
