package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	git "github.com/go-git/go-git/v5"
	github "github.com/google/go-github/v33/github"
)

var (
	outputDir      = flag.String("out", "./out", "Output directory")
	domainName     = flag.String("domain", "", "Domain name; overrides DOMAIN_NAME env variable")
	githubUsername = flag.String("gh-user", "", "GitHub username; overrides GITHUB_ACTOR env variable")

	indexTemplate    = template.Must(template.ParseFiles("templates/index.html"))
	redirectTemplate = template.Must(template.ParseFiles("templates/redirect.html"))
)

func main() {
	start := time.Now()
	defer fmt.Printf("Done in %v!\n", time.Since(start))

	flag.Parse()
	if len(*domainName) == 0 {
		*domainName = os.Getenv("DOMAIN_NAME")
		if len(*domainName) == 0 {
			log.Fatal("Domain name must be specified in DOMAIN_NAME env variable")
		}
	}
	if len(*githubUsername) == 0 {
		*githubUsername = os.Getenv("GITHUB_ACTOR")
		if len(*githubUsername) == 0 {
			log.Fatal("GitHub username must be specified in GITHUB_ACTOR env variable")
		}
	}
	fmt.Printf("Got configuration [domainName=%s, githubUsername=%s]\n", *domainName, *githubUsername)

	fmt.Printf("Generating the files (at %s)...\n", *outputDir)
	if err := os.MkdirAll(*outputDir, 0777); err != nil {
		log.Fatal(err)
	}

	generateIndexFile(*githubUsername)

	for _, p := range getRepositories(*githubUsername) {
		if p.Language != nil && *p.Language == "Go" && p.Private != nil && !*p.Private {
			fmt.Printf("> Found a Go repository \"%s\". Generating paths...\n", *p.Name)
			for _, repoPath := range getRepositoryPaths(p) {
				generateRedirectFile(*domainName, *p.Name, *githubUsername, repoPath)
			}
		} else {
			fmt.Printf("> Skipping \"%s\".\n", *p.Name)
		}
	}
}

func generateIndexFile(githubUsername string) {
	filePath := path.Join(*outputDir, "index.html")
	fmt.Printf("  + %s", "index.html\n")

	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = indexTemplate.Execute(file, struct {
		GitHubUsername string
	}{
		githubUsername,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func generateRedirectFile(domainName, packageName, githubUsername, outputPath string) {
	filePath := path.Join(*outputDir, outputPath+".html")
	fmt.Printf("  + %s", outputPath+".html\n")
	if err := os.MkdirAll(path.Dir(filePath), 0777); err != nil {
		log.Fatal(err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = redirectTemplate.Execute(file, struct {
		Domain, Package, GitHubUsername string
	}{
		domainName,
		packageName,
		githubUsername,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func getRepositories(username string) []*github.Repository {
	fmt.Printf("Getting the list of repositories for user %s from GitHub... ", username)

	client := github.NewClient(nil)
	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 10},
	}

	var allRepos []*github.Repository
	for {
		repos, resp, err := client.Repositories.List(context.Background(), username, opt)
		if err != nil {
			log.Fatal(err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	fmt.Printf("Found %v repositories.\n", len(allRepos))
	return allRepos
}

func getRepositoryPaths(repo *github.Repository) []string {
	tmpDir := path.Join(*outputDir, "tmp")
	tmpRepoPath := path.Join(tmpDir, *repo.Name)
	_, err := git.PlainClone(tmpRepoPath, false, &git.CloneOptions{
		URL: *repo.CloneURL,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpRepoPath)

	return listDirs(tmpRepoPath, tmpDir)
}

// listDirs returns a list of paths with all subdirectories within a given directory.
func listDirs(curPath, tmpDir string) []string {
	var dirs []string

	err := filepath.Walk(curPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		if strings.Contains(p, ".git") {
			return nil
		}
		resultPath := strings.TrimPrefix(p, tmpDir)
		_, i := utf8.DecodeRuneInString(resultPath)
		dirs = append(dirs, strings.TrimPrefix(resultPath, resultPath[i:]))
		return nil
	})
	if err != nil {
		log.Println(err)
	}

	return dirs
}
