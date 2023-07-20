package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"net/http"
	"os"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"

	//"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"

	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

var (
	pull     = flag.Bool("pull", true, "automatically pull changes")
	privkey  = flag.String("privkey", "/Users/yao/.ssh/id_rsa", "location of private key used for auth")
	username = flag.String("username", "2955109968@qq.com", "username used for auth")
	password = flag.String("password", "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", "password used for auth")
)

func main() {
	myMux := http.NewServeMux()
	myMux.HandleFunc("/", getUser)
	http.ListenAndServe(":8080", myMux)
}

type User struct {
	Username string `json:"name"`
}

func getUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var user User
	err := decoder.Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
	gitPullPrepare()
}

func fatal(format string, a ...interface{}) {
	fmt.Printf(format, a...)
	os.Exit(1)
}

func parseAuthArgs() (transport.AuthMethod, error) {
	if len(*username) > 0 {
		return &githttp.BasicAuth{
			Username: *username,
			Password: *password,
		}, nil
	}
	*privkey, _ = homedir.Expand(*privkey)
	auth, err := ssh.NewPublicKeysFromFile("git", *privkey, "")
	if err != nil {
		return nil, err
	}
	return auth, nil
}

func gitPullPrepare() {
	auth, err := parseAuthArgs()
	if err != nil {
		fatal("cannot parse key: %s\n", err)
	}
	path := "."
	log.Println("Checking repository:", path)
	r, err := git.PlainOpen(path)
	if err != nil {
		fatal("cannot open repository: %s\n", err)
	}
	w, err := r.Worktree()
	if err != nil {
		fatal("cannot access repository: %s\n", err)
	}

	if *pull {
		err = gitPull(r, w, auth)
		if err != nil {
			fatal("cannot pull from repository: %s\n", err)
		}
	}
}

func gitPull(r *git.Repository, w *git.Worktree, auth transport.AuthMethod) error {
	if !gitHasRemote(r) {
		log.Println("Not pulling: no remotes configured.")
		return nil
	}

	log.Println("Pulling changes...")
	err := w.Pull(&git.PullOptions{
		RemoteName: "origin",
		Auth:       auth,
	})
	if err == transport.ErrEmptyRemoteRepository {
		return nil
	}
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

func gitHasRemote(r *git.Repository) bool {
	remotes, _ := r.Remotes()
	return len(remotes) > 0
}
