package exec

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/jolt9dev/go-env"
	fs "github.com/jolt9dev/go-fs"
	"github.com/jolt9dev/go-xstrings"
)

var (
	whichCache = make(map[string]string)
)

type WhichParams struct {
	UseCache     bool
	PrependPaths []string
}

type WhichOption func(*WhichParams)

func WithUseCache() WhichOption {
	return func(p *WhichParams) {
		p.UseCache = true
	}
}

func WithPrependPaths(paths ...string) WhichOption {
	return func(p *WhichParams) {
		p.PrependPaths = append(p.PrependPaths, paths...)
	}
}

func Which(command string, options ...WhichOption) (string, bool) {
	params := &WhichParams{}

	for _, option := range options {
		option(params)
	}

	if command == "" {
		return "", false
	}

	base := filepath.Base(command)
	ext := filepath.Ext(command)
	name := base[0 : len(base)-len(ext)]
	if params.UseCache {
		path, ok := whichCache[name]
		if ok {
			return path, true
		}
	}

	if filepath.IsAbs(command) {
		fi, err := os.Lstat(command)

		if err != nil {
			return "", false
		}

		if fi.Mode()&os.ModeSymlink != 0 {
			path, err := exec.LookPath(command)
			if err != nil {
				return "", false
			}

			if params.UseCache {
				whichCache[name] = path
			}

			return path, true
		}
	}

	pathSegments := []string{}
	if len(params.PrependPaths) > 0 {
		pathSegments = append(pathSegments, params.PrependPaths...)
	}

	pathSegments = append(pathSegments, env.SplitPath()...)

	for i, path := range pathSegments {
		value, _ := env.Expand(path)
		if value == "" {
			continue
		}

		pathSegments[i] = value
	}

	for _, path := range pathSegments {
		if xstrings.IsEmptySpace(path) || !fs.Exists(path) {
			continue
		}

		if runtime.GOOS == "windows" {
			pathExt := env.Get("PATHEXT")
			if xstrings.IsEmptySpace(pathExt) {
				pathExt = ".com;.exe;.bat;.cmd;.vbs;.vbe;.js;.jse;.wsf;.wsh"
			} else {
				pathExt = strings.ToLower(pathExt)
			}

			extSegments := strings.Split(pathExt, ";")

			entries, err := os.ReadDir(path)
			if err != nil {
				// TODO: debug/trace this erro
				continue
			}

			hasExt := len(ext) > 0

			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}

				entryName := entry.Name()
				entryExt := filepath.Ext(entryName)
				entryHasExt := len(entryExt) > 0

				// must have an extension on windows to execute
				if !entryHasExt && !hasExt {
					continue
				}

				if entryHasExt && hasExt {
					if strings.EqualFold(entryName, command) {
						fp := filepath.Join(path, entryName)
						whichCache[name] = fp
						return fp, true
					}
					continue
				}

				if entryHasExt {
					entryExt = strings.ToLower(entryExt)
				}

				if entryHasExt && !hasExt && slices.Contains(extSegments, entryExt) {
					try := name + entryExt
					if strings.EqualFold(try, entryName) {
						fp := filepath.Join(path, entryName)
						whichCache[name] = fp
						return fp, true
					}
				}
			}
		} else {
			entries, err := os.ReadDir(path)
			if err != nil {
				// TODO: debug/trace this erro
				continue
			}
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}

				if strings.EqualFold(entry.Name(), name) {
					fp := filepath.Join(path, entry.Name())
					whichCache[name] = fp
					return fp, true
				}
			}
		}
	}

	return "", false
}
