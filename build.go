// +build ignore

package main

import (
	"log"
	"strings"
	"os/exec"
	"os"
	"flag"
	"runtime"
	"path/filepath"
	"fmt"
	"bytes"
	"crypto/md5"
	"io"
	"time"
	"strconv"
)

// dunno, why i am trying to build go with go. Seen in grafana project,
// looked cool, stole most of code :)

var (
	//versionRe = regexp.MustCompile(`-[0-9]{1,3}-g[0-9a-f]{5,10}`)
	goarch  string
	goos    string
	gocc    string
	cgo     bool
	pkgArch string
	version string = "v1"
	// deb & rpm does not support semver so have to handle their version a little differently
	linuxPackageVersion   string = "v1"
	linuxPackageIteration string = ""
	race                  bool
	phjsToRelease         string
	workingDir            string
	includeBuildNumber    bool     = true
	buildNumber           int      = 0
	binaries              []string = []string{"docker-dns"}
	isDev                 bool     = true
	enterprise            bool     = false
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)

	ensureGoPath()

	flag.StringVar(&goarch, "goarch", runtime.GOARCH, "GOARCH")
	flag.StringVar(&goos, "goos", runtime.GOOS, "GOOS")
	flag.StringVar(&gocc, "cc", "", "CC")
	flag.BoolVar(&cgo, "cgo-enabled", cgo, "Enable cgo")
	flag.StringVar(&pkgArch, "pkg-arch", "", "PKG ARCH")
	flag.BoolVar(&race, "race", race, "Use race detector")
	flag.BoolVar(&includeBuildNumber, "includeBuildNumber", includeBuildNumber, "IncludeBuildNumber in package name")
	flag.IntVar(&buildNumber, "buildNumber", 0, "Build number from CI system")
	flag.BoolVar(&isDev, "dev", isDev, "optimal for development, skips certain steps")
	flag.Parse()

	if flag.NArg() == 0 {
		log.Println("Usage: go run build.go build")
		return
	}

	workingDir, _ = os.Getwd()

	for _, cmd := range flag.Args() {
		switch cmd {
		case "setup":
			setup()

		case "clean":
			clean()

		case "build":
			//clean()
			for _, binary := range binaries {
				build(binary, "./cmd/"+binary, []string{})
			}

		default:
			log.Fatalf("Unknown command %q", cmd)
		}
	}

}

func setup() {
	runPrint("go", "get", "-v", "github.com/golang/dep")
	//runPrint("go", "install", "-v", "./pkg/cmd/grafana-server")
}

func clean() {
	if isDev {
		return
	}
}

func ensureGoPath() {
	if os.Getenv("GOPATH") == "" {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		gopath := filepath.Clean(filepath.Join(cwd, "../../../../"))
		log.Println("GOPATH is", gopath)
		os.Setenv("GOPATH", gopath)
	} else {
		log.Println("GOPATH is", os.Getenv("GOPATH"))
	}

}

func build(binaryName, pkg string, tags []string) {
	binary := fmt.Sprintf("./bin/%s-%s/%s", goos, goarch, binaryName)
	if isDev {
		//dont include os and arch in output path in dev environment
		binary = fmt.Sprintf("./bin/%s", binaryName)
	}

	if goos == "windows" {
		binary += ".exe"
	}

	if !isDev {
		rmr(binary, binary+".md5")
	}
	args := []string{"build", "-ldflags", ldflags()}
	if len(tags) > 0 {
		args = append(args, "-tags", strings.Join(tags, ","))
	}
	if race {
		args = append(args, "-race")
	}

	args = append(args, "-o", binary)
	args = append(args, pkg)

	if !isDev {
		setBuildEnv()
		runPrint("go", "version")
		fmt.Printf("Targeting %s/%s\n", goos, goarch)
	}

	runPrint("go", args...)

	if !isDev {
		// Create an md5 checksum of the binary, to be included in the archive for
		// automatic upgrades.
		err := md5File(binary)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func setBuildEnv() {
	os.Setenv("GOOS", goos)
	if goos == "windows" {
		// require windows >=7
		os.Setenv("CGO_CFLAGS", "-D_WIN32_WINNT=0x0601")
	}
	if goarch != "amd64" || goos != "linux" {
		// needed for all other archs
		cgo = true
	}
	if strings.HasPrefix(goarch, "armv") {
		os.Setenv("GOARCH", "arm")
		os.Setenv("GOARM", goarch[4:])
	} else {
		os.Setenv("GOARCH", goarch)
	}
	if goarch == "386" {
		os.Setenv("GO386", "387")
	}
	if cgo {
		os.Setenv("CGO_ENABLED", "1")
	}
	if gocc != "" {
		os.Setenv("CC", gocc)
	}
}

func rmr(paths ...string) {
	for _, path := range paths {
		log.Println("rm -r", path)
		os.RemoveAll(path)
	}
}

func ldflags() string {
	var b bytes.Buffer
	b.WriteString("-w")
	//b.WriteString(fmt.Sprintf(" -X main.version=%s", version))
	//b.WriteString(fmt.Sprintf(" -X main.commit=%s", getGitSha()))
	//b.WriteString(fmt.Sprintf(" -X main.buildstamp=%d", buildStamp()))
	return b.String()
}

func md5File(file string) error {
	fd, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fd.Close()

	h := md5.New()
	_, err = io.Copy(h, fd)
	if err != nil {
		return err
	}

	out, err := os.Create(file + ".md5")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "%x\n", h.Sum(nil))
	if err != nil {
		return err
	}

	return out.Close()
}

func buildStamp() int64 {
	bs, err := runError("git", "show", "-s", "--format=%ct")
	if err != nil {
		return time.Now().Unix()
	}
	s, _ := strconv.ParseInt(string(bs), 10, 64)
	return s
}

func runError(cmd string, args ...string) ([]byte, error) {
	ecmd := exec.Command(cmd, args...)
	bs, err := ecmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return bytes.TrimSpace(bs), nil
}

func getGitSha() string {
	v, err := runError("git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return "unknown-dev"
	}
	return string(v)
}

func runPrint(cmd string, args ...string) {
	log.Println(cmd, strings.Join(args, " "))
	ecmd := exec.Command(cmd, args...)
	ecmd.Stdout = os.Stdout
	ecmd.Stderr = os.Stderr
	err := ecmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
