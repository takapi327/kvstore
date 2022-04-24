# [tendermint](https://docs.tendermint.com/master/tutorials/go-built-in.html)のチュートリアル用リポジトリ

## 立ち上げと実行

まだtendermintをgo getで取得していない場合以下を実行

```console
$ go get github.com/tendermint/tendermint@v0.35.4
```

tendermintの設定ファイルが存在していない場合、もしくは更新が必要な場合は以下を実行

```console
$ go run github.com/tendermint/tendermint/cmd/tendermint@v0.35.4 init validator --home ./tendermint-home
```

ビルドしてmy-appを生成

```console
$ go build -mod=mod -o my-app
```

my-appを実行して、アプリケーションを起動させる。

ブロックの生成が開始しされ、ログ出力に反映されたら成功。

```console
$ ./my-app -tm-home ./tendermint-home
```
