package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

var repos = map[string]string{
	"cli":    "docker/cli",
	"engine": "moby/moby",
}

func getPRNumber(commit string) int {
	i := strings.Index(commit, "Merge pull request #")
	if i == -1 {
		return -1
	}

	var number int
	fmt.Sscanf(commit[i:], "Merge pull request #%d from", &number)
	return number
}

func isPR(commit string) bool {
	return strings.Contains(commit, "Merge pull request")
}

func getTitle(commit string) string {
	lines := strings.Split(commit, "\n")
	for n, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Upstream-commit:") {
			title := strings.TrimSpace(lines[n-1])
			return strings.ToUpper(title[0:1]) + title[1:]
		}
	}
	return ""
}

func getComponent(commit string) string {
	i := strings.Index(commit, "Component: ")
	if i == -1 {
		return ""
	}

	return strings.TrimSpace(commit[i+10:])
}

func main() {

	if len(os.Args) != 3 {
		log.Fatalf("usage: %s <title> <revision range>", os.Args[0])
	}

	r, err := git.PlainOpen(".")

	FromTo := strings.Split(os.Args[2], "..")

	commitsWeDontWant := make(map[string]*object.Commit)
	from, _ := r.ResolveRevision(plumbing.Revision(FromTo[0]))
	log.Printf("git log %s", from)
	cIter, err := r.Log(&git.LogOptions{From: *from})
	if err != nil {
		panic(err)
	}

	cIter.ForEach(func(c *object.Commit) error {
		firstLine := strings.Split(c.Message, "\n")[0]
		if isPR(c.Message) {
			firstLine = strings.Split(c.Message, "\n")[2]
		}
		commitsWeDontWant[firstLine] = c
		return nil
	})
	if err != nil {
		panic(err)
	}

	commits := make(map[plumbing.Hash]*object.Commit)
	to, _ := r.ResolveRevision(plumbing.Revision(FromTo[1]))
	log.Printf("git log %s", to)
	cIter, err = r.Log(&git.LogOptions{From: *to})
	if err != nil {
		panic(err)
	}
	cIter.ForEach(func(c *object.Commit) error {
		if isPR(c.Message) {

			lines := strings.Split(c.Message, "\n")
			thirdLine := lines[0]
			if len(lines) > 2 {
				thirdLine = lines[2]
			}
			_, ok2 := commitsWeDontWant[thirdLine]
			if !ok2 {
				commits[c.Hash] = c
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	changes := make(map[string][]string)

	for _, c := range commits {
		commit := c.Message

		component := getComponent(commit)
		number := getPRNumber(commit)
		title := getTitle(commit)

		dot := "*"

		if strings.HasPrefix(title, "Add") {
			dot = "+"
		} else if strings.HasPrefix(title, "Fix") {
			dot = "-"
		}

		lowerTitle := strings.ToLower(title)
		section := "---triage---"

		if strings.Contains(lowerTitle, "build") {
			section = "builder"
		}

		if strings.Contains(lowerTitle, "flag") || strings.Contains(lowerTitle, "compose") || strings.Contains(lowerTitle, "stack") {
			section = "client"
		}

		if strings.Contains(lowerTitle, "log") {
			section = "logging"
		}

		if strings.Contains(lowerTitle, "network") {
			section = "networking"
		}

		if strings.Contains(lowerTitle, "swarm") {
			section = "swarm mode"
		}

		if strings.Contains(lowerTitle, "deprecation") {
			section = "deprecation"
		}

		if strings.Contains(lowerTitle, "devmapper") || strings.Contains(lowerTitle, "aufs") {
			section = "runtime"
		}

		if strings.Contains(lowerTitle, "lcow") || strings.Contains(lowerTitle, "windows") || strings.Contains(lowerTitle, "microsoft") {
			section = "runtime"
		}

		if repo, ok := repos[component]; ok {
			changes[section] = append(changes[section], fmt.Sprintf("%s %s [%s#%d](https://github.com/%s/pull/%d)", dot, title, repo, number, repo, number))
		}

	}

	fmt.Println("# Changelog")
	fmt.Println("")
	fmt.Println("For more information on the list of deprecated flags and APIs, have a look at\nhttps://docs.docker.com/engine/deprecated/ where you can find the target removal dates")
	fmt.Println("")
	fmt.Println("##", os.Args[1])

	sections := []string{}
	for section, _ := range changes {
		sections = append(sections, section)
	}
	sort.Strings(sections)
	for _, section := range sections {
		sort.Strings(changes[section])
		fmt.Println("")
		fmt.Println("###", strings.Title(section))
		fmt.Println("")
		for _, change := range changes[section] {
			fmt.Println(change)
		}
	}
}
