cormoran
===

* [日本語(in japanese)](README_jp.md)

---

* cormoran is local save of the email attached file.
* read and local saving all of the specified folder.
* If the folder is not specified, it will display the selectable folders name.

---

## usage
```
usage: cormoran [options... -h(help)] <imap server address:port> [target folder]
Usage of bin/cormoran:
  -d debug mode
  -e string
        export directory (default "./")
  -l int
        download limit. 0 = unlimited (default 5)
  -p string
        authentication password
  -u string
        authentication accountname

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

