package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestReadAssetsFromDirAcceptsDisabledInitgame(t *testing.T) {
	root := t.TempDir()
	writeAssetSet(t, root, "initgame.dat.disabled")

	assets, err := readAssetsFromDir(root)
	if err != nil {
		t.Fatalf("readAssetsFromDir returned error: %v", err)
	}
	if len(assets["initgame.dat"]) == 0 {
		t.Fatal("initgame.dat asset was not loaded from initgame.dat.disabled")
	}
}

func TestReadAssetsFromZipAcceptsAliases(t *testing.T) {
	zipPath := filepath.Join(t.TempDir(), "assets.zip")
	if err := writeAssetZip(zipPath, "STATIC/initgame.dat.disabled", "STATIC/u7exultmsg.txt"); err != nil {
		t.Fatalf("writeAssetZip: %v", err)
	}

	assets, err := readAssetsFromZip(zipPath)
	if err != nil {
		t.Fatalf("readAssetsFromZip returned error: %v", err)
	}
	if len(assets["initgame.dat"]) == 0 {
		t.Fatal("initgame.dat asset was not loaded from initgame.dat.disabled")
	}
	if len(assets["exultmsg.txt"]) == 0 {
		t.Fatal("exultmsg.txt asset was not loaded from u7exultmsg.txt")
	}
}

func TestDisableInitGame(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "initgame.dat")
	if err := os.WriteFile(path, []byte("bad initgame"), 0o644); err != nil {
		t.Fatalf("write initgame.dat: %v", err)
	}

	if err := disableInitGame(path); err != nil {
		t.Fatalf("disableInitGame returned error: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("initgame.dat should be renamed, stat err=%v", err)
	}
	if _, err := os.Stat(path + ".disabled"); err != nil {
		t.Fatalf("initgame.dat.disabled should exist: %v", err)
	}
}

func writeAssetSet(t *testing.T, root, initName string) {
	t.Helper()
	for _, name := range requiredFiles {
		writeName := name
		if name == "initgame.dat" {
			writeName = initName
		}
		if err := os.WriteFile(filepath.Join(root, writeName), []byte(writeName), 0o644); err != nil {
			t.Fatalf("write %s: %v", writeName, err)
		}
	}
}

func writeAssetZip(path, initPath, exultmsgPath string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	zw := zip.NewWriter(file)
	defer zw.Close()

	for _, name := range requiredFiles {
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
