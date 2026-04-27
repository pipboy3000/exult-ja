# CI 方針

GitHub Actions では、まず Go 製導入ツールを通常 CI として扱う。
Exult 本体のビルドは依存が重いため、初期段階では手動実行のワークフローに分ける。

## `installer-ci.yml`

対象:

- Go 導入ツールのテスト
- macOS / Windows / Linux 向けバイナリ作成
- ビルド成果物の artifact 化

実行タイミング:

- push
- pull request
- workflow_dispatch

## `exult-build.yml`

対象:

- macOS 用 Exult 本体のビルド
- Windows 用 Exult 本体のビルド

実行タイミング:

- workflow_dispatch

このワークフローは、依存パッケージや upstream Exult 側の変更で壊れやすい。
まずは配布検証用の足場として扱い、安定後にリリース作成へ組み込む。

## リポジトリ前提

`exult-build.yml` は、リポジトリ内に `exult-src/` が存在する前提で動く。
この `exult-src/` は workflow 実行時に `exult_repository` / `exult_ref` から checkout される。
指定先は日本語表示対応を含む修正版 Exult ソースである必要がある。

GitHub 化するときは、次の構成を基本にする。

- 修正版 Exult fork を別リポジトリに置き、`exult-build.yml` でその fork を checkout する

初期運用では、修正版 Exult fork を分ける方が本家追従と upstream PR の整理をしやすい。
