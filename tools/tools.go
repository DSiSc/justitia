package tools

import (
	"bytes"
	"encoding/hex"
	"errors"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
)

func HexToAddress(s string) types.Address {
	return BytesToAddress(FromHex(s))
}

// FromHex returns the bytes represented by the hexadecimal string s.
// s may be prefixed with "0x".
func FromHex(s string) []byte {
	if len(s) > 1 {
		if s[0:2] == "0x" || s[0:2] == "0X" {
			s = s[2:]
		}
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	return Hex2Bytes(s)
}

// Hex2Bytes returns the bytes represented by the hexadecimal string str.
func Hex2Bytes(str string) []byte {
	h, _ := hex.DecodeString(str)
	return h
}

func BytesToAddress(b []byte) types.Address {
	var a types.Address
	SetBytes(b, &a)
	return a
}

func SetBytes(b []byte, a *types.Address) {
	if len(b) > len(a) {
		b = b[len(b)-types.AddressLength:]
	}
	copy(a[types.AddressLength-len(b):], b)
}

// Home returns the home directory for the executing user.
//
// This uses an OS-specific method for discovering the home directory.
// An error is returned if a home directory cannot be detected.
func Home() (string, error) {
	user, err := user.Current()
	if nil == err {
		return user.HomeDir, nil
	}

	if "windows" == runtime.GOOS {
		return homeWindows()
	}

	// Unix-like system, so just assume Unix
	return homeUnix()
}

func homeUnix() (string, error) {
	// First prefer the HOME environmental variable
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	// If that fails, try the shell
	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		log.Error("sh -c eval echo ~$USER error.")
		return "", err
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		log.Error("blank output when reading home directory")
		return "", errors.New("blank output when reading home directory")
	}

	return result, nil
}

func homeWindows() (string, error) {
	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := drive + path
	if drive == "" || path == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		log.Error("Get home path error.")
		return "", errors.New("HOMEDRIVE, HOMEPATH, and USERPROFILE are blank")
	}

	return home, nil
}

func EnsureFolderExist(folderPath string) {
	_, err := os.Stat(folderPath)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(folderPath, 0755)
			if err != nil {
				log.Error("Can not create folder %s: %v", folderPath, err)
			}
		} else {
			log.Error("Can not create folder %s: %v", folderPath, err)
		}
	}
}
