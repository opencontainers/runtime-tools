package config

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
)

const configPath = "cases.conf"

var (
	// BundleMap for config, key is the bundlename, value is the params
	BundleMap = make(map[string]string)
	configLen int
)

func init() {
	f, err := os.Open(configPath)
	if err != nil {
		logrus.Fatalf("open file %v error %v", configPath, err)
	}
	defer f.Close()

	rd := bufio.NewReader(f)
	count := 0
	for {
		line, err := rd.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}
		prefix := strings.Split(line, "=")
		caseName := strings.TrimSpace(prefix[0])
		caseArg := strings.TrimPrefix(line, caseName+"=")
		for i, arg := range splitArgs(caseArg) {
			BundleMap[caseName+strconv.FormatInt(int64(i), 10)] = arg
			count = count + 1
		}
	}
	configLen = count
}

func splitArgs(args string) []string {
	argArray := strings.Split(args, ";")
	resArray := make([]string, len(argArray))
	for count, arg := range argArray {
		resArray[count] = strings.TrimSpace(arg)
	}
	return resArray
}
