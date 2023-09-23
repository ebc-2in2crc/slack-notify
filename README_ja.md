[English](README.md) | [日本語](README_ja.md)

# slack-notify

[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/ebc-2in2crc/slack-notify)](https://goreportcard.com/report/github.com/ebc-2in2crc/slack-notify)
[![Go Version](https://img.shields.io/github/go-mod/go-version/ebc-2in2crc/slack-notify)](https://img.shields.io/github/go-mod/go-version/ebc-2in2crc/slack-notify)
[![Version](https://img.shields.io/github/release/ebc-2in2crc/slack-notify.svg?label=version)](https://img.shields.io/github/release/ebc-2in2crc/slack-notify.svg?label=version)

`slack-notify` は、Google カレンダーに登録されたイベントを Slack に連携するコマンドおよび GitHub Action です。

## Description

`slack-notify` は次のことを行います。

- Google カレンダーからイベントを取得します
- 取得したイベントを Slack に投稿します
- Slack に投稿するメッセージは、カスタマイズ可能です

## Usage

```bash
$ slack-notify \
  -credentials ${GOOGLE_CREDENTIALS} \
  -calendar-id ${GOOGLE_CALENDAR_ID} \
  -slack-token ${SLACK_TOKEN} \
  -slack-channel-id ${SLACK_CHANNEL_ID} \
  -location Asia/Tokyo \
  -message '本日のイベントをお知らせします。' \
  -alternative-message '本日のイベントはありません。'

# ヘルプ
$ slack-notify -h
Usage of slack-notify:
  -alternative-message string
    	Specify alternative message
  -calendar-id string
    	Specify Google Calendar ID
  -credentials string
    	Specify credentials
  -credentials-file string
    	Specify credentials file
  -dry-run
    	Specify dry-run
  -event-filter-regexp string
    	Specify event filter regexp (default ".")
  -location string
    	Specify Location (default "UTC")
  -message string
    	Specify message
  -message-template-file string
        Specify custom message template file
  -slack-channel-id string
    	Specify Slack Channel ID
  -slack-token string
    	Specify Slack Access Token
  -target-date string
    	Specify targetDate date. e.g. 2020-01-01
  -timeout duration
    	Specify timeout (default 15m0s)
  -v	Show version
  -webhook string
        Specify Slack Webhook URL
```

## メッセージのカスタマイズ

デフォルトでは、Slack に投稿されるメッセージは次のテンプレートが使われます。
テンプレートについては [text/template](https://golang.org/pkg/text/template/) を参照してください。

```go
{{.Msg}}

{{range .Events -}}
• {{.Summary}}
{{end}}`
```

実際に Slack に投稿されるメッセージは、次のようになります。

```text
本日のイベントをお知らせします。

• あるイベント
• 別のイベント
• また別のイベント
```

テンプレートに渡されるデータは、次のような構造体です。

```go
type EventData struct {
    Msg    string // -message で指定されたメッセージ
    Events []*calendar.Event
}
```

メッセージをカスタマイズするには、`-message-template-file` でテンプレートファイルを指定します。

```bash
$ cat template.txt
{{.Msg}}

{{range $i, $v := .Events -}}
{{$i}}. {{$v.Summary}}
{{end}}

$ slack-notify \
  -message-template-file template.txt \
  # 省略
```

実際に Slack に投稿されるメッセージは、次のようになります。

```text
本日のイベントをお知らせします。

0. あるイベント
1. 別のイベント
2. また別のイベント
```

## Installation

### Developer

```bash
$ go install github.com/ebc-2in2crc/slack-notify/cmd/slack-notify@latest
```

### User

次の URL からダウンロードします。

- [https://github.com/ebc-2in2crc/slack-notify/releases](https://github.com/ebc-2in2crc/slack-notify/releases)

Homebrew を使うこともできます。

```bash
$ brew install ebc-2in2crc/tap/slack-notify
```

### GitHub Action

アクション `ebc-2in2crc/slack-notify@v0` は、Linux 用の slack-notify バイナリを `/usr/local/bin` にインストールします。
このアクションはインストールのみを実行します。

```yaml
jobs:
  slack-notify:
    name: slack-notify
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3
      - name: Install slack-notify
        uses: ebc-2in2crc/slack-notify@v0
      - name: Notify
        run: |-
          slack-notify -v
```

## Contribution

1. Fork this repository
2. Create your feature branch (git checkout -b my-new-feature)
3. Commit your changes (git commit -am 'Add some feature')
4. Rebase your local changes against the master branch
5. Run test suite with the `make test` command and confirm that it passes
6. Run `make fmt` and `make lint`
7. Create new Pull Request

## Licence

[MIT](https://github.com/ebc-2in2crc/slack-notify/blob/master/LICENSE)

## Author

[ebc-2in2crc](https://github.com/ebc-2in2crc)
