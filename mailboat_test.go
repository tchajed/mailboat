package mailboat

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/tchajed/goose/machine/filesys"

	"github.com/tchajed/mailboat/globals"
)

type MailboatSuite struct {
	suite.Suite
}

func (suite *MailboatSuite) SetupTest() {
	filesys.Fs = filesys.NewMemFs()
	filesys.Fs.Mkdir(SpoolDir)
	for uid := uint64(0); uid < NumUsers; uid++ {
		filesys.Fs.Mkdir(GetUserDir(uid))
	}
	Open()
}

func (suite *MailboatSuite) TearDownTest() {
	globals.Shutdown()
}

func TestMailboatSuite(t *testing.T) {
	suite.Run(t, new(MailboatSuite))
}

var msg1 = []byte{0x3, 0x2, 0x10}
var msg2 = []byte{0x1}

func (suite *MailboatSuite) messageContents(msgs []Message) (contents []string) {
	for i, msg := range msgs {
		for _, otherMsg := range msgs[i+1:] {
			suite.NotEqual(otherMsg, msg, "duplicate id")
		}
	}

	for _, msg := range msgs {
		contents = append(contents, msg.Contents)
	}
	return
}

func (suite *MailboatSuite) pickup(user uint64) []string {
	msgs := Pickup(user)
	Unlock(user)
	return suite.messageContents(msgs)
}

func (suite *MailboatSuite) MessagesMatch(msgs [][]byte, actual []string) {
	var msgStrings []string
	for _, msg := range msgs {
		msgStrings = append(msgStrings, string(msg))
	}
	suite.ElementsMatch(msgStrings, actual)
}

func (suite *MailboatSuite) TestDeliverPickup() {
	Deliver(0, msg1)
	Deliver(0, msg2)
	msgs := suite.pickup(0)
	suite.MessagesMatch([][]byte{msg1, msg2}, msgs)
	suite.MessagesMatch(nil, suite.pickup(1))
}

func (suite *MailboatSuite) TestDeliverPickupLargeMessages() {
	largeMsg := make([]byte, 4096+20)
	for i := range largeMsg {
		largeMsg[i] = byte(i % 256)
	}
	largeMsg2 := make([]byte, 4096*3)
	for i := range largeMsg {
		largeMsg[i] = byte(i % 256)
	}
	Deliver(0, largeMsg)
	Deliver(0, largeMsg2)
	msgs := suite.pickup(0)
	suite.MessagesMatch([][]byte{largeMsg, largeMsg2}, msgs)
}

func (suite *MailboatSuite) TestDeliverPickupMultipleUsers() {
	Deliver(0, msg1)
	Deliver(1, msg2)
	Deliver(0, msg2)
	Deliver(2, msg1)
	suite.MessagesMatch([][]byte{msg1, msg2}, suite.pickup(0))
	suite.MessagesMatch([][]byte{msg2}, suite.pickup(1))
	suite.MessagesMatch([][]byte{msg1}, suite.pickup(2))
}

func (suite *MailboatSuite) TestDeliverDelete() {
	Deliver(0, msg1)
	Deliver(0, msg2)
	Deliver(1, msg2)
	user1Messages := Pickup(1)
	Delete(1, user1Messages[0].Id)
	Unlock(1)
	suite.MessagesMatch([][]byte{msg1, msg2}, suite.pickup(0))
	suite.MessagesMatch(nil, suite.pickup(1))
}

func (suite *MailboatSuite) TestRecoverQuiescent() {
	Deliver(0, msg1)
	Deliver(0, msg2)
	suite.MessagesMatch([][]byte{msg1, msg2}, suite.pickup(0))
	globals.Shutdown()
	Recover()
	Open()
	suite.MessagesMatch([][]byte{msg1, msg2}, suite.pickup(0))
}

// Pick up is different from gomail, which only retrieves msgids
func do_bench_loop(tid int, msg string, niter int, nsmtpiter int, npopiter int) error {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for l := 0; l < niter; l++ {
		for i := 0; i < nsmtpiter; i++ {
			u := uint64(r.Int()) % NumUsers
			Deliver(u, []byte(msg))
		}
		for i := 0; i < npopiter; i++ {
			u := uint64(r.Int()) % NumUsers
			msgs := Pickup(u)
			for _, m := range msgs {
				Delete(u, m.Id)
			}
			Unlock(u)
		}
	}
	return nil
}

func TestMixedLoad(t *testing.T) {
	os.RemoveAll("/tmp/mailboat")
	os.MkdirAll("/tmp/mailboat", 0744)
	filesys.Fs = filesys.NewDirFs("/tmp/mailboat/")
	filesys.Fs.Mkdir(SpoolDir)
	for uid := uint64(0); uid < NumUsers; uid++ {
		filesys.Fs.Mkdir(GetUserDir(uid))
	}
	Open()

	nprocEnv := os.Getenv("GOMAIL_NPROC")
	if nprocEnv == "" {
		nprocEnv = "1"
	}
	nproc64, err := strconv.ParseInt(nprocEnv, 10, 64)
	if err != nil {
		t.Fatal(err)
	}

	niterEnv := os.Getenv("GOMAIL_NITER")
	if niterEnv == "" {
		niterEnv = "1000"
	}
	niter64, err := strconv.ParseInt(niterEnv, 10, 64)
	if err != nil {
		t.Fatal(err)
	}

	nproc := int(nproc64)
	niter := int(niter64)

	var wg sync.WaitGroup
	start := time.Now()
	wg.Add(nproc)
	for g := 0; g < nproc; g++ {
		go func(g int) {
			defer wg.Done()
			err := do_bench_loop(g, "Hello world.", niter, 1, 1)
			if err != nil {
				t.Fatal(err)
			}
		}(g)
	}
	wg.Wait()

	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Printf("%d threads, %d iter, %v elapsed\n", nproc, niter, elapsed)
}
