package version

import (
	"time"
)

const dev = "dev"

// Provisioned by ldflags
var (
	version    string
	commitHash string
	buildDate  string
)

type Info struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Hash    string `json:"hash"`
	Date    string `json:"date"`
}

func init() {
	if version == "" {
		version = dev
	}
	if commitHash == "" {
		commitHash = dev
	}
	if buildDate == "" {
		buildDate = time.Now().Format(time.RFC3339)
	}
}

func Full() *Info {
	return &Info{
		Name:    "Ops-Tool",
		Version: version,
		Hash:    commitHash,
		Date:    buildDate,
	}
}
