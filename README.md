---
description: "Sakana AI の Fugu と Sakura AI Engine の Go コーディングモデルを同一ベンチマークで比較した調査。"
date: 2026-06-23
lastmod: 2026-06-23
type: "blog"
draft: false
tags:
  - go/golang
  - llm
  - benchmark
  - sakana-ai/fugu
categories:
  - research
---

# Sakana AI Fugu：Go コーディング能力調査

## 概要

AI Agent の学習モデルで、Go 言語（golang 以下 go）のコーディング向けモデルを選定するため、[Sakana AI の Fugu](https://console.sakana.ai/models) モデル（モデル ID: `fugu` 🐡）を比較・評価した。

本調査は、別調査の「[🌸 Sakura AI Engine の比較調査](https://github.com/KEINOS/report-sakura_ai_engine-for-go)」と同じ 3 課題・同じテストで 5 回検証し、🌸 Sakura AI Engine の 10 モデルの調査結果と比較した。

AI Agent やハーネスによる会話履歴、メモリー、自己学習コンテキストの影響を避けるため、🐡 Sakana AI の OpenAI API 互換エンドポイントへ各モデルを直接リクエストし、初回の Go コード生成能力を測定した。

各検証では、モデルに同一の system prompt、user prompt、temperature: 0 を設定し、3 種類の課題をそれぞれ 1 回ずつ生成させた。回答は人手で修正せず、Go 1.26.4 環境で同じ調査を 5 回実施した。課題の作成と調査の実施には Codex を利用している。なお `fugu` は `temperature: 0` を受理するが無視する仕様であり、決定性を保証する条件ではない。

このリポジトリには、全 5 回・計 15 件の API 応答、15 件の抽出コード、検証ログを保存している。

- 調査日：2026 年 6 月 23 日
- 調査員：Codex
- 監修・編集：[KEINOS](https://github.com/KEINOS/)
  - 補助 Reviewer：Copilot、Claude、Agy、Hermes、Codex
- 実行環境：Go 1.26.4、Python 3.14.6、darwin/arm64
- API：`https://api.sakana.ai/v1/chat/completions`
- モデル：`fugu` のみ
- 契約：Standard subscription
- API 要求数：3 課題 × 5 回 = 15 件。消費トークンは合計 124,222 token（入力 17,157、出力 107,065）。1 課題・1 API 要求あたりの消費目安は約 8,300 token と言える。

## 結論

Fugu は **5 回すべてで 19/19 件の機能テストに合格**した。

各回の ParallelMapOrdered は race detector を有効にした 20 回反復試験にもすべて合格した。今回比較した範囲では、🌸 Sakura AI Engine の 10 モデルを含めて最も高い正確性と再現性を示した。

一方、各回の 3 課題の応答時間中央値をさらに 5 回で中央値化すると 202.87 秒だった。🌸 Sakura AI Engine の高品質群では、`preview/Kimi-K2.6` の 125.31 秒、`preview/Qwen3.6-35B-A3B` の 50.56 秒、`Qwen3-Coder-480B-A35B-Instruct-FP8` の 8.51 秒より遅い。Fugu は短い補完を大量・対話的に処理する用途より、数分待っても初回成功率を高めたい複雑なコード生成に向く。

Fugu は次の用途に適している。

- 複雑な Go 実装を、速度より初回成功率を優先して生成する
- 今回検証した ParallelMapOrdered に近い、`context`、panic、キャンセル、goroutine、race が絡む並行処理の実装候補を生成する
- 複数の要件を同時に満たす必要がある単発タスクで、生成後にテストや人手のレビューを行う前提の実装候補を作る
- Standard plan の範囲で、日常的ではなく時折重い課題を処理する

次の用途には適性が低い。

- IDE 補完や短い反復など、数秒の応答を必要とする作業
- 多数の独立タスクを直列に流すバッチ処理
- 実行前に固定トークン単価や所要時間を厳密に見積もる必要がある処理
- Go 1.26 の最新構文を、`go fix` なしで必ず使わせたい場合

実務では、通常の実装を `Qwen3-Coder-480B-A35B-Instruct-FP8` のような高速モデルで行い、今回の ParallelMapOrdered に近い複雑なコード生成だけ Fugu に切り替える構成が候補になる。速度より今回の課題群における一貫した正確性を重視する場合は Fugu を直接使う価値がある。既存コードのレビュー能力は本調査では検証していない。

## 評価課題

### TopKFrequent

単語の出現頻度を集計し、頻度降順・辞書順で上位 `k` 件を返す。大規模入力と小さな `k` に対する効率、入力非破壊、非 `nil` の空スライスも検証した。

### ParallelMapOrdered

順序を維持する並行 map 処理。worker 数制限、親 context のキャンセル、最初のエラーの保持、panic のエラー化、goroutine leak、race の有無を検証した。

### ParseOptions

文字列オプションを解析する。部分結果、型付きエラー、`errors.Is`、デフォルト値、重複ラベル、Go 1.26 の構文・API を検証した。

完全な入力は [`prompts/`](prompts/)、テストは [`fixtures/`](fixtures/) に保存している。🌸 Sakura AI Engine 調査のプロンプトとテストから、末尾改行を除く意味のある変更は加えていない。

## 評価方法

各回で Fugu に 3 課題を 1 回ずつ生成させ、回答を人手で修正せず、原則として次の順で検証した。

```text
go build ./...
go test -json -count=1 -timeout=15s ./...
go vet ./...
golangci-lint run --no-config --default=standard
go fix ./...
go test -race -count=1 -timeout=30s ./...
go test -run=^$ -bench=. -benchmem -count=3 ./...
```

通常テスト、vet、lint の後に一時コピーへ `go fix` を適用し、3 課題の race test と benchmark はその変更後のコードで実行した。ParallelMapOrdered の 20 回反復試験だけは、生成された各回のコードを未変更のまま race detector 付きで実行した。

🌸 Sakura AI Engine 調査との比較軸は次のとおり。

| 評価軸 | 値 |
| :----- | :-- |
| Functionality | 3 課題、計 19 テストの合格率 |
| Effectiveness | 全テストに合格した課題数 / 3 |
| Reliability | ParallelMapOrdered の合格率と 3 課題の race 成功率の平均 |
| Usability | `gofmt`、`go vet`、`golangci-lint` の成功率の平均 |
| Responsiveness | 3 課題の API 応答時間中央値 |
| Modern Syntax | コンパイル可能かつ `go fix` で変更されなかった課題数 / 3 |

Fugu はタスクに応じて基盤モデルを動的に選ぶ multi-agent system である。Chat Completions API は `temperature` を受理するが無視する。🌸 Sakura AI Engine 調査とのリクエスト形式を合わせるため `temperature: 0` は送信したが、決定性を保証する条件ではない。API の詳細は [`API_SPEC.md`](API_SPEC.md) にまとめた。

## 5 回の調査結果

| 回 | 機能 | 並行処理 | Race | 中央応答時間 | Usability | 初回から Modern | 入力 token | 出力 token |
| :-: | :--: | :------: | :--: | :----------: | :-------: | :---------------: | ----------: | ----------: |
| 1 | **19/19** | **6/6** | **3/3** | 202.87秒 | 8/9 | 2/3 | 1,674 | 26,314 |
| 2 | **19/19** | **6/6** | **3/3** | 131.17秒 | 8/9 | 1/3 | 1,177 | 21,412 |
| 3 | **19/19** | **6/6** | **3/3** | 215.45秒 | 9/9 | 2/3 | 1,674 | 25,742 |
| 4 | **19/19** | **6/6** | **3/3** | 284.21秒 | 9/9 | 2/3 | 11,455 | 16,077 |
| 5 | **19/19** | **6/6** | **3/3** | 125.13秒 | 9/9 | 2/3 | 1,177 | 17,520 |

5 回の合格数は 19 → 19 → 19 → 19 → 19 で、平均 19.0、幅 0 だった。3 課題の応答時間中央値をさらに 5 回で中央値化すると 202.87 秒である。

全生成コードは異なる SHA-256 値を持ち、同一出力の再利用ではない。それでも未変更コードの通常機能テスト、`go fix` 後コードの 3 課題の race test、未変更コードの ParallelMapOrdered 20 回反復試験は、5 回すべてで合格した。

### 品質上の軽微な差

機能上の失敗はなかったが、次の差はあった。

- 第 1 回 ParseOptions：未使用の戻り値を `errcheck` が 1 件検出
- 第 2 回 ParseOptions：`gofmt` による空白調整が必要
- 5 回中 4 回の ParseOptions：`strings.Split` が `go fix` により `strings.SplitSeq` へ変更
- 第 2・3 回の ParallelMapOrdered：worker 数の条件分岐が `min` へ変更

つまり、Fugu は正しいコードを安定して生成したが、Go 1.26 の最新 API を常に初回から選ぶわけではなかった。

## 🌸 Sakura AI Engine との比較

### 5 回の合格数と速度

| モデル | 第1回 | 第2回 | 第3回 | 第4回 | 第5回 | 平均合格数 | 合格数の幅 | 応答時間中央値 |
| :----- | :---: | :---: | :---: | :---: | :---: | :--------: | :----------: | :------------: |
| 🐡 **Sakana `fugu`** | **19/19** | **19/19** | **19/19** | **19/19** | **19/19** | **19.0** | **0** | 202.87秒 |
| 🌸 Sakura `preview/Qwen3.6-35B-A3B` | 18/19 | 18/19 | 18/19 | 18/19 | 18/19 | 18.0 | **0** | 50.56秒 |
| 🌸 Sakura `Qwen3-Coder-480B-A35B-Instruct-FP8` | 17/19 | 17/19 | 17/19 | 16/19 | 17/19 | 16.8 | 1 | **8.51秒** |
| 🌸 Sakura `preview/Kimi-K2.6` | 13/19 | 19/19 | 13/19 | 19/19 | 19/19 | 16.6 | 6 | 125.31秒 |
| 🌸 Sakura `gpt-oss-120b` | 18/19 | 7/19 | 13/19 | 18/19 | 19/19 | 15.0 | 12 | 8.60秒 |
| 🌸 Sakura `preview/Phi-4-multimodal-instruct` | 10/19 | 7/19 | 10/19 | 10/19 | 10/19 | 9.4 | 3 | 2.04秒 |

Fugu は Qwen3.6 より約 4.0 倍、Qwen3-Coder-480B より約 23.8 倍遅かった。一方、🌸 Sakura 側で完全合格した Kimi と gpt-oss は回ごとの変動があり、Fugu は 5 回とも完全合格だった。

### Pareto front

🌸 Sakura AI Engine 調査の第 5 回では、Functionality 80% 以上を品質ゲートとした Pareto front は次の 4 モデルだった。

- `gpt-oss-120b`
- `preview/Qwen3.6-35B-A3B`
- `Qwen3-Coder-480B-A35B-Instruct-FP8`
- `preview/Kimi-K2.6`

Fugu の固定トークン単価は公開されていない。Fugu は選択された基盤モデルに応じて課金され、複数 agent の場合は最上位モデルに基づく 1 つの rate が使われる。このため 🌸 Sakura の固定単価と同じ Cost 軸へ数値を入れると誤解を招く。以下は **Cost を除いた 6 軸**で、第 2〜5 回の完全な評価ベクトルを平均し、応答時間は中央値で比較した結果である。候補集合は平均 Functionality 80% 以上に限定した。🌸 Sakura 側の値は同調査の [第 2 回](https://github.com/KEINOS/report-sakura_ai_engine-for-go/blob/main/run2-summary.json)、[第 3 回](https://github.com/KEINOS/report-sakura_ai_engine-for-go/blob/main/run3-summary.json)、[第 4 回](https://github.com/KEINOS/report-sakura_ai_engine-for-go/blob/main/run4-summary.json)、[第 5 回](https://github.com/KEINOS/report-sakura_ai_engine-for-go/blob/main/run5-summary.json)の集計値から算出した。

| モデル | Functionality | Effectiveness | Reliability | Usability | Modern | 応答時間 |
| :----- | ------------: | ------------: | ----------: | --------: | -----: | -------: |
| 🐡 **Sakana `fugu`** | **100.0%** | **100.0%** | **100.0%** | **97.2%** | **58.3%** | 173.31秒 |
| 🌸 Sakura `preview/Qwen3.6-35B-A3B` | 94.7% | 66.7% | 75.0% | 88.9% | 33.3% | 50.53秒 |
| 🌸 Sakura `preview/Kimi-K2.6` | 92.1% | 91.7% | 83.3% | 91.7% | 25.0% | 114.45秒 |
| 🌸 Sakura `Qwen3-Coder-480B-A35B-Instruct-FP8` | 88.2% | 33.3% | 56.3% | **97.2%** | 0.0% | **8.50秒** |

この品質ゲート付き比較では、表の 4 モデルすべてが Pareto front に入る。Fugu より速いモデルは少なくとも一部の品質軸で劣り、Fugu より品質が高いモデルは存在しないためである。

ただし Fugu 自身も速度で大きく劣るため、他の 3 モデルを Pareto front から除外しない。`gpt-oss-120b` は 🌸 Sakura 側の第 5 回単独の品質ゲートには合格したが、第 2〜5 回の平均 Functionality が 75.0% のため、この集約比較では候補集合から除外した。

Fugu は「最速候補」ではなく、既存 Pareto front に **最高品質・最高再現性・最高待ち時間**という新しい端点を追加するモデルと解釈できる。

## 応答時間と token の変動

課題別の応答時間は TopKFrequent が 15.32〜44.49 秒、ParallelMapOrdered が 201.42〜284.21 秒、ParseOptions が 125.13〜333.51 秒だった。

第 4 回 ParallelMapOrdered だけ入力 token が 10,450 となり、他 4 回の 669 から大きく増えた。最終コードは全テストに合格したが、動的ルーティングにより待ち時間と token 使用量の予測性は低い。複雑な課題で長い client timeout が必要という公式ドキュメントの注意とも整合する。

## 制約

- 課題は 3 種類で、Go コーディング能力全体を代表するものではない
- 各回・各課題の生成は 1 回であり、標本数は 5
- 🐡 Fugu が内部で選んだ provider や基盤モデルは記録できない
- Standard plan の subscription 利用で、request ごとの実測金額は取得していない
- 🌸 Sakura の第 1 回には Reliability、Usability などの完全な内訳がないため、6 軸の集約比較は第 2〜5 回を使用した
- Modern Syntax は `go fix` がコードを変更しないことを modern syntax の代理指標としている。正しく保守可能なコードでも新しい API へ機械変換できる場合は減点され、逆に古い書き方でも `go fix` の変換対象でなければ検出できない
- 3 課題の race test は `go fix` 後の一時コピーに対して実行しているため、Reliability の race 成分は常に未変更の生成コードだけを評価した値ではない。ただし ParallelMapOrdered の 20 回反復試験は未変更コードで実施した
- コードレビュー能力は検証しておらず、並行処理についても 1 種類の課題を 5 回試した結果である。高リスクな本番実装への一般化には追加検証が必要である
- `max_completion_tokens: 10000` を送信したが、API の `completion_tokens` は最大 12,683 を記録した。Fugu の orchestration を含む usage 値について、この指定をリクエスト全体の厳密な消費上限とはみなせない
- Chat Completions を公平性のため使用したが、Sakana AI は新規統合に Responses API を推奨している

## 再実行方法

API key は `SAKANA_AI_API_KEY` に設定する。公式例の `SAKANA_API_KEY` とは変数名が異なるので注意する。

```sh
export SAKANA_AI_API_KEY="your API key"
RESEARCH_RUN=6 python3 research.py all
```

段階別にも実行できる。

```sh
RESEARCH_RUN=6 python3 research.py generate
RESEARCH_RUN=6 python3 research.py evaluate
RESEARCH_RUN=6 python3 research.py repeat
RESEARCH_RUN=6 python3 research.py summarize
```

`research.py` は Python 標準ライブラリだけを使用する。必要な外部コマンドは Go 1.26、`gofmt`、`go vet`、`golangci-lint` である。同じ調査番号に `solution.go` がある場合はキャッシュとして再利用する。

## 検証ファイル

- [`API_SPEC.md`](API_SPEC.md)：AI Agent 引き継ぎ用の Sakana API 仕様
- [`research.py`](research.py)：生成、検証、反復試験、集計ハーネス
- [`run1-summary.json`](run1-summary.json)〜[`run5-summary.json`](run5-summary.json)：各回の集計
- [`logs-run1/`](logs-run1/)〜[`logs-run5/`](logs-run5/)：要求、評価、反復ログ
- [`responses-run1/`](responses-run1/)〜[`responses-run5/`](responses-run5/)：API 生応答と未修正の生成コード
- [`prompts/`](prompts/)：3 課題の user prompt
- [`fixtures/`](fixtures/)：Go 1.26 のテストと benchmark

## 参考資料

- [Sakana Fugu - Get started](https://console.sakana.ai/get-started)
- [Sakana Fugu - Models and API fields](https://console.sakana.ai/models)
- [Sakana Fugu - Pricing](https://console.sakana.ai/pricing)
- [Sakura AI Engine Go モデル比較調査](https://github.com/KEINOS/report-sakura_ai_engine-for-go)
- [クラスメソッド：Sakana Fugu ファーストタッチ](https://dev.classmethod.jp/articles/sakana-fugu-ga-first-touch/)
