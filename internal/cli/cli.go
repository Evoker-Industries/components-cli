package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Evoker-Industries/components-cli/internal/component"
	"github.com/Evoker-Industries/components-cli/internal/config"
	"github.com/Evoker-Industries/components-cli/internal/forge"
	"github.com/Evoker-Industries/components-cli/internal/lockfile"
	"github.com/Evoker-Industries/components-cli/internal/secrets"
	"github.com/Evoker-Industries/components-cli/internal/source"
	"github.com/Evoker-Industries/components-cli/internal/storage"
	"github.com/Evoker-Industries/components-cli/internal/util"
)

type Options struct {
	File    string
	Storage string
	Lock    string
	Secrets string
	Offline bool
	Verbose bool
}

func defaults() Options {
	return Options{File: "components.json", Storage: ".components", Lock: "components.lock.json", Secrets: ".components.secrets.json"}
}

func Run(args []string) error {
	opts := defaults()
	fs := flag.NewFlagSet("components", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.StringVar(&opts.File, "file", opts.File, "registry file")
	fs.StringVar(&opts.Storage, "storage", opts.Storage, "storage root")
	fs.StringVar(&opts.Lock, "lock", opts.Lock, "lock file")
	fs.StringVar(&opts.Secrets, "secrets", opts.Secrets, "secrets file")
	fs.BoolVar(&opts.Offline, "offline", false, "offline mode")
	fs.BoolVar(&opts.Verbose, "verbose", false, "verbose output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	rest := fs.Args()
	if len(rest) == 0 {
		return usageError()
	}
	cmd := rest[0]
	args = rest[1:]
	switch cmd {
	case "init":
		return runInit(opts, args)
	case "add":
		return runAdd(opts, args)
	case "validate":
		return runValidate(opts)
	case "check":
		return runCheck(opts)
	case "list":
		return runList(opts)
	case "show":
		return runShow(opts, args)
	case "install":
		return runInstall(opts, args)
	case "update":
		return runUpdate(opts, args)
	case "current":
		return runCurrent(opts, args)
	case "rollback":
		return runRollback(opts, args)
	case "clean":
		return runClean(opts)
	case "secret":
		return runSecret(opts, args)
	default:
		return fmt.Errorf("unknown command %q", cmd)
	}
}

func usageError() error {
	return errors.New("usage: components [--file path] [--storage path] [--lock path] [--secrets path] <command>")
}

func runInit(opts Options, args []string) error {
	f := flag.NewFlagSet("init", flag.ContinueOnError)
	f.SetOutput(os.Stderr)
	force := f.Bool("force", false, "overwrite existing files")
	if err := f.Parse(args); err != nil {
		return err
	}
	if err := config.InitEmpty(opts.File, *force); err != nil {
		return err
	}
	if err := secrets.Init(opts.Secrets, *force); err != nil {
		return err
	}
	if err := util.EnsureGitIgnoreEntry(".", relFromRepo(opts.Secrets)); err != nil {
		return err
	}
	lf, err := lockfile.Load(opts.Lock)
	if err != nil {
		return err
	}
	return lockfile.Write(opts.Lock, lf)
}

func runAdd(opts Options, args []string) error {
	f := flag.NewFlagSet("add", flag.ContinueOnError)
	f.SetOutput(os.Stderr)
	version := f.String("version", "latest", "version")
	manifestPath := f.String("path", "", "manifest path")
	directory := f.String("directory", ".", "component directory")
	force := f.Bool("force", false, "replace existing component")
	installNow := f.Bool("install", false, "install immediately")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() < 2 {
		return errors.New("usage: components add <name> <source>")
	}
	name := f.Arg(0)
	src := f.Arg(1)
	if _, err := source.Parse(src); err != nil {
		return err
	}
	reg, err := config.Load(opts.File)
	if err != nil {
		return err
	}
	if _, exists := reg.Components[name]; exists && !*force {
		return fmt.Errorf("component %q already exists", name)
	}
	if strings.TrimSpace(*manifestPath) == "" {
		safe := storage.SafeName(name)
		*manifestPath = filepath.Join(filepath.Dir(opts.File), "components", safe, "component.json")
	}
	manifestAbs, err := util.MustAbs(*manifestPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(manifestAbs), 0o755); err != nil {
		return err
	}
	comp := &component.Component{Name: name, Source: src, Version: *version, Directory: *directory}
	if err := component.Save(manifestAbs, comp); err != nil {
		return err
	}
	regPathAbs, err := util.MustAbs(opts.File)
	if err != nil {
		return err
	}
	regDir := filepath.Dir(regPathAbs)
	reg.Components[name] = config.ComponentRef{Ref: util.RelOrAbs(regDir, manifestAbs)}
	if err := config.Save(opts.File, reg); err != nil {
		return err
	}
	if *installNow {
		return runInstall(opts, []string{name})
	}
	return nil
}

func runValidate(opts Options) error {
	reg, err := config.Load(opts.File)
	if err != nil {
		return err
	}
	if err := config.Validate(reg); err != nil {
		return err
	}
	for name, ref := range reg.Components {
		p := config.ResolveRef(reg, ref.Ref)
		if _, err := component.Load(p); err != nil {
			return fmt.Errorf("component %q: %w", name, err)
		}
	}
	return nil
}

func runCheck(opts Options) error {
	if err := runValidate(opts); err != nil {
		return err
	}
	reg, _ := config.Load(opts.File)
	fr := forge.NewRegistry()
	for name, ref := range reg.Components {
		comp, err := component.Load(config.ResolveRef(reg, ref.Ref))
		if err != nil {
			return err
		}
		uri, err := source.Parse(comp.Source)
		if err != nil {
			return fmt.Errorf("component %q: %w", name, err)
		}
		if _, err := fr.Resolve(uri, reg.Forges); err != nil {
			return fmt.Errorf("component %q: %w", name, err)
		}
	}
	return nil
}

func runList(opts Options) error {
	reg, err := config.Load(opts.File)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(reg.Components))
	for name := range reg.Components {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Println(name)
	}
	return nil
}

func runShow(opts Options, args []string) error {
	if len(args) != 1 {
		return errors.New("usage: components show <name>")
	}
	reg, err := config.Load(opts.File)
	if err != nil {
		return err
	}
	ref, ok := reg.Components[args[0]]
	if !ok {
		return fmt.Errorf("component %q not found", args[0])
	}
	comp, err := component.Load(config.ResolveRef(reg, ref.Ref))
	if err != nil {
		return err
	}
	fmt.Printf("name: %s\nsource: %s\nversion: %s\ndirectory: %s\n", comp.Name, comp.Source, comp.Version, comp.Directory)
	return nil
}

func runInstall(opts Options, args []string) error {
	if len(args) != 1 {
		return errors.New("usage: components install <name>")
	}
	name := args[0]
	reg, err := config.Load(opts.File)
	if err != nil {
		return err
	}
	ref, ok := reg.Components[name]
	if !ok {
		return fmt.Errorf("component %q not found", name)
	}
	comp, err := component.Load(config.ResolveRef(reg, ref.Ref))
	if err != nil {
		return err
	}
	uri, err := source.Parse(comp.Source)
	if err != nil {
		return err
	}
	resolved, err := forge.NewRegistry().Resolve(uri, reg.Forges)
	if err != nil {
		return err
	}
	requested := comp.Version
	if requested == "" {
		requested = "latest"
	}
	resolvedVersion := requested
	if opts.Offline {
		if _, err := (storage.FS{Root: opts.Storage}).Current(name); err == nil {
			return nil
		}
		return errors.New("offline mode requires an already installed component")
	}
	fs := storage.FS{Root: opts.Storage}
	installPath, err := fs.InstallVersion(name, resolvedVersion)
	if err != nil {
		return err
	}
	metadata := fmt.Sprintf("source=%s\nclone=%s\n", comp.Source, resolved.CloneURL)
	if err := os.WriteFile(filepath.Join(installPath, "manifest.txt"), []byte(metadata), 0o644); err != nil {
		return err
	}
	if err := fs.SetCurrent(name, resolvedVersion); err != nil {
		return err
	}
	lf, err := lockfile.Load(opts.Lock)
	if err != nil {
		return err
	}
	lf.Components[name] = lockfile.ComponentLock{
		Source:           comp.Source,
		RequestedVersion: requested,
		ResolvedVersion:  resolvedVersion,
		Path:             installPath,
		Verified:         true,
	}
	return lockfile.Write(opts.Lock, lf)
}

func runUpdate(opts Options, args []string) error {
	reg, err := config.Load(opts.File)
	if err != nil {
		return err
	}
	if len(args) == 1 {
		return runInstall(opts, args)
	}
	if len(args) > 1 {
		return errors.New("usage: components update [name]")
	}
	for name := range reg.Components {
		if err := runInstall(opts, []string{name}); err != nil {
			return err
		}
	}
	return nil
}

func runCurrent(opts Options, args []string) error {
	if len(args) != 1 {
		return errors.New("usage: components current <name>")
	}
	cur, err := (storage.FS{Root: opts.Storage}).Current(args[0])
	if err != nil {
		return err
	}
	fmt.Println(cur)
	return nil
}

func runRollback(opts Options, args []string) error {
	if len(args) != 1 {
		return errors.New("usage: components rollback <name>")
	}
	name := args[0]
	componentDir := (storage.FS{Root: opts.Storage}).ComponentPath(name)
	entries, err := os.ReadDir(componentDir)
	if err != nil {
		return err
	}
	versions := []string{}
	for _, e := range entries {
		if e.IsDir() {
			versions = append(versions, e.Name())
		}
	}
	sort.Strings(versions)
	if len(versions) < 2 {
		return errors.New("no prior version available")
	}
	return (storage.FS{Root: opts.Storage}).SetCurrent(name, versions[len(versions)-2])
}

func runClean(opts Options) error {
	return os.RemoveAll(filepath.Join(opts.Storage, ".staging"))
}

func runSecret(opts Options, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: components secret <set|list|remove>")
	}
	st, err := secrets.Load(opts.Secrets)
	if err != nil {
		return err
	}
	switch args[0] {
	case "set":
		if len(args) < 2 {
			return errors.New("usage: components secret set <name> <value>")
		}
		if len(args) < 3 {
			return errors.New("secret value required in non-interactive mode")
		}
		if err := st.Set(args[1], args[2]); err != nil {
			return err
		}
		if err := st.Save(); err != nil {
			return err
		}
		return util.EnsureGitIgnoreEntry(".", relFromRepo(opts.Secrets))
	case "list":
		for _, k := range st.Keys() {
			fmt.Println(k)
		}
		return nil
	case "remove":
		if len(args) != 2 {
			return errors.New("usage: components secret remove <name>")
		}
		if err := st.Remove(args[1]); err != nil {
			return err
		}
		return st.Save()
	default:
		return fmt.Errorf("unknown secret command %q", args[0])
	}
}

func relFromRepo(path string) string {
	abs, err := util.MustAbs(path)
	if err != nil {
		return path
	}
	repo, err := os.Getwd()
	if err != nil {
		return path
	}
	rel, err := filepath.Rel(repo, abs)
	if err != nil {
		return path
	}
	if strings.HasPrefix(rel, "..") {
		return path
	}
	if !strings.HasPrefix(rel, ".") {
		return "./" + rel
	}
	return rel
}
