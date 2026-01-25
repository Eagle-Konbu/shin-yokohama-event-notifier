# 新横浜イベント通知bot

## Overview

新横浜近辺で開催されるイベント情報を収集し、Discordに通知するためのBot。

日産スタジアム、横浜アリーナ、KOSÉ新横浜スケートセンターを対象に、各公式サイトを1日1回スクレイピングし、当日のイベント情報をDiscordへ通知する。

本プロジェクトは個人利用を目的とした小規模なBotである。

---

## Target Event Venues

本Botでは、以下の施設を対象にイベント情報を取得する。

- 日産スタジアム
- 横浜アリーナ
- KOSÉ新横浜スケートセンター

---

## How It Works

1. Amazon EventBridgeにより、1日1回Lambdaを実行
2. 各対象施設の公式サイトからイベント情報をスクレイピング
3. 取得した情報を整形
4. Discord Webhookを通じて通知を送信

---

## Architecture

- Language: Go
- Runtime: AWS Lambda
- Scheduler: Amazon EventBridge
- Notification: Discord Webhook

---

## Execution Schedule

- 実行頻度: 1日1回
- 実行方式: Amazon EventBridge によるスケジュール実行

---

## Environment Variables

| Name | Description |
| ---- | ----------- |
| DISCORD_WEBHOOK_URL | Discord の Webhook URL |

---

## Notes

- スクレイピング対象サイトの構造変更により、取得に失敗する可能性があります
- 本Botは個人利用を目的としています
- 各サイトの利用規約に配慮した運用を行ってください

---

## License

MIT