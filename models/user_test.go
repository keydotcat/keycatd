package models

import (
	"fmt"
	"testing"

	"github.com/keydotcat/backend/util"
)

func TestCreateUser(t *testing.T) {
	ctx := getCtx()
	vkp := VaultKeyPair{}
	u, err := NewUser(ctx, "test", "easdsa", "asdas@asdas.com", "somepass", make([]byte, 32), make([]byte, 32), vkp)
	if err != nil {
		fmt.Println(util.GetStack(err))
		t.Fatal(err)
	}
	if u.Id != "test" {
		fmt.Println(util.GetStack(err))
		t.Errorf("Invalid username: %s vs test", u.Id)
	}
	u, err = NewUser(ctx, "test", "easdsa", "asdas@asdas.com", "somepass", make([]byte, 32), make([]byte, 32), vkp)
	if err != nil {
		t.Fatal(err)
	}

}
