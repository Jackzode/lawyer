package checker

import (
	"encoding/json"
	"github.com/lawyer/commons/constant"
	"os"
	"path/filepath"
	"sync"

	"github.com/lawyer/pkg/dir"
)

var (
	reservedUsernameMapping = make(map[string]bool)
	reservedUsernameInit    sync.Once
)

func initReservedUsername() {
	//reservedUsernamesJsonFilePath := filepath.Join(cli.ConfigFileDir, cli.DefaultReservedUsernamesConfigFileName)
	reservedUsernamesJsonFilePath := filepath.Join("", "")
	if dir.CheckFileExist(reservedUsernamesJsonFilePath) {
		// if reserved username file exists, read it and replace configuration
		reservedUsernamesJsonFile, err := os.ReadFile(reservedUsernamesJsonFilePath)
		if err == nil {
			constant.ReservedUsernames = reservedUsernamesJsonFile
		}
	}
	var usernames []string
	_ = json.Unmarshal(constant.ReservedUsernames, &usernames)
	for _, username := range usernames {
		reservedUsernameMapping[username] = true
	}
}

// IsReservedUsername checks whether the username is reserved
func IsReservedUsername(username string) bool {
	reservedUsernameInit.Do(initReservedUsername)
	return reservedUsernameMapping[username]
}
