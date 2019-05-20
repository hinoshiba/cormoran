package main

import (
	"os"
	"io"
	"io/ioutil"
	"fmt"
	"log"
	"flag"
	"net/mail"
	"path/filepath"
	"mime"
	"mime/multipart"
	"time"
	"context"
	"syscall"
	"strings"
	"encoding/base64"
	"golang.org/x/crypto/ssh/terminal"
)

import (
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-imap"
)

type File struct {
	name string
	data []byte
}

func (self *File) Write(dir string) error {
	path := filepath.Join(dir, self.name)
	if err := ioutil.WriteFile(path, self.data, 0644); err != nil {
		debugLog(err)
		return err
	}
	debugLog("wrote file", self.name)
	return nil
}

const (
	RABBIT string = "usage: cormoran [options... -h(help)] <imap server address:port> [target folder]"
)

var (
	Limit      int
	Sleep      time.Duration
	Mode       string
	ExportPath string

	Account    string
	Passwd     string

	Server     string
	Folder     string

	Debug      bool
)

func die(s string, msg ...interface{}) {
	fmt.Fprintf(os.Stderr, s + "\n" , msg...)
	os.Exit(1)
}

func debugLog(msg ...interface{}) {
	if !Debug {
		return
	}
	log.Println(msg...)
}

func cormoran() error {
	ctx, own_done := context.WithCancel(context.Background())
	defer own_done()
	debugLog("start cormoran")

	if err := setPassword(); err != nil {
		return err
	}

	c, err := dialImap()
	if err != nil {
		return err
	}
	debugLog("success imap connection")

	if err := c.Login(Account, Passwd); err != nil {
		return err
	}
	debugLog("success login :", Account)

	if Folder == "" {
		debugLog("udefined target folder. \nswitch to folders print mode.")
		if err := echoMboxs(ctx, c); err != nil {
			return nil
		}
		fmt.Printf("please select folder and try again with choose folder name.\n")
		return nil
	}

	mb, err := c.Select(Folder, false)
	if err != nil {
		return err
	}
	debugLog("selected folder :", Folder)

	debugLog("target is ", mb.Messages)
	seqset := allSeqset(mb)
	mos := make(chan *imap.Message)
	done := make(chan error)
	go func() {
		ctxc, _ := context.WithCancel(ctx)
		select {
			case done <- c.Fetch(seqset, []string{"BODY[]"}, mos):
				return
			case <- ctxc.Done():
				done <- nil
				return
		}
	}()

	p_cnt := 0
	for mo := range mos {
		p_cnt++

		if Limit != 0 {
			if p_cnt > Limit {
				debugLog("limit break")
				return nil
			}
		}

		time.Sleep(Sleep)

		fmt.Printf("\nProgress : %v / %v [ ", p_cnt, mb.Messages)
		il := mo.GetBody("BODY[]")

		m, err := imapliteral2gomessage(il)
		if err != nil {
			fmt.Printf("] failed\n")
			debugLog("target failed :", il)
			continue
		}
		fmt.Printf("subject: %s ]\n", m.Header.Get("Subject"))

		fs, err := detachFiles(m)
		if err != nil {
			fmt.Println("failed detach: %s\n", err)
			continue
		}
		debugLog("attache file detect :", len(*fs))


		for _, f := range *fs {
			gf := f

			if exits(filepath.Join(ExportPath, f.name)) {
				gf.name = fmt.Sprintf("%v-%s", p_cnt, gf.name)
			}
			fmt.Printf("write : %s\n", gf.name)

			gf.Write(ExportPath)
		}
	}

	if err = <- done; err != nil {
		return err
	}

	return nil
}
func exits(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func detachFiles(m *mail.Message) (*[]File, error) {
	mediaType, params, err := mime.ParseMediaType(m.Header.Get("Content-Type"))
	if err != nil {
		return nil, nil
	}
	if !strings.HasPrefix(mediaType, "multipart/") {
		return nil, nil
	}

	var ret []File
	mr := multipart.NewReader(m.Body, params["boundary"])
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		var f File
		if p.FileName() == "" {
			continue
		}
		f.name = p.FileName()

		b, err := ioutil.ReadAll(p)
		if err != nil {
			debugLog("read failed :", f.name)
			continue
		}
		dec, err := base64.StdEncoding.DecodeString(string(b))
		if err != nil {
			debugLog("base64 decode failed :", f.name)
			continue
		}
		f.data = dec

		ret = append(ret, f)
	}
	return &ret, nil
}

func imapliteral2gomessage(l imap.Literal) (*mail.Message, error) {
		r := strings.NewReader(fmt.Sprintf("%s",l))
		m, err := mail.ReadMessage(r)
		if err != nil {
			return nil, err
		}

		return m, nil
}

func allSeqset(mb *imap.MailboxStatus) *imap.SeqSet {
	seqset := new(imap.SeqSet)

	from := uint32(1)
	to := mb.Messages
	seqset.AddRange(from, to)

	return seqset
}

func echoMboxs(ctx context.Context, c *client.Client) error {
	ctxc, _ := context.WithCancel(ctx)
	mbs := make(chan *imap.MailboxInfo)
	done := make(chan error)

	go func() {
		select {
			case done <- c.List("", "*", mbs):
				return
			case <- ctxc.Done():
				done <- nil
				return
		}
	}()

	for mb := range mbs {
		fmt.Printf("* %s\n", mb.Name)
	}

	return <- done
}

func dialImap() (*client.Client, error) {
	debugLog("dial imap server :", Server)

	ct, err := client.DialTLS(Server, nil)
	if err == nil {
		return ct, nil
	}
	debugLog("faled tls connection:", err)

	debugLog("try none tls connection")
	cp, err := client.Dial(Server)
	if err == nil {
		return cp, nil
	}
	debugLog("faled none tls connection:", err)
	return nil, err
}

func setPassword() error {
	if Account == "" {
		return nil
	}
	if Passwd != "" {
		return nil
	}

	debugLog("read account password")
	pwd, err := termHideInput("input account password >>")
	if err != nil {
		return err
	}
	Passwd = pwd
	debugLog("set a new password")
	return nil
}

func termHideInput(msg string) (string, error) {
	fmt.Printf("%s", msg)
	b, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	fmt.Printf("\n")
	return string(b), nil
}

func init() {
	var limit     int
	var account   string
	var e_path    string
	var passwd    string
	var debug     bool
	var sleep     int

	flag.IntVar(&limit, "l", 5, "download limit. 0 = unlimited")
	flag.StringVar(&account, "u", "", "authentication accountname")
	flag.StringVar(&e_path, "e", "./", "export directory")
	flag.BoolVar(&debug, "d", false, "debug mode")
	flag.StringVar(&passwd, "p", "", "authentication password. interactive if not entered")
	flag.IntVar(&sleep, "s", 500, "(millisecond) wait time per one email")
	flag.Parse()

	if flag.NArg() < 1 {
		die(RABBIT)
	}

	if flag.Arg(0) == "" {
		die(RABBIT)
	}
	Server = flag.Arg(0)
	Folder = flag.Arg(1)

	if e_path == "" {
		die("empty export path.")
	}
	ExportPath = e_path
	if limit < 0 {
		die("limit less than 0.")
	}
	Limit = limit

    if sleep < 0 {
		die("sleep less than 0.")
	}
	Sleep = time.Duration(sleep) * time.Millisecond

	Account = account
	Passwd =  passwd
	Debug = debug
}

func main() {
	if err := cormoran(); err != nil {
		die("%s", err)
	}
}
