package generate_cloud_config

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

type AutoInstallKeyboard struct {
	Layout string `yaml:"layout"`
}

type User struct {
	Name              string   `yaml:"name"`
	Passwd            string   `yaml:"passwd"`
	PrimaryGroup      string   `yaml:"primary_group,omitempty"`
	Groups            []string `yaml:"groups,omitempty"`
	LockPasswd        bool     `yaml:"lock_passwd,omitempty"`
	SshAuthorizedKeys []string `yaml:"ssh_authorized_keys,omitempty"`
	Sudo              string   `yaml:"sudo,omitempty"`
	Shell             string   `yaml:"shell,omitempty"`
}

type UserData struct {
	Hostname string `yaml:"hostname"`
	Users    []User `yaml:"users"`
}

type SSH struct {
	InstallServer  bool     `yaml:"install-server"`
	AllowPw        bool     `yaml:"allow-pw"`
	AuthorizedKeys []string `yaml:"authorized-keys,omitempty"`
}

type StorageLayoutMatch struct {
	Serial *string `yaml:"serial"`
}

type StorageLayout struct {
	Name  string             `yaml:"name"`
	Match StorageLayoutMatch `yaml:"match"`
}

type Storage struct {
	Layout StorageLayout `yaml:"layout"`
}

type AutoInstall struct {
	Version       int                 `yaml:"version"`
	Timezone      string              `yaml:"timezone"`
	Locale        string              `yaml:"locale"`
	Keyboard      AutoInstallKeyboard `yaml:"keyboard"`
	UserData      UserData            `yaml:"user-data"`
	Ssh           SSH                 `yaml:"ssh"`
	Storage       Storage             `yaml:"storage"`
	Packages      []string            `yaml:"packages"`
	EarlyCommands []string            `yaml:"early-commands"`
	LateCommands  []string            `yaml:"late-commands"`
	Shutdown      string              `yaml:"shutdown"`
}

type CloudConfig struct {
	AutoInstall AutoInstall `yaml:"autoinstall"`
}

type CloudConfigContext struct {
	Hostname         string
	AdminUsername    string
	AdminPassword    string
	RootPassword     string
	SSHKeys          []string
	DiskSerial       string
	PlexClaim        string
	CloudflaredToken string
}

func getEarlyCommands(ctx CloudConfigContext) (commands []string, err error) {
	err = fs.WalkDir(filesFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			// Create the directory in the temporary location
			return nil
		} else {
			// Open the embedded file
			srcFile, err := filesFS.Open(path)
			if err != nil {
				return fmt.Errorf("error opening embedded file %s: %w", path, err)
			}
			defer func(srcFile fs.File) {
				e := srcFile.Close()
				if e != nil {
					err = e
				}
			}(srcFile)

			pathSplit := strings.Split(path, "/")
			var pathWithoutPrefix string
			if pathSplit[0] == "files" {
				pathWithoutPrefix = "/" + strings.TrimSuffix(strings.Join(pathSplit[1:], "/"), ".tpl")
			} else {
				pathWithoutPrefix = "/" + strings.TrimSuffix(strings.Join(pathSplit[:len(pathSplit)-1], "/"), ".tpl")
			}
			dir := filepath.Dir(pathWithoutPrefix)
			baseName := pathSplit[len(pathSplit)-1]
			var contents bytes.Buffer
			if strings.HasSuffix(path, ".tpl") {
				tmpl, err := template.ParseFS(filesFS, path)
				if err != nil {
					return fmt.Errorf("error parsing template %s: %w", path, err)
				}
				if err = tmpl.ExecuteTemplate(&contents, filepath.Base(path), ctx); err != nil {
					return fmt.Errorf("error executing template %s: %w", path, err)
				}
				baseName = strings.TrimSuffix(baseName, ".tpl")
			} else {
				if content, err := filesFS.ReadFile(path); err != nil {
					return err
				} else {
					buf := bytes.NewBuffer(content)
					contents = *buf
				}
			}

			base64Content := base64.StdEncoding.EncodeToString([]byte(strings.Replace(contents.String(), "#executable\n", "", 1)))
			var outputFile string
			if baseName == "authorized_keys.tpl" {
				outputFile = fmt.Sprintf("/home/%s/.ssh/authorized_keys", ctx.AdminUsername)
				command := fmt.Sprintf(`curtin in-target -- echo "%s" | base64 -d > %s`, base64Content, "home/root/.ssh/authorized_keys")
				commands = append(commands, "curtin in-target -- mkdir -p /home/root/.ssh", command)
			} else {
				outputFile = pathWithoutPrefix
			}

			command := fmt.Sprintf(`curtin in-target -- echo "%s" | base64 -d > %s`, base64Content, outputFile)
			if !slices.Contains(commands, fmt.Sprintf("curtin in-target -- mkdir -p '%s'", dir)) {
				commands = append(commands, fmt.Sprintf("curtin in-target -- mkdir -p '%s'", dir))
			}
			commands = append(commands, command)
			if strings.HasPrefix(contents.String(), "#executable") {
				commands = append(commands, fmt.Sprintf("curtin in-target -- chmod +x %s", outputFile))
			}

			return nil
		}
	})

	return
}

func getDockerCommands() []string {
	return []string{
		"curtin in-target -- apt update",
		"curtin in-target -- chmod a+r /etc/apt/keyrings/docker.asc",
		`curtin in-target -- bash -c 'DEBIAN_FRONTEND=noninteractive apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin'`,
	}
}

func getNvidiaCommands() []string {
	return []string{
		"curtin in-target -- update-initramfs -u",
		"curtin in-target -- apt update",
		"curtin in-target -- bash -c 'DEBIAN_FRONTEND=noninteractive ubuntu-drivers install --gpgpu'",
		"curtin in-target -- apt update",
		"curtin in-target -- bash -c 'DEBIAN_FRONTEND=noninteractive apt install -y nvidia-container-toolkit'",
		"curtin in-target -- nvidia-ctk runtime configure --runtime=docker",
	}
}

func getBaseAutoinstall(ctx CloudConfigContext) (autoInstall CloudConfig, err error) {
	ssh := SSH{InstallServer: true}
	if len(ctx.SSHKeys) > 0 {
		ssh.AllowPw = false
		ssh.AuthorizedKeys = ctx.SSHKeys
	} else {
		ssh.AllowPw = true
	}

	earlyCommands, err := getEarlyCommands(ctx)
	if err != nil {
		return
	}

	lateCommands := []string{
		`curtin in-target -- sed -i 's|GRUB_CMDLINE_LINUX_DEFAULT=|GRUB_CMDLINE_LINUX_DEFAULT=\"nosplash usb-storage.quirks=2109:0715:j\" /etc/default/grub'`,
		"curtin in-target -- update-grub",
	}
	lateCommands = append(lateCommands, getDockerCommands()...)
	lateCommands = append(lateCommands, getNvidiaCommands()...)

	serialMatch := fmt.Sprintf("*%s*", ctx.DiskSerial)

	autoInstall = CloudConfig{
		AutoInstall: AutoInstall{
			Version:  1,
			Timezone: "Etc/UTC",
			Locale:   "en_US.UTF-8",
			Keyboard: AutoInstallKeyboard{Layout: "us"},
			UserData: UserData{
				Hostname: ctx.Hostname,
				Users: []User{
					{
						Name:              "root",
						Passwd:            ctx.RootPassword,
						LockPasswd:        false,
						SshAuthorizedKeys: ctx.SSHKeys,
					},
					{
						Name:              ctx.AdminUsername,
						Passwd:            ctx.AdminPassword,
						PrimaryGroup:      ctx.AdminUsername,
						Groups:            []string{"sudo"},
						LockPasswd:        false,
						SshAuthorizedKeys: ctx.SSHKeys,
						Sudo:              "ALL=(ALL) NOPASSWD:ALL",
						Shell:             "/bin/bash",
					},
				},
			},
			Ssh:     ssh,
			Storage: Storage{Layout: StorageLayout{Name: "lvm", Match: StorageLayoutMatch{Serial: &serialMatch}}},
			Packages: []string{
				"vim",
				"curl",
				"git",
				"htop",
				"net-tools",
				"mdadm",
				"ca-certificates",
				"ubuntu-drivers-common",
				"build-essential",
				"dkms",
				"linux-headers-generic",
			},
			EarlyCommands: earlyCommands,
			LateCommands:  lateCommands,
			Shutdown:      "reboot",
		},
	}

	return
}

func GenerateCloudConfig(ctx CloudConfigContext) (config string, err error) {
	cfg, err := getBaseAutoinstall(ctx)
	if err != nil {
		return
	}

	yamlData, err := yaml.Marshal(cfg)
	if err != nil {
		return
	}

	config = string(yamlData)
	config = "#cloud-config\n" + config
	return
}

func WriteCloudConfig(cfg string, outputPath string) error {
	dir := filepath.Dir(outputPath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating output directory %s: %v", dir, err)
	}

	if f, err := os.Create(outputPath); err != nil {
		return fmt.Errorf("error creating output file %s: %v", outputPath, err)
	} else {
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				panic(fmt.Errorf("error closing output file %s: %v", outputPath, err))
			}
		}(f)
		if _, err = f.WriteString(cfg); err != nil {
			return fmt.Errorf("error writing cloud-config file %s: %v", outputPath, err)
		}
	}

	return nil
}
