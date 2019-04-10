package mailboat

import (
	"testing"

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
		filesys.Fs.Mkdir(getUserDir(uid))
	}

	globals.Init(NumUsers)
}

func (suite *MailboatSuite) TearDownTest() {
	globals.Shutdown()
}

func TestMailboatSuite(t *testing.T) {
	suite.Run(t, new(MailboatSuite))
}

var msg1 = []byte{0x3, 0x2, 0x10}
var msg2 = []byte{0x1}

func (suite *MailboatSuite) messageContents(msgs []Message) (contents [][]byte) {
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

func (suite *MailboatSuite) pickup(user uint64) [][]byte {
	msgs := Pickup(user)
	Unlock(user)
	return suite.messageContents(msgs)
}

func (suite *MailboatSuite) TestDeliverPickup() {
	Deliver(0, msg1)
	Deliver(0, msg2)
	msgs := suite.pickup(0)
	suite.ElementsMatch([][]byte{msg1, msg2}, msgs)
	suite.ElementsMatch(nil, suite.pickup(1))
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
	suite.ElementsMatch([][]byte{largeMsg, largeMsg2}, msgs)
}

func (suite *MailboatSuite) TestDeliverPickupMultipleUsers() {
	Deliver(0, msg1)
	Deliver(1, msg2)
	Deliver(0, msg2)
	Deliver(2, msg1)
	suite.ElementsMatch([][]byte{msg1, msg2}, suite.pickup(0))
	suite.ElementsMatch([][]byte{msg2}, suite.pickup(1))
	suite.ElementsMatch([][]byte{msg1}, suite.pickup(2))
}

func (suite *MailboatSuite) TestDeliverDelete() {
	Deliver(0, msg1)
	Deliver(0, msg2)
	Deliver(1, msg2)
	user1Messages := Pickup(1)
	Delete(1, user1Messages[0].Id)
	Unlock(1)
	suite.ElementsMatch([][]byte{msg1, msg2}, suite.pickup(0))
	suite.ElementsMatch(nil, suite.pickup(1))
}

func (suite *MailboatSuite) TestRecoverQuiescent() {
	Deliver(0, msg1)
	Deliver(0, msg2)
	suite.ElementsMatch([][]byte{msg1, msg2}, suite.pickup(0))
	globals.Shutdown()
	Recover()
	suite.ElementsMatch([][]byte{msg1, msg2}, suite.pickup(0))
}
