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

func hashPassword(pwd string) string {
	return pwd
}

func getEarlyCommands(ctx CloudConfigContext) (commands []string, err error) {
	tmpl := template.New("files")
	eng := template.Must(tmpl.ParseFS(filesFS, "files/**/*.tpl"))

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
			pathWithoutPrefix := strings.Join(pathSplit[:len(pathSplit)-1], "/")
			dir := filepath.Dir(path)
			baseName := pathSplit[len(pathSplit)-1]
			var contents bytes.Buffer
			if strings.HasSuffix(path, ".tpl") {
				if err = eng.ExecuteTemplate(&contents, baseName, ctx); err != nil {
					return err
				}
			} else {
				if content, err := filesFS.ReadFile(path); err != nil {
					return err
				} else {
					buf := bytes.NewBuffer(content)
					contents = *buf
				}
			}

			base64Content := base64.StdEncoding.EncodeToString([]byte(strings.Replace(contents.String(), "#executable", "", 1)))
			var outputFile string
			if baseName == "authorized_keys.tpl" {
				outputFile = fmt.Sprintf("/home/%s/.ssh/authorized_keys", ctx.AdminUsername)
				command := fmt.Sprintf(`curtin in-target -- echo "%s" | base64 -d > %s`, base64Content, "home/root/.ssh/authorized_keys")
				commands = append(commands, "curtin in-target -- mkdir -p /home/root/.ssh", command)
			} else if baseName == ".bashrc" {
				outputFile = fmt.Sprintf("/home/%s/.bashrc", ctx.AdminUsername)
				command := fmt.Sprintf(`curtin in-target -- echo "%s" | base64 -d > %s`, base64Content, "home/root/.bashrc")
				if !slices.Contains(commands, "curtin in-target --mkdir -p /home/root") {
					commands = append(commands, "curtin in-target -- mkdir -p /home/root")
				}
				commands = append(commands, command)
			} else if baseName == ".vimrc" {
				outputFile = fmt.Sprintf("/home/%s/.vimrc", ctx.AdminUsername)
				command := fmt.Sprintf(`curtin in-target -- echo "%s" | base64 -d > %s`, base64Content, "home/root/.vimrc")
				if !slices.Contains(commands, "curtin in-target --mkdir -p /home/root") {
					commands = append(commands, "curtin in-target -- mkdir -p /home/root")
				}
				commands = append(commands, command)
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
		"curtin in-target -- install -m 0755 -d /etc/apt/keyrings",
		"curtin in-target -- curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc",
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
		"curtin in-target -- bash -c 'DEBIAN_FRONTEND=noninteractive apt install -y nvidia-container-toolkit",
		"curtin in-target -- nvidia-ctk runtime configure --runtime=docker",
	}
}

func getBaseAutoinstall(ctx CloudConfigContext) (autoInstall map[string]any, err error) {
	ssh := map[string]any{
		"install-server": true,
	}
	if len(ctx.SSHKeys) > 0 {
		ssh["allow-pw"] = false
		ssh["authorized-keys"] = ctx.SSHKeys
	} else {
		ssh["allow-pw"] = true
	}

	earlyCommands, err := getEarlyCommands(ctx)
	if err != nil {
		return
	}

	lateCommands := []string{
		`curtin in-target -- sed -i 's|GRUB_CMDLINE_LINUX_DEFAULT=|GRUB_CMDLINE_LINUX_DEFAULT=\"nosplash usb-storage.quirks=2109:0715:j\" /etc/default/grub`,
		"curtin in-target -- update-grub",
	}
	lateCommands = append(lateCommands, getDockerCommands()...)
	lateCommands = append(lateCommands, getNvidiaCommands()...)

	autoInstall = map[string]any{
		"autoinstall": map[string]any{
			"version":  1,
			"timezone": "Etc/UTC",
			"locale":   "en_US.UTF-8",
			"keyboard": map[string]any{
				"layout": "us",
			},
			"user-data": map[string]any{
				"hostname": ctx.Hostname,
				"users": []map[string]any{
					{
						"name":                "root",
						"passwd":              hashPassword(ctx.RootPassword),
						"lock_passwd":         false,
						"ssh_authorized_keys": ctx.SSHKeys,
					},
					{
						"name":                ctx.AdminUsername,
						"passwd":              hashPassword(ctx.AdminPassword),
						"primary_group":       ctx.AdminUsername,
						"groups":              []string{"sudo"},
						"lock_passwd":         false,
						"ssh_authorized_keys": ctx.SSHKeys,
						"sudo":                "ALL=(ALL) NOPASSWD:ALL",
						"shell":               "/bin/bash",
					},
				},
			},
			"ssh": ssh,
			"storage": map[string]any{
				"layout": map[string]any{
					"name": "lvm",
					"match": map[string]any{
						"serial": fmt.Sprintf("*%s*", ctx.DiskSerial),
					},
				},
			},
			"packages": []string{
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
			"early-commands": earlyCommands,
			"late-commands":  lateCommands,
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
