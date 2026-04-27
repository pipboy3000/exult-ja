package main

import (
	"archive/zip"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var requiredFiles = []string{
	"usecode",
	"initgame.dat",
	"text.flx",
	"exultmsg.txt",
	"exult_u7j.cfg",
	"fonts.vga",
	"mainshp.flx",
	"endshape.flx",
	"endgame.dat",
}

var assetAliases = map[string][]string{
	"initgame.dat": {"initgame.dat.disabled"},
	"exultmsg.txt": {"u7exultmsg.txt"},
}

type options struct {
	gameDir  string
	assets   string
	patchDir string
	exult    string
	dryRun   bool
}

func main() {
	opts := parseFlags()
	if err := run(opts); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() options {
	var opts options
	flag.StringVar(&opts.gameDir, "game", "", "path to blackgate directory")
	flag.StringVar(&opts.assets, "assets", "", "path to japanese asset zip or extracted directory")
	flag.StringVar(&opts.patchDir, "patch-dir", "", "optional output patch directory")
	flag.StringVar(&opts.exult, "exult", "", "optional path to exult binary or app")
	flag.BoolVar(&opts.dryRun, "dry-run", false, "print actions without writing files")
	flag.Parse()
	return opts
}

func run(opts options) error {
	if opts.gameDir == "" {
		return errors.New("--game is required")
	}
	if opts.assets == "" {
		return errors.New("--assets is required")
	}

	gameDir, err := filepath.Abs(opts.gameDir)
	if err != nil {
		return fmt.Errorf("resolve --game: %w", err)
	}
	assetsPath, err := filepath.Abs(opts.assets)
	if err != nil {
		return fmt.Errorf("resolve --assets: %w", err)
	}

	staticDir, err := findStaticDir(gameDir)
	if err != nil {
		return err
	}
	if opts.patchDir == "" {
		opts.patchDir = defaultPatchDir(gameDir)
	}
	patchDir, err := filepath.Abs(opts.patchDir)
	if err != nil {
		return fmt.Errorf("resolve --patch-dir: %w", err)
	}

	if opts.exult != "" {
		if err := validateExultPath(opts.exult); err != nil {
			return err
		}
	}

	var assets map[string][]byte
	switch {
	case isZipPath(assetsPath):
		assets, err = readAssetsFromZip(assetsPath)
	default:
		assets, err = readAssetsFromDir(assetsPath)
	}
	if err != nil {
		return err
	}

	fmt.Printf("platform : %s\n", runtime.GOOS)
	fmt.Printf("game     : %s\n", gameDir)
	fmt.Printf("static   : %s\n", staticDir)
	fmt.Printf("assets   : %s\n", assetsPath)
	fmt.Printf("patch    : %s\n", patchDir)
	if opts.exult != "" {
		fmt.Printf("exult    : %s\n", opts.exult)
	}

	if opts.dryRun {
		fmt.Println("mode     : dry-run")
		return nil
	}

	if err := os.MkdirAll(patchDir, 0o755); err != nil {
		return fmt.Errorf("create patch dir: %w", err)
	}
	for _, name := range requiredFiles {
		target := filepath.Join(patchDir, name)
		if err := os.WriteFile(target, assets[name], 0o644); err != nil {
			return fmt.Errorf("write %s: %w", target, err)
		}
	}
	if err := disableInitGame(filepath.Join(patchDir, "initgame.dat")); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("staged japanese patch files.")
	fmt.Println()
	fmt.Println("next:")
	fmt.Printf("  1. ensure Exult blackgate patch path points to: %s\n", patchDir)
	fmt.Println("  2. keep original game data in STATIC/static untouched")
	fmt.Println("  3. launch the modified Exult and verify dialogue/books/signs/subtitles")
	fmt.Println()
	fmt.Println("known limitations:")
	fmt.Println("  - dialogue may still clip slightly at the right edge on rare lines")
	fmt.Println("  - UI is not fully localized")
	return nil
}

func findStaticDir(gameDir string) (string, error) {
	for _, name := range []string{"STATIC", "static"} {
		path := filepath.Join(gameDir, name)
		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			return path, nil
		}
	}
	return "", fmt.Errorf("expected STATIC or static under: %s", gameDir)
}

func validateExultPath(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve --exult: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return fmt.Errorf("stat --exult: %w", err)
	}
	if info.IsDir() {
		if runtime.GOOS == "darwin" && strings.HasSuffix(abs, ".app") {
			return nil
		}
		return fmt.Errorf("--exult points to a directory, expected executable or .app: %s", abs)
	}
	return nil
}

func defaultPatchDir(gameDir string) string {
	if cfgPatch, err := configuredPatchDir(); err == nil && cfgPatch != "" {
		return cfgPatch
	}

	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		parent := filepath.Dir(gameDir)
		return filepath.Join(parent, "blackgate-patch-ja")
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Exult", "blackgate-patch-ja")
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData != "" {
			return filepath.Join(localAppData, "Exult", "blackgate-patch-ja")
		}
		return filepath.Join(home, "AppData", "Local", "Exult", "blackgate-patch-ja")
	default:
		parent := filepath.Dir(gameDir)
		return filepath.Join(parent, "blackgate-patch-ja")
	}
}

func configuredPatchDir() (string, error) {
	var cfgPath string
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "", nil
	}

	switch runtime.GOOS {
	case "darwin":
		cfgPath = filepath.Join(home, "Library", "Preferences", "exult.cfg")
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		cfgPath = filepath.Join(localAppData, "Exult", "exult.cfg")
	default:
		return "", nil
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", fmt.Errorf("read exult.cfg: %w", err)
	}

	re := regexp.MustCompile(`(?s)<blackgate>.*?<patch>\s*([^<]+?)\s*</patch>.*?</blackgate>`)
	match := re.FindSubmatch(data)
	if len(match) < 2 {
		return "", nil
	}
	return strings.TrimSpace(string(match[1])), nil
}

func isZipPath(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".zip")
}

func readAssetsFromZip(path string) (map[string][]byte, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}
	defer reader.Close()

	assets := make(map[string][]byte, len(requiredFiles))
	for _, name := range requiredFiles {
		var found bool
		for _, f := range reader.File {
			if !matchesAssetName(filepath.Base(f.Name), name) {
				continue
			}
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("open %s in zip: %w", name, err)
			}
			data, readErr := io.ReadAll(rc)
			closeErr := rc.Close()
			if readErr != nil {
				return nil, fmt.Errorf("read %s in zip: %w", name, readErr)
			}
			if closeErr != nil {
				return nil, fmt.Errorf("close %s in zip: %w", name, closeErr)
			}
			assets[name] = data
			found = true
			break
		}
		if !found {
			return nil, fmt.Errorf("required file missing in zip: %s", name)
		}
	}
	return assets, nil
}

func readAssetsFromDir(root string) (map[string][]byte, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("stat asset dir: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("asset path is not a zip or directory: %s", root)
	}

	index := make(map[string]string, len(requiredFiles))
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		for _, name := range requiredFiles {
			if matchesAssetName(base, name) && index[name] == "" {
				index[name] = path
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk asset dir: %w", err)
	}

	assets := make(map[string][]byte, len(requiredFiles))
	for _, name := range requiredFiles {
		path := index[name]
		if path == "" {
			return nil, fmt.Errorf("required file missing in asset dir: %s", name)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		assets[name] = data
	}
	return assets, nil
}

func disableInitGame(path string) error {
	disabledPath := path + ".disabled"
	_ = os.Remove(disabledPath)
	if err := os.Rename(path, disabledPath); err != nil {
		return fmt.Errorf("disable initgame.dat: %w", err)
	}
	return nil
}

func matchesAssetName(actual, required string) bool {
	if strings.EqualFold(actual, required) {
		return true
	}
	for _, alias := range assetAliases[required] {
		if strings.EqualFold(actual, alias) {
			return true
		}
	}
	return false
}
