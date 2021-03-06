// generate.go creates the data.go which is required for the --init
// run option of godev. run go generate in the cmd directory to trigger
// this script

// +build ignore

package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"
	"time"
)

const RelativePathToDataFiles = "/data/generate"

type FileContent string

type FileHash string

type Version string

type Commit string

func main() {
	CreateDataDotGo()
}

// CreateDataDotGo creates the file at ~/data.go
func CreateDataDotGo() {
	appVersion, appCommit := getRepoVersion()
	dockerfile, dockerfileHash := getFileContentsAsString("Dockerfile")
	makefile, makefileHash := getFileContentsAsString("Makefile")
	dotGitignore, dotGitignoreHash := getFileContentsAsString(".gitignore")
	dotDockerignore, dotDockerignoreHash := getFileContentsAsString(".dockerignore")
	goDotMod, goDotModHash := getFileContentsAsString("go.mod")
	mainDotGo, mainDotGoHash := getFileContentsAsString("main.go")

	if dataGoFile, err := os.Create("./data.go"); err != nil {
		panic(err)
	} else {
		defer dataGoFile.Close()
		DataDotGoTemplate.Execute(dataGoFile, struct {
			AppVersion          Version
			AppCommit           Commit
			Dockerfile          FileContent
			DockerfileHash      FileHash
			DotGitignore        FileContent
			DotGitignoreHash    FileHash
			DotDockerignore     FileContent
			DotDockerignoreHash FileHash
			GoDotMod            FileContent
			GoDotModHash        FileHash
			MainDotGo           FileContent
			MainDotGoHash       FileHash
			Makefile            FileContent
			MakefileHash        FileHash
			Timestamp           time.Time
		}{
			AppVersion:          appVersion,
			AppCommit:           appCommit,
			Dockerfile:          dockerfile,
			DockerfileHash:      dockerfileHash,
			DotDockerignore:     dotDockerignore,
			DotDockerignoreHash: dotDockerignoreHash,
			DotGitignore:        dotGitignore,
			DotGitignoreHash:    dotGitignoreHash,
			GoDotMod:            goDotMod,
			GoDotModHash:        goDotModHash,
			MainDotGo:           mainDotGo,
			MainDotGoHash:       mainDotGoHash,
			Makefile:            makefile,
			MakefileHash:        makefileHash,
			Timestamp:           time.Now(),
		})
	}
}

// getFileContentsAsString exists for brevity in the main function
// by assisting in resolving the data structure - convention over configuration!
// returns the value followed by the hash
func getFileContentsAsString(filename string) (FileContent, FileHash) {
	if contents, err := ioutil.ReadFile(resolveDataFile(filename)); err != nil {
		panic(err)
	} else {
		hash := md5.Sum(contents)
		md5hash := fmt.Sprintf("%x", hash[:])
		return FileContent(string(contents)), FileHash(md5hash)
	}
}

// getRepoVersion retrieves the semantic versioning tag and the
// sha commit hash to populate godev's --version
func getRepoVersion() (Version, Commit) {
	_, err := exec.LookPath("git")
	if err != nil {
		panic(err)
	}

	var versionOutput bytes.Buffer
	getVersion := exec.Command("git", "describe", "--tags", "--abbrev=0")
	getVersion.Stdout = &versionOutput
	getVersion.Stderr = &versionOutput
	getVersion.Run()

	var commitOutput bytes.Buffer
	getCommit := exec.Command("git", "log", "-n", "1", "--format='%H'")
	getCommit.Stdout = &commitOutput
	getCommit.Stderr = &commitOutput
	getCommit.Run()

	version := strings.Trim(versionOutput.String(), " \n")
	commit := strings.Trim(commitOutput.String(), " -'\n")[:7]
	return Version(version), Commit(commit)
}

// resolveDataFile exists to avoid listing path.Join excessively
func resolveDataFile(filename string) string {
	if cwd, err := os.Getwd(); err != nil {
		panic(err)
	} else {
		return path.Join(cwd, fmt.Sprintf("%s/%s", RelativePathToDataFiles, filename))
	}
}

const generatedFileWarning = `
// WARNING DO NOT MANUALLY EDIT - YOUR CHANGES WILL BE OVERRIDDEN
// MAKE CHANGES AT ~/app/data/generate AND RUN make generate TO REGENERATE
// THE FOLLOWING FILE
//
// GENERATED BY GO:GENERATE AT {{.Timestamp}}
//
// FILE GENERATED USING ~/app/data/generate.go
`

var DataDotGoTemplate = template.Must(template.New("test").Parse(`// > data.go
` + generatedFileWarning + `
package main

// Version is used by godev for reporting the version when installed via 'go get'
const Version = "{{.AppVersion}}"

// Commit is used by godev for reporting the version when installed via 'go get'
const Commit = "{{.AppCommit}}"

// DataDockerfile defines the 'Dockerfile' contents when --init is used
// hash:{{.DockerfileHash}}
const DataDockerfile = ` + "`" + `{{.Dockerfile}}
` + "`" + `

// DataMakefile defines the 'Makefile' contents when --init is used
// hash:{{.MakefileHash}}
const DataMakefile = ` + "`" + `{{.Makefile}}
` + "`" + `

// DataDotGitignore defines the '.gitignore' contents when --init is used
// hash:{{.DotGitignoreHash}}
const DataDotGitignore = ` + "`" + `{{.DotGitignore}}
` + "`" + `

// DataDotDockerignore defines the '.dockerignore' contents when --init is used
// hash:{{.DotDockerignoreHash}}
const DataDotDockerignore = ` + "`" + `{{.DotDockerignore}}
` + "`" + `

// DataMainDotgo defines the '.dockerignore' contents when --init is used
// hash:{{.MainDotGoHash}}
const DataMainDotgo = ` + "`" + `{{.MainDotGo}}
` + "`" + `

// DataGoDotMod defines the 'go.mod' contents when --init is used
// hash:{{.GoDotModHash}}
const DataGoDotMod = ` + "`" + `{{.GoDotMod}}
` + "`" + `

` + generatedFileWarning + `
// < data.go
`))
