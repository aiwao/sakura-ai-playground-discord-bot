# Sakura AI Engine playground discord bot
さくらのAIエンジンplaygroundを使った、Discordで動くaiチャットボット

## 環境変数
```
BOT_TOKEN: Discord BOTのトークン
OWNER_ID: オーナー限定コマンドを実行できるユーザーのID
LOAD_SESSION_DELAY: 認証情報からセッションを作成する際の遅延
CHECK_MAIL_DELAY: セッション作成時、認証コード受信メールを確認する際の遅延
MAX_SAKURA_SESSIONS: 最大さくらAIセッション数
MAX_INVALID_REQUEST_COUNT: 何回リクエストに失敗したら、そのセッションを無効化するか
```

## 実行
```
docker compose --env-file .env.local up
```
