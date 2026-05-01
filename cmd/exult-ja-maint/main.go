package main

import (
	"archive/zip"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var staticFiles = []string{
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

var staticFileAliases = map[string][]string{
	"initgame.dat": {"initgame.dat.disabled"},
	"exultmsg.txt": {"u7exultmsg.txt"},
}

type compileStep struct {
	src   string
	obj   string
	extra []string
}

type linkStep struct {
	out  string
	objs []string
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "build-u7j-tools":
		err = buildU7JTools(os.Args[2:])
	case "analyze-bg":
		err = analyzeBG(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  exult-ja-maint build-u7j-tools --src <u7j-src-dir>")
	fmt.Fprintln(os.Stderr, "  exult-ja-maint analyze-bg --assets <bg-zip> --u7j-src <u7j-src-dir> [--out <output-dir>]")
}

func buildU7JTools(args []string) error {
	fs := flag.NewFlagSet("build-u7j-tools", flag.ContinueOnError)
	srcDir := fs.String("src", "", "path to u7j source directory")
	cc := fs.String("cc", envDefault("CC", "cc"), "C compiler")
	cflags := fs.String("cflags", envDefault("CFLAGS", "-g -Wall -O"), "C compiler flags")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *srcDir == "" {
		return errors.New("--src is required")
	}

	srcAbs, err := filepath.Abs(*srcDir)
	if err != nil {
		return fmt.Errorf("resolve --src: %w", err)
	}
	if err := requireDir(srcAbs); err != nil {
		return err
	}

	flags := strings.Fields(*cflags)
	for _, step := range compileSteps() {
		argv := append([]string{"-c", "-o", step.obj}, flags...)
		argv = append(argv, step.src)
		argv = append(argv, step.extra...)
		if err := runLogged(srcAbs, "", *cc, argv...); err != nil {
			return err
		}
	}
	for _, step := range linkSteps() {
		argv := append([]string{"-o", step.out}, flags...)
		argv = append(argv, step.objs...)
		if err := runLogged(srcAbs, "", *cc, argv...); err != nil {
			return err
		}
	}

	fmt.Printf("built U7J non-GTK tools in %s\n", srcAbs)
	return nil
}

func analyzeBG(args []string) error {
	fs := flag.NewFlagSet("analyze-bg", flag.ContinueOnError)
	assets := fs.String("assets", "", "path to Black Gate Japanese asset zip")
	u7jSrc := fs.String("u7j-src", "", "path to u7j source directory")
	outDir := fs.String("out", filepath.Join(os.TempDir(), "exult-ja-bg-analysis"), "output directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *assets == "" {
		return errors.New("--assets is required")
	}
	if *u7jSrc == "" {
		return errors.New("--u7j-src is required")
	}

	assetsAbs, err := filepath.Abs(*assets)
	if err != nil {
		return fmt.Errorf("resolve --assets: %w", err)
	}
	u7jSrcAbs, err := filepath.Abs(*u7jSrc)
	if err != nil {
		return fmt.Errorf("resolve --u7j-src: %w", err)
	}
	outAbs, err := filepath.Abs(*outDir)
	if err != nil {
		return fmt.Errorf("resolve --out: %w", err)
	}

	if err := requireDir(u7jSrcAbs); err != nil {
		return err
	}
	if !toolsReady(u7jSrcAbs) {
		if err := buildU7JTools([]string{"--src", u7jSrcAbs}); err != nil {
			return err
		}
	}

	staticDir := filepath.Join(outAbs, "static")
	reportDir := filepath.Join(outAbs, "reports")
	if err := os.MkdirAll(staticDir, 0o755); err != nil {
		return fmt.Errorf("create static dir: %w", err)
	}
	if err := os.MkdirAll(reportDir, 0o755); err != nil {
		return fmt.Errorf("create reports dir: %w", err)
	}
	if err := extractStaticFiles(assetsAbs, staticDir); err != nil {
		return err
	}
	if err := ensureDisasmPatchConfig(u7jSrcAbs); err != nil {
		return err
	}

	if err := runLogged(
		reportDir,
		filepath.Join(reportDir, "u7initgame.log"),
		filepath.Join(u7jSrcAbs, exeName("u7initgame")),
		"--write-en-names",
		"--no-conv-ja-names",
		filepath.Join(staticDir, "initgame.dat"),
	); err != nil {
		return err
	}
	if err := runLogged(
		reportDir,
		filepath.Join(reportDir, "u7textflx.log"),
		filepath.Join(u7jSrcAbs, exeName("u7textflx")),
		"--write-en-text",
		"--write-dump-text",
		"--no-conv-ja-text",
		filepath.Join(staticDir, "text.flx"),
	); err != nil {
		return err
	}
	if err := runLogged(
		u7jSrcAbs,
		filepath.Join(reportDir, "u7disasm.log"),
		filepath.Join(u7jSrcAbs, exeName("u7disasm")),
		"--bg",
		"--no-bug-fix",
		"--no-src",
		"--no-dep-dump",
		"-t", filepath.Join(reportDir, "usecode.from-current-data.txt"),
		"-a", filepath.Join(reportDir, "usecode.asm"),
		filepath.Join(staticDir, "usecode"),
	); err != nil {
		return err
	}

	if err := copyFile(filepath.Join(staticDir, "exult_u7j.cfg"), filepath.Join(reportDir, "exult_u7j.cfg")); err != nil {
		return err
	}
	fmt.Printf("analysis written to %s\n", outAbs)
	return nil
}

func compileSteps() []compileStep {
	return []compileStep{
		{"u7utils.c", "u7utils.o", nil},
		{"u7opcode.c", "u7opcode.o", nil},
		{"u7opcode.c", "u7opcode_asm.o", []string{"-DUSE_EXT_CALL_OPERAND"}},
		{"u7charmap.c", "u7charmap.o", nil},
		{"u7files.c", "u7files.o", nil},
		{"u7shape_utils.c", "u7shape_utils.o", nil},
		{"u7data_utils.c", "u7data_utils.o", nil},
		{"u7config_file.c", "u7config_file.o", nil},
		{"u7asm.c", "u7asm.o", nil},
		{"u7disasm.c", "u7disasm.o", nil},
		{"u7disasm_patch.c", "u7disasm_patch.o", nil},
		{"u7initgame.c", "u7initgame.o", nil},
		{"u7textflx.c", "u7textflx.o", nil},
		{"u7make_charmap.c", "u7make_charmap.o", nil},
		{"u7exultmsg.c", "u7exultmsg.o", nil},
		{"u7bg_text.c", "u7bg_text.o", nil},
	}
}

func linkSteps() []linkStep {
	return []linkStep{
		{exeName("u7asm"), []string{"u7asm.o", "u7utils.o", "u7opcode_asm.o", "u7charmap.o"}},
		{exeName("u7disasm"), []string{"u7disasm.o", "u7disasm_patch.o", "u7utils.o", "u7opcode.o", "-lz"}},
		{exeName("u7initgame"), []string{"u7initgame.o", "u7utils.o", "u7files.o", "u7charmap.o", "u7shape_utils.o", "u7data_utils.o"}},
		{exeName("u7textflx"), []string{"u7textflx.o", "u7utils.o", "u7files.o", "u7charmap.o", "u7shape_utils.o", "u7data_utils.o"}},
		{exeName("u7make_charmap"), []string{"u7make_charmap.o", "u7utils.o", "u7charmap.o", "u7config_file.o"}},
		{exeName("u7exultmsg"), []string{"u7exultmsg.o", "u7utils.o", "u7charmap.o"}},
		{exeName("u7bg_text"), []string{"u7bg_text.o", "u7utils.o", "u7files.o", "u7charmap.o", "u7shape_utils.o", "u7data_utils.o"}},
	}
}

func toolsReady(srcDir string) bool {
	for _, name := range []string{"u7disasm", "u7textflx", "u7initgame"} {
		if _, err := os.Stat(filepath.Join(srcDir, exeName(name))); err != nil {
			return false
		}
	}
	return true
}

func extractStaticFiles(zipPath, outDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer reader.Close()

	for _, name := range staticFiles {
		var found *zip.File
		for _, f := range reader.File {
			if matchesStaticFileName(filepath.Base(f.Name), name) {
				found = f
				break
			}
		}
		if found == nil {
			return fmt.Errorf("required file missing in zip: %s", name)
		}
		rc, err := found.Open()
		if err != nil {
			return fmt.Errorf("open %s in zip: %w", name, err)
		}
		target := filepath.Join(outDir, name)
		wc, err := os.Create(target)
		if err != nil {
			_ = rc.Close()
			return fmt.Errorf("create %s: %w", target, err)
		}
		_, copyErr := io.Copy(wc, rc)
		closeReadErr := rc.Close()
		closeWriteErr := wc.Close()
		if copyErr != nil {
			return fmt.Errorf("extract %s: %w", name, copyErr)
		}
		if closeReadErr != nil {
			return fmt.Errorf("close zip entry %s: %w", name, closeReadErr)
		}
		if closeWriteErr != nil {
			return fmt.Errorf("close %s: %w", target, closeWriteErr)
		}
	}
	return nil
}

func matchesStaticFileName(actual, required string) bool {
	if strings.EqualFold(actual, required) {
		return true
	}
	for _, alias := range staticFileAliases[required] {
		if strings.EqualFold(actual, alias) {
			return true
		}
	}
	return false
}

func ensureDisasmPatchConfig(srcDir string) error {
	target := filepath.Join(srcDir, "u7disasm_patch.cfg")
	if _, err := os.Stat(target); err == nil {
		return nil
	}
	source := filepath.Join(filepath.Dir(srcDir), "64bit", "u7disasm_patch.cfg")
	if _, err := os.Stat(source); err != nil {
		return fmt.Errorf("u7disasm_patch.cfg not found at %s or %s", target, source)
	}
	return copyFile(source, target)
}

func copyFile(src, dst string) error {
	input, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer input.Close()

	output, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	if copyErr != nil {
		return fmt.Errorf("copy %s to %s: %w", src, dst, copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close %s: %w", dst, closeErr)
	}
	return nil
}

func runLogged(dir, logPath, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if logPath == "" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		logFile, err := os.Create(logPath)
		if err != nil {
			return fmt.Errorf("create log %s: %w", logPath, err)
		}
		defer logFile.Close()
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}
	if err := cmd.Run(); err != nil {
		if logPath != "" {
			return fmt.Errorf("%s failed, see %s: %w", filepath.Base(name), logPath, err)
		}
		return fmt.Errorf("%s failed: %w", filepath.Base(name), err)
	}
	return nil
}

func requireDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat %s: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", path)
	}
	return nil
}

func exeName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func envDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
