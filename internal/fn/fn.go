package fn

import (
	"bufio"
	"os"
	"sort"
	"strings"
)

func GetUserInputFromStdin() string {
	var lines []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	input := strings.Join(lines, "\n")
	return input
}

func GetOrderMaps(m map[string]string) map[string]string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	n := make(map[string]string)
	for _, k := range keys {
		n[k] = m[k]
	}
	return n
}
