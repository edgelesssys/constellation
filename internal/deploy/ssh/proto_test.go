package ssh

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToAndFromProtoSlice(t *testing.T) {
	assert := assert.New(t)

	DemoSSHUser1 := UserKey{
		Username:  "test-user-2",
		PublicKey: "ssh-rsa abcdefg",
	}

	DemoSSHUser2 := UserKey{
		Username:  "test-user-2",
		PublicKey: "ssh-rsa hijklmnop",
	}

	// Input usually consists of pointers (from config parsing)
	DemoSSHUsersPointers := make([]*UserKey, 0)
	DemoSSHUsersPointers = append(DemoSSHUsersPointers, &DemoSSHUser1)
	DemoSSHUsersPointers = append(DemoSSHUsersPointers, &DemoSSHUser2)

	// Expected output usually does not consist of pointers
	DemoSSHUsersNoPointers := make([]UserKey, 0)
	DemoSSHUsersNoPointers = append(DemoSSHUsersNoPointers, DemoSSHUser1)
	DemoSSHUsersNoPointers = append(DemoSSHUsersNoPointers, DemoSSHUser2)

	ToProtoArray := ToProtoSlice(DemoSSHUsersPointers)
	FromProtoArray := FromProtoSlice(ToProtoArray)

	assert.Equal(DemoSSHUsersNoPointers, FromProtoArray)
}
