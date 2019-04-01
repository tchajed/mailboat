package mailboat

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tchajed/goose/machine/filesys"
)

type MailboatSuite struct {
	suite.Suite
}

func (suite *MailboatSuite) SetupTest() {
	filesys.Fs = filesys.NewMemFs()
	filesys.Fs.Mkdir(SpoolDir)
	for uid := uint64(0); uid < 100; uid++ {
		filesys.Fs.Mkdir(getUserDir(uid))
	}
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

func (suite *MailboatSuite) TestDeliverPickup() {
	Deliver(0, msg1)
	Deliver(0, msg2)
	msgs := suite.messageContents(Pickup(0))
	suite.ElementsMatch([][]byte{msg1, msg2}, msgs)
	suite.ElementsMatch(nil, suite.messageContents(Pickup(1)))
}

func (suite *MailboatSuite) TestDeliverPickupMultipleUsers() {
	Deliver(0, msg1)
	Deliver(1, msg2)
	Deliver(0, msg2)
	Deliver(2, msg1)
	suite.ElementsMatch([][]byte{msg1, msg2}, suite.messageContents(Pickup(0)))
	suite.ElementsMatch([][]byte{msg2}, suite.messageContents(Pickup(1)))
	suite.ElementsMatch([][]byte{msg1}, suite.messageContents(Pickup(2)))
}

func (suite *MailboatSuite) TestDeliverDelete() {
	Deliver(0, msg1)
	Deliver(0, msg2)
	Deliver(1, msg2)
	user1Messages := Pickup(1)
	Delete(1, user1Messages[0].Id)
	suite.ElementsMatch([][]byte{msg1, msg2}, suite.messageContents(Pickup(0)))
	suite.ElementsMatch(nil, suite.messageContents(Pickup(1)))
}

func (suite *MailboatSuite) TestRecoverQuiescent() {
	Deliver(0, msg1)
	Deliver(0, msg2)
	suite.ElementsMatch([][]byte{msg1, msg2}, suite.messageContents(Pickup(0)))
	Recover()
	suite.ElementsMatch([][]byte{msg1, msg2}, suite.messageContents(Pickup(0)))
}