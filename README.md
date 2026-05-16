# Exult 日本語化パッチ導入メモ

このリポジトリは、修正版 Exult と Ultima VII 用日本語素材を組み合わせて、日本語で遊べる状態を再現するための作業用リポジトリです。

現時点の方針は次のとおりです。

- Exult 本体は修正版を配布する
- ゲーム本体データは各自で用意する
- 日本語翻訳素材は各自で用意する
- 導入処理は Go 製ツールで共通化する

## 現状

現在の実装対象は `Black Gate + Forge of Virtue` です。

既知の制限:

- 会話の右端が、たまに少し見切れることがあります
- UI は完全日本語化していません

## 予定している配布形態

- macOS 用修正版 Exult.app の zip
- Windows 用修正版 Exult の zip
- 共通の導入ツール `exult-ja-install`

配布物の具体的な構成は [docs/distribution-layout-ja.md](docs/distribution-layout-ja.md) にまとめています。

## 生成物 / 配布候補

現在の配布候補は GitHub Actions の `exult-build` で生成しています。

- [exult-build workflow](https://github.com/pipboy3000/exult-ja/actions/workflows/exult-build.yml)
- [直近の成功 run](https://github.com/pipboy3000/exult-ja/actions/runs/25948908972)
- [macOS artifact](https://api.github.com/repos/pipboy3000/exult-ja/actions/artifacts/7029278643/zip)
- [Windows artifact](https://api.github.com/repos/pipboy3000/exult-ja/actions/artifacts/7029284935/zip)

Actions artifacts は GitHub へのログインが必要で、保持期限があります。恒久的な一般配布は GitHub Releases へ移す予定です。

## 現在の導入ツール

初期版の Go 製導入ツールは、次のように使う想定です。

```sh
exult-ja-install --game "/Library/Application Support/Exult/blackgate" --assets "/path/to/u7j-assets.zip"
```

Windows の例:

```powershell
.\exult-ja-install.exe --game "C:\Games\Ultima7\blackgate" --assets "C:\Users\you\Downloads\u7j-assets.zip"
```

主な仕様:

- `--assets` は zip と展開済みディレクトリの両対応
- `--patch-dir` を省略した場合は、OS ごとの既定パスか既存の `exult.cfg` 設定を優先
- `initgame.dat` は配置後に自動で `initgame.dat.disabled` へ退避
- 既存の `blackgate-patch-ja` を入力に再利用する場合も、`initgame.dat.disabled` を認識

既定の patch 配置先:

- macOS: `~/Library/Application Support/Exult/blackgate-patch-ja`
- Windows: `%LOCALAPPDATA%\Exult\blackgate-patch-ja`

既存の `exult.cfg` に `blackgate` の `<patch>` 設定がある場合は、その値を優先します。

## 導入ツールの役割

導入ツールは、ユーザーが用意した次の素材を受け取ります。

- Ultima VII: The Black Gate の game data
- 日本語素材 zip または展開済みディレクトリ

導入ツールは次を行います。

- `blackgate` ディレクトリの検証
- `STATIC/static` の検証
- `blackgate-patch-ja` の生成
- 必要ファイルの配置
- `initgame.dat` の無効化

## 参考

配布基盤の計画は [notes/exult-ja-distribution-plan-2026-04-24.md](notes/exult-ja-distribution-plan-2026-04-24.md) に保存しています。

CI 方針は [docs/ci-ja.md](docs/ci-ja.md) にまとめています。

保守者向けツールは [docs/maintainer-tools-ja.md](docs/maintainer-tools-ja.md) にまとめています。
