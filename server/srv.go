package server

import (
	"github.com/tchajed/goose/machine/filesys"
	"github.com/tchajed/mailboat"

	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"strconv"
	"strings"
)

// A go mail server that is equivalent to gomail from cspec

func nameToU(u string) (uint64, error) {
	prefix := "user"
	if !strings.HasPrefix(u, prefix) {
		return 0, fmt.Errorf("username %s does not start with \"user\"", u)
	}
	i, err := strconv.ParseUint(u[len(prefix):], 10, 64)
	if err != nil || i >= mailboat.NumUsers {
		return 0, fmt.Errorf("username %s does not have a valid number", u)
	}
	return i, err
}

type Message struct {
	Client string
	From   string
	To     string
	Data   []string
}

func reply(c net.Conn, format string, elems ...interface{}) {
	var err error
	s := fmt.Sprintf(format, elems...)
	b := []byte(s + "\r\n")
	_, err = c.Write(b)
	if err != nil {
		log.Fatal(err)
	}
}

func (msg *Message) process_msg(tid int) error {
	fmt.Printf("process msg %v tid %v\n", msg, tid)
	var buffer bytes.Buffer
	for _, s := range msg.Data {
		buffer.WriteString(s)
	}
	b := buffer.Bytes()

	uid, err := nameToU(msg.To)
	if err != nil {
		return err
	}
	mailboat.Deliver(uid, b)
	return nil
}

func process_data(tp *textproto.Reader) (error, []string) {
	lines, err := tp.ReadDotLines()
	if err != nil {
		return err, nil
	}
	return nil, lines
}

func process_smtp(c net.Conn, tid int) {
	defer c.Close()
	reader := bufio.NewReader(c)
	tp := textproto.NewReader(reader)
	var msg *Message

	reply(c, "220 Ready")
	for {
		line, err := tp.ReadLine()
		if err != nil {
			reply(c, "500 Error")
			break
		}
		words := strings.Fields(line)
		// fmt.Printf("msg: %v\n", words)
		if len(words) <= 0 {
			break
		}
		switch words[0] {
		case "HELO":
			msg = &Message{}
			reply(c, "250 OK")
		case "EHLO":
			msg = &Message{}
			reply(c, "250 OK")
		case "MAIL":
			reply(c, "250 OK")
		case "RCPT":
			if len(words) < 3 || msg == nil {
				reply(c, "500 incorrect RCPT or no HELO")
				break
			}
			u := words[2]
			u = strings.Replace(u, "<", "", -1)
			u = strings.Replace(u, ">", "", -1)
			msg.To = u
			reply(c, "250 OK")
		case "DATA":
			reply(c, "354 Proceed with data")
			err, lines := process_data(tp)
			if err != nil || msg == nil {
				reply(c, "500 Error process_data")
				break
			}
			msg.Data = lines
			err = msg.process_msg(tid)
			if err != nil {
				reply(c, "500 Error process_msg")
				break
			} else {
				reply(c, "250 OK")
			}
		case "QUIT":
			reply(c, "221 Bye")
			break
		default:
			log.Printf("Unknown command %v", words)
			reply(c, "500 Error")
			break
		}

	}
}

func smtp() {
	conn, err := net.Listen("tcp", "localhost:2525")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	for c := 0; ; c++ {
		nc, err := conn.Accept()
		if err == nil {
			go process_smtp(nc, c)
		}
	}
}

type mailbox struct {
	u    string
	id   uint64
	msgs []mailboat.Message
}

func mkMailbox(u string) (*mailbox, error) {
	userId, err := nameToU(u)
	if err != nil {
		return nil, err
	}
	mbox := &mailbox{u: u, id: userId, msgs: mailboat.Pickup(userId)}
	return mbox, nil
}

func (mbox *mailbox) unlock() {
	mailboat.Unlock(mbox.id)
}

func (mbox *mailbox) i2msg(words []string) (*mailboat.Message, bool) {
	if len(words) < 2 {
		return nil, false
	}
	i, err := strconv.Atoi(words[1])
	if err != nil {
		return nil, false
	}
	if len(mbox.msgs) < i+1 {
		return nil, false
	}
	return &mbox.msgs[i], true
}

func (mbox *mailbox) dele(m *mailboat.Message) {
	mailboat.Delete(mbox.id, m.Id)
}

func send_data(tw *textproto.Writer, c string) bool {
	data := []byte(c)
	tw.PrintfLine("+OK")
	dwr := tw.DotWriter()
	_, err := dwr.Write(data)
	if err != nil {
		return false
	}
	err = dwr.Close()
	if err != nil {
		return false
	}
	return true
}

func process_pop(c net.Conn, tid int) {
	defer c.Close()

	reader := bufio.NewReader(c)
	tr := textproto.NewReader(reader)
	writer := bufio.NewWriter(c)
	tw := textproto.NewWriter(writer)
	var mbox *mailbox

	tw.PrintfLine("+OK")

	for {
		line, err := tr.ReadLine()
		if err != nil {
			fmt.Printf("err: reading %v\n", err)
			tw.PrintfLine("-ERR")
			break
		}

		words := strings.Fields(line)
		fmt.Printf("msg: %v\n", words)
		if len(words) <= 0 {
			tw.PrintfLine("-ERR")
			break
		}

		command := strings.ToUpper(words[0])
		switch command {
		case "CAPA":
			tw.PrintfLine("+OK")
			tw.PrintfLine("USER")
			tw.PrintfLine(".")
		case "USER":
			if len(words) < 2 {
				tw.PrintfLine("-ERR")
				break
			}
			if mbox != nil {
				tw.PrintfLine("-ERR already logged in")
				break
			}
			mbox, err = mkMailbox(words[1])
			if err != nil {
				tw.PrintfLine("-ERR readdir")
				break
			}
			tw.PrintfLine("+OK")
		case "PASS":
			if len(words) < 2 {
				tw.PrintfLine("-ERR")
				// any password is ok
				tw.PrintfLine("+OK")
			}
		case "LIST":
			if mbox == nil {
				tw.PrintfLine("-ERR readdir")
				break
			}
			tw.PrintfLine("+OK")
			for i, msg := range mbox.msgs {
				tw.PrintfLine("%d %d", i, len(msg.Contents))
			}
			tw.PrintfLine(".")
			tw.PrintfLine("+OK")
		case "RETR":
			if mbox == nil {
				tw.PrintfLine("-ERR mbox")
				break
			}
			msg, ok := mbox.i2msg(words)
			if !ok {
				tw.PrintfLine("-ERR file")
				break
			}
			ok = send_data(tw, msg.Contents)
			if !ok {
				tw.PrintfLine("-ERR data")
				return
			}
		case "DELE":
			if mbox == nil {
				tw.PrintfLine("-ERR mbox")
				return
			}

			msg, ok := mbox.i2msg(words)
			if !ok {
				tw.PrintfLine("-ERR file")
				return
			}
			mbox.dele(msg)
			tw.PrintfLine("+OK")
		case "QUIT":
			if mbox != nil {
				mbox.unlock()
			}
			tw.PrintfLine("+OK")
			return
		default:
			if mbox != nil {
				mbox.unlock()
			}
			fmt.Println("err: unknown command", command)
			tw.PrintfLine("-ERR")
			return
		}
	}
}

func pop() {
	conn, err := net.Listen("tcp", "localhost:2110")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	for c := 0; ; c++ {
		nc, err := conn.Accept()
		if err == nil {
			go process_pop(nc, c)
		}
	}
}

// Start starts accepting connections and processing SMTP and POP3 requests.
//
// Before calling Start, set up the filesystem by initializing filesys.Fs.
func Start(init bool) {
	if init {
		filesys.Fs.Mkdir(mailboat.SpoolDir)
		for uid := uint64(0); uid < mailboat.NumUsers; uid++ {
			filesys.Fs.Mkdir(mailboat.GetUserDir(uid))
		}
		mailboat.Open()
	} else {
		mailboat.Recover()
	}

	go smtp()
	pop()
}
