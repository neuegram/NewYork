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
	"time"
)

func main() {
	Fuzz(os.Args[1], os.Args[2], os.Args[3])
}

func Fuzz(targetProc string, dirname string, extension string) {
	if !strings.HasSuffix(dirname, "\\") {
		dirname += "\\"
	}
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		fmt.Println(err)
	}
	MakeEvents()
	EnableGFlags(targetProc)
	for {
		for i := range files {
			buf, err := ioutil.ReadFile(dirname + files[i].Name())
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
			for {
				if _, err := os.Stat(fuzzed.Name()); err == nil {
					break
				}
			}
			WinDbg(fuzzed.Name())
			CrashHandler(filename, extension)
			CpuKill(targetProc, true)
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

func EnableGFlags(targetProc string) {
	cmd := exec.Command("cmd", "/c", "gflags /p /enable "+targetProc+" /full")
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

func WinDbg(filename string) {
	go func(filename string) {
		cmd := exec.Command("cmd", "/c", "debug.bat", filename)
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
		}
	}(filename)
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
		if err = os.Remove("crash.log"); err != nil {
			fmt.Println(err)
		}
	}
}

func TaskList(targetProc string) bool {
	cmd := exec.Command("tasklist")
	out, err := cmd.Output()
	outStr := string(out)
	if err != nil {
		fmt.Println(err)
	}
	if strings.Contains(outStr, targetProc) {
		return true
	}
	return false
}

func TaskKill(targetProc string) {
	cmd := exec.Command("taskkill", "/im", targetProc, "/f")
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
}

func CpuKill(targetProc string, launching bool) {
	cmd := exec.Command("cpu.bat", targetProc[:len(targetProc)-4])
	out, err := cmd.Output()
	outStr := string(out)
	if err != nil {
		fmt.Println(err)
	}
	if launching {
		for strings.Contains(outStr, "0.000000") || strings.Contains(outStr, "-1") {
			cmd = exec.Command("cpu.bat", targetProc[:len(targetProc)-4])
			out, err = cmd.Output()
			outStr = string(out)
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(1)
		}
	}
	cmd = exec.Command("cpu.bat", targetProc[:len(targetProc)-4])
	out, err = cmd.Output()
	outStr = string(out)
	if err != nil {
		fmt.Println(err)
	}
	if strings.Contains(outStr, "-1") {
		fmt.Printf("%s not running\n", targetProc)
	} else if strings.Contains(outStr, "0.000000") {
		TaskKill("windbg.exe")
	} else {
		CpuKill(targetProc, false)
	}
}
