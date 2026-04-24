# 修正版 Exult の配布基盤整備計画（macOS + Windows）

## Summary
配布方針を `OS ごとの修正版 Exult 本体` + `共通の Go 製導入ツール` に統一する。
macOS と Windows の両方を対象にし、利用者はビルド不要、`日本語素材` と `ゲームデータ` を別途用意して 1 コマンドで導入できる状態を目標にする。
初期リリースでは zip 配布を主軸にし、notarization / code signing / installer は後工程に回す。

## Key Changes
- 配布単位を 3 つに分ける
  - macOS 用修正版 Exult 配布物
  - Windows 用修正版 Exult 配布物
  - 共通 Go 製導入ツール
- 日本語素材は同梱しない前提で固定する
  - ユーザーが別途用意した zip または展開済みディレクトリを入力にする
  - 許諾が得られるまでは翻訳素材は配布物に含めない
- 導入処理を Go に一本化する
  - sh / bat / PowerShell を分けない
  - パス処理、zip 展開、検証、退避、コピー、エラー出力を共通化する
- 配布形態は当面 zip を主軸にする
  - macOS Gatekeeper と Windows SmartScreen の警告回避手順は README に明記
  - 署名や installer 化は次段階に切り分ける

## CLI / Implementation Changes
- Go ツールは 1 コマンド型 CLI にする
  - 例: `exult-ja-install --game <blackgate-dir> --assets <zip-or-dir> --exult <exult-dir-or-app>`
  - 初期版ではサブコマンドは作らない
- 入力は zip と展開済みディレクトリの両対応
  - どちらでも同じ必須ファイル検証を行う
- Go ツールの責務
  - `blackgate` ディレクトリ確認
  - `STATIC/static` 確認
  - 日本語素材から必要ファイルを `blackgate-patch-ja` に配置
  - `initgame.dat` を自動退避または無効化
  - Exult 側の patch path / 起動方法を案内
  - 既知の制限を表示
- OS 差分の扱い
  - 共通化するのは導入ロジックのみ
  - 差分として許容するのは `exult` / `exult.exe`、`.app` 構造、既定パス候補、OS 警告文言
- リポジトリに追加する主要成果物
  - 利用者向け README
  - Go ツールのソース
  - CI からの macOS/Windows ビルド定義
  - 保守者向け最小構成メモの正式ドキュメント化

## Test Plan
- 正常系
  - macOS / Windows の両方で、ビルド済み Exult + ユーザー提供 game data + ユーザー提供日本語素材から導入できる
  - zip 入力とディレクトリ入力の両方で成功する
  - 導入後に起動し、会話・本・字幕が日本語表示される
- 異常系
  - `blackgate` パス誤り
  - `STATIC/static` 不在
  - zip 不正、必須ファイル不足
  - `initgame.dat` 退避失敗
  - Exult 配置先不正
- 回帰確認
  - 起動時クラッシュしない
  - アバターが移動できる
  - 既知制限として「会話右端がたまに少し見切れる」が README に明記されている
- 受け入れ基準
  - ビルドできないユーザーが README と Go ツールだけで導入できる
  - OS ごとのスクリプト二重保守が不要になる
  - 配布物に含まない素材が明確に区別されている

## Assumptions
- 第一対象は `Black Gate + Forge of Virtue`
- 配布主軸は zip、installer は後工程
- 日本語翻訳素材は許諾が得られるまで配布物に含めない
- Go ツールは導入専用で、Exult 本体のビルドまでは担わない
- UI 完全日本語化と会話右端の最終微調整は配布基盤整備の後工程とする
