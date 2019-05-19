cormoran
===

* cormoran はemailの添付ファイル保存ツールです
* フォルダ内に含まれる全ての添付ファイルをローカルに保存します
* もし、フォルダを指定しない場合、指定可能なフォルダ一覧を表示します

---

## 使い方
```
usage: cormoran [options... -h(ヘルプ)] <IMAPサーバアドレス:ポート> [対象フォルダ]
Usage of bin/cormoran:
  -d デバックモード
  -e string
        エクスポートディレクトリ（デフォルト : カレントディレクトリ）
  -l int
        ダウンロードリミット。0でリミッター無し（デフォルト : 5）
  -p string
        アカウントパスワード（未入力は対話式）
  -u string
        認証アカウント

```

## sample
```
$ bin/cormoran -e ./test/ -u hinoshiba@example.com imap.example.com:993 80_test/10_testdata
input account password >>

Progress : 1 / 3
load email subject: Testdata
write : 1-testdata.xml
Progress : 2 / 3
load email subject: Testdata, testdata02
write : 2-testdata.xml
write : 2-testdata02.xml
Progress : 3 / 3
load email subject: testdata02
write : 3-testdata02.xml
```
