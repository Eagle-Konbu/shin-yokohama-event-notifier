# Discord Message Format

This document describes the Discord Webhook Embed format used in this project.

## Overview

Daily event information is sent to Discord using the Embed format for a well-formatted message.

**Note:** The actual messages are sent in Japanese, as this bot targets venues in the Shin-Yokohama area, Japan.

## Format Specification

- **Title**: 📅 新横浜 イベント情報
- **Description**: 本日のイベント情報をお知らせします。
- **Color**: Changes based on the number of venues with events
  - 0 venues: Blue (ColorBlue)
  - 1 venue: Yellow (ColorYellow)
  - 2+ venues: Red (ColorRed)

## Field Structure

Each venue is added as a Field:
- **Name**: Emoji + venue name (e.g., 🏟️ 横浜アリーナ)
- **Value**: Event list (e.g., ・**18:00〜** Event name)
- **Inline**: false

Venues with no events display "本日の予定はありません" (No schedule for today).

## Example

```json
{
  "embeds": [
    {
      "title": "📅 新横浜 イベント情報",
      "description": "本日のイベント情報をお知らせします。",
      "color": 2326507,
      "fields": [
        {
          "name": "🏟️ 横浜アリーナ",
          "value": "・**18:00〜** アーティストA ライブツアー 2026\n・**19:00〜** プロバスケットボール 公式戦",
          "inline": false
        },
        {
          "name": "⚽ 日産スタジアム",
          "value": "・**14:00〜** サッカー国際親善試合",
          "inline": false
        },
        {
          "name": "⛸️ KOSÉ新横浜スケートセンター",
          "value": "本日の予定はありません",
          "inline": false
        }
      ]
    }
  ]
}
```
