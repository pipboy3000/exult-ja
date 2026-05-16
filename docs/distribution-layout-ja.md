# 配布用ディレクトリ構成

初期リリースでは、配布物を OS ごとの zip に分ける。
ゲーム本体データと日本語翻訳素材は含めない。

## macOS

```text
exult-ja-macos/
  Exult.app
  exult-ja-install
  README.md
  licenses/
```

macOS 配布では `Exult.app` を基本形にする。
raw binary の `exult` は Homebrew などの dylib 依存を引きずるため、初期配布物にはしない。

想定する patch 配置先:

```text
~/Library/Application Support/Exult/blackgate-patch-ja
```

`exult.cfg` に既存の `blackgate` patch 設定がある場合、導入ツールはその値を優先する。

## Windows

```text
exult-ja-windows/
  Exult.exe
  exult-ja-install.exe
  README.md
  licenses/
```

想定する patch 配置先:

```text
%LOCALAPPDATA%\Exult\blackgate-patch-ja
```

`%LOCALAPPDATA%\Exult\exult.cfg` に既存の `blackgate` patch 設定がある場合、導入ツールはその値を優先する。

## 含めないもの

- Ultima VII: The Black Gate / Forge of Virtue のゲームデータ
- 日本語翻訳素材
- ユーザーの savegame / gamedat

## 導入コマンド例

macOS:

```sh
./exult-ja-install --game "/Library/Application Support/Exult/blackgate" --assets "/path/to/u7j-assets.zip"
```

Windows:

```powershell
.\exult-ja-install.exe --game "C:\Games\Ultima7\blackgate" --assets "C:\Users\you\Downloads\u7j-assets.zip"
```

## リリース作成時の確認

- `exult-ja-install` / `exult-ja-install.exe` が同梱されている
- Exult 本体は修正版である
- `README.md` が同梱されている
- 翻訳素材とゲームデータが混入していない
- 既知の制限が README に残っている

## ソース構成の注意

配布物を CI で作る場合、GitHub 上でも修正版 Exult ソースを取得できる必要がある。
`exult-build.yml` は `workflow_dispatch` の `exult_repository` / `exult_ref` で修正版 Exult ソースを checkout する。
そのため、配布用 repo とは別に修正版 Exult fork を用意しておく。
