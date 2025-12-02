package builder

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"
)

type BuildContext struct {
	Hostname         string
	AdminUsername    string
	AdminPassword    string
	RootPassword     string
	SSHKeys          []string
	Type             string
	DiskSerial       string
	PlexClaim        string
	CloudflaredToken string
}

type Templates struct {
	DockerCompose   bytes.Buffer
	FirstBootScript bytes.Buffer
	Hostname        bytes.Buffer
	AuthorizedKeys  bytes.Buffer
}

var dockerPath = []string{"os", "mkosi.extra", "opt", "containers"}

const dockerComposeFilename = "compose.yml"

var postInstallPath = []string{"os", "mkosi.extra", "opt", "postinstall"}

const firstBootScriptFilename = "first-boot.sh"

var hostnamePath = []string{"os", "mkhost.extra", "etc"}

const hostnameFilename = "hostname"

func joinPath(paths []string) string {
	return strings.Join(paths, string(os.PathSeparator))
}

func generateTemplates(ctx *BuildContext) (templates *Templates, err error) {
	tmpl := template.New("templates")

	tmpl = template.Must(tmpl.ParseFS(templateFs, "templates/**/*.tpl"))

	var dockerCompose bytes.Buffer
	err = tmpl.ExecuteTemplate(&dockerCompose, "compose.yml.tpl", ctx)
	if err != nil {
		return
	}

	var firstBootScript bytes.Buffer
	err = tmpl.ExecuteTemplate(&firstBootScript, "first-boot.sh.tpl", ctx)
	if err != nil {
		return
	}

	var hostname bytes.Buffer
	err = tmpl.ExecuteTemplate(&hostname, "hostname.tpl", ctx)
	if err != nil {
		return
	}

	var sshAuthorizedKeys bytes.Buffer
	err = tmpl.ExecuteTemplate(&sshAuthorizedKeys, "authorized_keys.tpl", ctx)
	if err != nil {
		return
	}

	templates = &Templates{
		DockerCompose:   dockerCompose,
		FirstBootScript: firstBootScript,
		Hostname:        hostname,
		AuthorizedKeys:  sshAuthorizedKeys,
	}

	return
}

func makeFiles(osDir, adminUsername string, templates *Templates) (err error) {
	composeFullPath := []string{osDir}
	composeFullPath = append(composeFullPath, dockerPath...)
	composeFullPath = append(composeFullPath, dockerComposeFilename)
	composeDir := filepath.Dir(joinPath(composeFullPath))

	firstBootScriptFullPath := []string{osDir}
	firstBootScriptFullPath = append(firstBootScriptFullPath, postInstallPath...)
	firstBootScriptFullPath = append(firstBootScriptFullPath, firstBootScriptFilename)
	firstBootScriptDir := filepath.Dir(joinPath(firstBootScriptFullPath))

	hostnameFullPath := []string{osDir}
	hostnameFullPath = append(hostnameFullPath, hostnamePath...)
	hostnameFullPath = append(hostnameFullPath, hostnameFilename)
	hostnameDir := filepath.Dir(joinPath(hostnameFullPath))

	sshAuthorizedKeysFullPath := []string{osDir}
	sshAuthorizedKeysFullPath = append(sshAuthorizedKeysFullPath, "os", "mkosi.extra", "home", adminUsername, ".ssh", "authorized_keys.tpl")
	sshAuthorizedKeysDir := filepath.Dir(joinPath(sshAuthorizedKeysFullPath))

	if e := os.MkdirAll(composeDir, 0755); e != nil {
		err = e
		return
	}
	if file, createErr := os.Create(joinPath(composeFullPath)); createErr == nil {
		defer func(file *os.File) {
			e := file.Close()
			if e != nil {
				err = e
				return
			}
		}(file)
		if _, e := file.Write(templates.DockerCompose.Bytes()); e != nil {
			err = e
			return
		}
	}

	if e := os.MkdirAll(firstBootScriptDir, 0755); e != nil {
		err = e
		return
	}
	if file, e := os.Create(joinPath(firstBootScriptFullPath)); e == nil {
		if _, err = file.Write(templates.FirstBootScript.Bytes()); err != nil {
			return
		}
		defer func(file *os.File) {
			e := file.Close()
			if e != nil {
				err = e
				return
			}
		}(file)
	}

	if e := os.MkdirAll(hostnameDir, 0755); e != nil {
		err = e
		return
	}
	if file, e := os.Create(joinPath(hostnameFullPath)); e == nil {
		if _, err = file.Write(templates.Hostname.Bytes()); err != nil {
			return
		}
		defer func(file *os.File) {
			e := file.Close()
			if e != nil {
				err = e
				return
			}
		}(file)
	}

	if e := os.MkdirAll(sshAuthorizedKeysDir, 0755); e != nil {
		err = e
		return
	}
	if file, e := os.Create(joinPath(sshAuthorizedKeysFullPath)); e == nil {
		if _, err = file.Write(templates.AuthorizedKeys.Bytes()); err != nil {
			return
		}
		defer func(file *os.File) {
			e := file.Close()
			if e != nil {
				err = e
				return
			}
		}(file)
	}

	return
}

func createOsDir() (dir string, err error) {
	dir, err = copyOsFsToTempDir()
	return
}

func createInstallerDir() (dir string, err error) {
	dir, err = copyInstallerFsToTempDir()
	return
}

func runMkosiBuild() (output string, err error) {
	cmd := exec.Command("mkosi", "build")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return
	}

	err = cmd.Start()
	if err != nil {
		return
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			output += scanner.Text() + "\n"
			log.Debugf("mkosi build stdout: %s", scanner.Text())
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		output += scanner.Text() + "\n"
		for scanner.Scan() {
			log.Debugf("mkosi build stderr: %s", scanner.Text())
		}
	}()

	err = cmd.Wait()

	time.Sleep(1 * time.Second) // Give goroutines a chance to finish printing

	return
}

func buildOs(osDir string, ctx *BuildContext) (imgFile string, err error) {
	templates, err := generateTemplates(ctx)
	if err != nil {
		return
	}

	if err = makeFiles(osDir, ctx.AdminUsername, templates); err != nil {
		err = fmt.Errorf("error making templated files: %s", err.Error())
		return
	}

	if err = os.Chdir(osDir); err != nil {
		err = fmt.Errorf("error changing to directory %s: %s", osDir, err.Error())
		return
	}

	if output, e := runMkosiBuild(); e != nil {
		err = fmt.Errorf("error running mkosi build: %s", output)
		return
	}

	imgFile = joinPath([]string{osDir, "mkosi.output", "ubuntu.img"})

	return
}

func buildInstaller(installerDir string) (installerFile string, err error) {

	if err = os.Chdir(installerDir); err != nil {
		return
	}

	if output, e := runMkosiBuild(); e != nil {
		err = fmt.Errorf("error running mkosi build: %s", output)
		return
	}

	installerFile = joinPath([]string{installerDir, "mkosi.output", "installer.img"})

	return
}

func copyFile(src, dst string) (err error) {
	sourceFile, err := os.Open(src)
	if err != nil {
		return
	}
	defer func(sourceFile *os.File) {
		err := sourceFile.Close()
		if err != nil {
			return
		}
	}(sourceFile)

	destinationFile, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func(destinationFile *os.File) {
		err := destinationFile.Close()
		if err != nil {
			return
		}
	}(destinationFile)

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return
	}

	// Optionally, ensure data is flushed to disk
	err = destinationFile.Sync()
	if err != nil {
		return
	}

	return
}

func Build(ctx *BuildContext) (installerFile string, err error) {
	log.Debugln("Creating OS directory")
	osDir, err := createOsDir()
	if err != nil {
		return
	}
	log.Debugln("OS directory created")

	//defer func(path string) {
	//	log.Debugln("Cleaning up OS directory")
	//	err := os.RemoveAll(path)
	//	if err != nil {
	//		return
	//	}
	//	log.Debugln("OS directory removed")
	//}(osDir)

	log.Debugln("Creating installer directory")
	installerDir, err := createInstallerDir()
	if err != nil {
		return
	}
	log.Debugln("Installer directory created")

	log.Infoln("Building OS")
	if imgFile, buildErr := buildOs(osDir, ctx); buildErr == nil {
		log.Infoln("OS build complete")
		log.Debugln("Copying OS to installer")
		dest := joinPath([]string{installerDir, "mkosi.extra", "root", "ubuntu.img"})
		if e := copyFile(imgFile, dest); e != nil {
			err = e
			return
		}
		log.Debugln("Copied OS to installer")
		log.Infoln("Building installer")
		installerFile, err = buildInstaller(installerDir)
		log.Infoln("Installer build complete")
	} else {
		err = buildErr
		log.Errorln("Error building os", buildErr)
	}

	return
}
