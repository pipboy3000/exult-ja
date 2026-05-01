package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractStaticFilesAcceptsAliases(t *testing.T) {
	zipPath := filepath.Join(t.TempDir(), "assets.zip")
	if err := writeStaticZip(zipPath, "STATIC/initgame.dat.disabled", "STATIC/u7exultmsg.txt"); err != nil {
		t.Fatalf("writeStaticZip: %v", err)
	}

	outDir := t.TempDir()
	if err := extractStaticFiles(zipPath, outDir); err != nil {
		t.Fatalf("extractStaticFiles returned error: %v", err)
	}

	initgame, err := os.ReadFile(filepath.Join(outDir, "initgame.dat"))
	if err != nil {
		t.Fatalf("read extracted initgame.dat: %v", err)
	}
	if string(initgame) != "initgame.dat" {
		t.Fatalf("initgame.dat content = %q, want %q", initgame, "initgame.dat")
	}

	exultmsg, err := os.ReadFile(filepath.Join(outDir, "exultmsg.txt"))
	if err != nil {
		t.Fatalf("read extracted exultmsg.txt: %v", err)
	}
	if string(exultmsg) != "exultmsg.txt" {
		t.Fatalf("exultmsg.txt content = %q, want %q", exultmsg, "exultmsg.txt")
	}
}

func writeStaticZip(path, initPath, exultmsgPath string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	zw := zip.NewWriter(file)
	defer zw.Close()

	for _, name := range staticFiles {
		zipName := "STATIC/" + name
		switch name {
		case "initgame.dat":
			zipName = initPath
		case "exultmsg.txt":
			zipName = exultmsgPath
		}
		w, err := zw.Create(zipName)
		if err != nil {
			return err
		}
		if _, err := w.Write([]byte(name)); err != nil {
			return err
		}
	}
	return nil
}
