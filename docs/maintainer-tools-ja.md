# 保守者向けツール整理

利用者向けの導入処理は `exult-ja-install` に集約する。
旧 `scripts/stage_bg_patch_from_zip.sh` は同じ目的なので、リポジトリには入れない。

保守者向けの調査処理は `exult-ja-maint` に寄せる。

## コマンド

U7J の非 GTK ツールをビルドする:

```sh
go run ./cmd/exult-ja-maint build-u7j-tools --src /path/to/u7j/src
```

Black Gate 用日本語素材 zip を解析する:

```sh
go run ./cmd/exult-ja-maint analyze-bg --assets /path/to/u7j-assets.zip --u7j-src /path/to/u7j/src --out /tmp/exult-ja-bg-analysis
```

## 旧 scripts の扱い

旧 `scripts/` はローカル調査用として扱い、通常は commit しない。
必要な処理は Go 製ツールへ移してからリポジトリへ入れる。

- `stage_bg_patch_from_zip.sh`
  - `exult-ja-install` で置き換え済み
- `build_u7j_non_gtk_tools.sh`
  - `exult-ja-maint build-u7j-tools` へ移行
- `analyze_bg_zip_without_extract.sh`
  - `exult-ja-maint analyze-bg` へ移行
- `append_u7j_exultmsg_ui.py`
  - UI 日本語化を再開するときに、必要なら Go へ移植する

## 注意

`exult-ja-maint` は保守者向けで、U7J 旧ツールのソースや C コンパイラが必要。
通常の利用者は使わない。
