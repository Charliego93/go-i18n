package i18n

import (
	"fmt"
	"golang.org/x/text/language"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Loader interface{ ParseMessage(i *I18n) error }
type LoaderOp interface{ apply(cfg *FSLoader) }
type LoaderOpFunc func(cfg *FSLoader)
type unmarshalls map[string]UnmarshalFunc
type unmarshal struct {
	format string
	fn     UnmarshalFunc
}

func (u unmarshalls) apply(l *FSLoader)  { l.ums = u }
func (c LoaderOpFunc) apply(l *FSLoader) { c(l) }
func (u unmarshal) apply(l *FSLoader) {
	if l.ums == nil {
		l.ums = make(map[string]UnmarshalFunc)
	}
	l.ums[u.format] = u.fn
}

// WithUnmarshalls register multi format unmarshal func
func WithUnmarshalls(fns map[string]UnmarshalFunc) LoaderOp { return unmarshalls(fns) }

// WithUnmarshal register single format unmarshal func
func WithUnmarshal(format string, fn UnmarshalFunc) LoaderOp { return unmarshal{format, fn} }

func NewLoaderWithPath(path string, opts ...LoaderOp) Option {
	l := &FSLoader{fs: os.DirFS(path)}
	for _, opt := range opts {
		opt.apply(l)
	}
	return loader{l}
}

func NewLoaderWithFS(fs fs.FS, opts ...LoaderOp) Option {
	l := &FSLoader{fs: fs}
	for _, opt := range opts {
		opt.apply(l)
	}
	return loader{l}
}

type FSLoader struct {
	fs  fs.FS
	ums map[string]UnmarshalFunc
}

func (c *FSLoader) ParseMessage(i *I18n) error {
	for format, ufn := range c.ums {
		i.RegisterUnmarshalFunc(format, ufn)
	}

	return c.parseMessage(i, ".")
}

func (c *FSLoader) parse(name string, buf []byte) error {
	ns := strings.Split(name, ".")
	if len(name) == 0 || len(ns) < 2 {
		return fmt.Errorf("the file %s not ext", name)
	}

	format := ns[1]
	if _, ok := c.ums[format]; !ok {
		i.registerUnmarshalFunc(format)
	}

	tag, err := language.Parse(ns[0])
	if err != nil {
		return err
	}
	i.SetLocalizer(tag)
	i.MastParseMessageFileBytes(buf, name)
	return nil
}

func (c *FSLoader) parseMessage(i *I18n, path string) error {
	entries, err := fs.ReadDir(c.fs, path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		fp := filepath.Join(path, name)
		if entry.IsDir() {
			err := c.parseMessage(i, fp)
			if err != nil {
				return err
			}
		} else {
			buf, err := fs.ReadFile(c.fs, fp)
			if err != nil {
				return err
			}
			err = c.parse(name, buf)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
