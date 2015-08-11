package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"os/exec"
	"strings"
)

func main() {

}

func Fuzz(dirname string, extension string) {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		fmt.Println(err)
	}
	MakeEvents()
	EnableGFlags("vlc.exe")
	for {
		for i := range files {
			buf, err := ioutil.ReadFile(files[i].Name())
			if err != nil {
				fmt.Println(err)
			}
			mutant := MillerMutate(buf)
			random := make([]byte, 16)
			if _, err := io.ReadFull(rand.Reader, random); err != nil {
				fmt.Println(err)
			}
			filename := fmt.Sprintf("%X", random) + extension
			fuzzed, err := os.Create(filename)
			if err != nil {
				fmt.Println(err)
			}
			if _, err := fuzzed.Write(mutant); err != nil {
				fmt.Println(err)
			}
			fuzzed.Close()
			WinDbg("vlc.exe --play-and-exit" + fuzzed.Name())
			CrashHandler(filename, extension)
		}
	}
}

func MakeEvents() {
	cmd := "sxe -c \".logopen crash.log;.load msec.dll;!exploitable;.logclose crash.log;qq\" -h av\nsxe -c \".logopen crash.log;.load msec.dll;!exploitable;.logclose crash.log;qq\" -h bpe"
	if _, err := os.Stat("events.wds"); err == nil {
		if err = os.Remove("events.wds"); err != nil {
			fmt.Println(err)
		}
	}
	file, err := os.Create("events.wds")
	if err != nil {
		fmt.Println(err)
	}
	if _, err = file.WriteString(cmd); err != nil {
		fmt.Println(err)
	}
	file.Close()
}

func EnableGFlags(targetCmd string) {
	cmdStr := "gflags.exe /p enable " + targetCmd + " /full"
	cmd := exec.Command(cmdStr)
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
}

func MillerMutate(buf []byte) []byte {
	r, err := rand.Int(rand.Reader, big.NewInt(int64(math.Ceil((float64(len(buf)) / 1000)))))
	if err != nil {
		fmt.Println(err)
	}
	numWrites := r.Int64() + 1
	for i := int64(0); i < numWrites; i++ {
		r, err := rand.Int(rand.Reader, big.NewInt(256))
		randByte := byte(r.Int64())
		r, err = rand.Int(rand.Reader, big.NewInt(int64(len(buf))))
		if err != nil {
			fmt.Println(err)
		}
		randNum := r.Int64()
		buf[randNum] = randByte
	}
	return buf
}

func WinDbg(targetCmd string) {
	cmdStr := "windbg.exe -G -Q -c \"$$><events.wds;g\" " + targetCmd
	cmd := exec.Command(cmdStr)
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
}

func CrashHandler(filename string, extension string) {
	if _, err := os.Stat("crash.log"); err == nil {
		buf, err := ioutil.ReadFile("crash.log")
		if err != nil {
			fmt.Println(err)
		}
		output := string(buf)
		hash := strings.SplitAfter(output, "Hash=")[1]
		if _, err := os.Stat("blacklist.txt"); err == nil {
			buf, err = ioutil.ReadFile("blacklist.txt")
			if err != nil {
				fmt.Println(err)
			}
			blacklist := string(buf)
			if !strings.Contains(blacklist, hash) {
				if err = os.Rename(filename, hash+extension); err != nil {
					fmt.Println(err)
				}
				file, err := os.OpenFile("blacklist.txt", os.O_APPEND|os.O_WRONLY, 0600)
				if err != nil {
					fmt.Println(err)
				}
				file.WriteString(hash + "\n")
				file.Close()
			} else {
				if err = os.Remove(filename); err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}
