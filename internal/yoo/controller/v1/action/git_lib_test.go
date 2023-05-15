package action

import (
	"fmt"
	"testing"

	"github.com/mitchellh/go-homedir"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

func TestGitLibrary(t *testing.T) {
	//privateKeyFile, err := os.Open("/Users/luohao/.ssh/id_ed25519")
	//if err != nil {
	//	t.Error(err)
	//}
	//defer privateKeyFile.Close()
	//
	//privateKeyContent, err := io.ReadAll(privateKeyFile)
	//if err != nil {
	//	t.Error(err)
	//}
	homeDir, err := homedir.Dir()
	if err != nil {
		t.Error(err)
	}
	publicKeys, err := ssh.NewPublicKeysFromFile("git", fmt.Sprintf("%s/.ssh/id_ed25519", homeDir), "")
	if err != nil {
		t.Error(err)
	}

	repo, err := git.PlainOpen("/Users/luohao/work/go/yoo")
	if err != nil {
		t.Error(err)
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		t.Error(err)
		return
	}
	refs, err := remote.List(&git.ListOptions{
		Auth: publicKeys,
	})
	if err != nil {
		t.Error(err)
		return
	}

	for _, ref := range refs {
		fmt.Println(ref.Name())
		if ref.Name() == "refs/heads/master" {
			fmt.Println("master: hash", ref.Hash())
		}
	}

	h, err := repo.ResolveRevision(plumbing.Revision("master"))
	if err != nil {
		t.Error(err)
	}

	fmt.Println("local master hash: ", h.String())

	// rev-parse

}
