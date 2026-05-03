# stations_watcher

気象庁の震度観測点データを取得し、差分があれば GitHub Issue を作成します。

## 動作確認

```sh
$ go run ./cmd/fetch-csv -o Stations.csv
```

## Lambda 関数で使う


```sh
$ GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bootstrap ./cmd/bootstrap
$ zip lambda.zip bootstrap
```
