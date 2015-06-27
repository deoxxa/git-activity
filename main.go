package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

var (
	minWeight = flag.Int("min_occurence", 3, "minimum concurrent edit occurence")
)

func main() {
	flag.Parse()

	lstree, err := exec.Command("git", "ls-tree", "-r", "HEAD", "--name-only").Output()
	if err != nil {
		panic(err)
	}

	tree := make(map[string]bool)
	for _, v := range strings.Split(strings.TrimSpace(string(lstree)), "\n") {
		tree[v] = true
	}

	log := exec.Command("git", "log", "--name-only")

	out, err := log.StdoutPipe()
	if err != nil {
		panic(err)
	}

	br := bufio.NewReader(out)

	var (
		commitID  string
		inHeaders bool
	)

	commits := make(map[string][]string)

	go func() {
		for {
			l, err := br.ReadString('\n')

			if err == io.EOF {
				break
			} else if err != nil {
				panic(err)
			}

			l = strings.TrimRight(l, "\r\n")

			if strings.HasPrefix(l, "commit ") {
				commitID = strings.TrimPrefix(l, "commit ")
				inHeaders = true

				continue
			}

			if l == "" {
				inHeaders = false

				continue
			}

			if !strings.HasPrefix(l, "    ") && !inHeaders {
				if tree[l] {
					commits[commitID] = append(commits[commitID], l)
				}
			}
		}
	}()

	if err := log.Run(); err != nil {
		panic(err)
	}

	files := make(map[string]map[string]int)

	max := 0

	for _, v := range commits {
		for _, f1 := range v {
			if _, ok := files[f1]; !ok {
				files[f1] = make(map[string]int)
			}

			for _, f2 := range v {
				if f1 == f2 {
					continue
				}

				files[f1][f2] = files[f1][f2] + 1

				if files[f1][f2] > max {
					max = files[f1][f2]
				}
			}
		}
	}

	fmt.Printf("graph G {\n")

	done := make(map[[2]string]bool)

	for f1, v := range files {
		for f2, c := range v {
			if done[[2]string{f2, f1}] {
				continue
			}

			done[[2]string{f1, f2}] = true

			if c >= *minWeight {
				fmt.Printf("  %q -- %q [weight=%d penwidth=%f]\n", f1, f2, c, float32(c)/float32(max)*10)
			}
		}
	}

	fmt.Printf("}\n")
}
