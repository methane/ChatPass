# ChatPass

ChatWork にメッセージを投稿するためのプロキシです。
ChatPass を使うことで、直接投稿するのに比べて

* 投稿者が Token を知らなくてもいい
* Rate Limit を待ってくれる

というメリットが有ります。

## Usage

起動:

```
./chatpass -token="自分のAPI Token" -addr=":6789"
```

メッセージ送信:

```
$ curl -X POST -d "body=Hello+ChatWork" http://localhost:6789/rooms/123456/messages
```
