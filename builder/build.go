package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/hunoz/ubuntu-iso-builder/utils"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Dependency struct {
	Name         string
	CheckCommand func() bool
	FixMessage   string
}

func commandIsInPath(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

var lbaRegex = regexp.MustCompile(".*EFI image start and size: (?P<lba>\\d+)\\s+\\*\\s+(?P<block_size>\\d+)\\s+,\\s+(?P<block_count>\\d+).*")

type ISOBuilder struct {
	cloudConfig    string
	osType         string
	version        string
	outputPath     string
	progressReader *utils.ProgressReader
}

func (b *ISOBuilder) ubuntuType() string {
	if b.osType == "server" {
		return "live-server"
	} else {
		return "desktop"
	}
}

func (b *ISOBuilder) isoUrl() string {
	name := fmt.Sprintf("ubuntu-%s-%s-amd64.iso", b.version, b.ubuntuType())
	return fmt.Sprintf("https://releases.ubuntu.com/%s/%s", b.version, name)
}

func (b *ISOBuilder) extractDir() string {
	return fmt.Sprintf("%s%s%s", b.outputPath, string(os.PathSeparator), "source-files")
}

func (b *ISOBuilder) sourceIsoPath() string {
	return fmt.Sprintf("%s%s%s", b.outputPath, string(os.PathSeparator), fmt.Sprintf("ubuntu-%s-%s-amd64.iso", b.version, b.ubuntuType()))
}

func (b *ISOBuilder) destIsoPath() string {
	return fmt.Sprintf("%s%s%s", b.outputPath, string(os.PathSeparator), "ubuntu.iso")
}

func (b *ISOBuilder) checkDependencies() bool {
	commands := []Dependency{
		{
			Name:         "xorisso",
			CheckCommand: func() bool { return commandIsInPath("xorriso") },
			FixMessage:   "xorisso needs to be installed",
		},
		{
			Name:         "7z",
			CheckCommand: func() bool { return commandIsInPath("7z") },
			FixMessage:   "7z needs to be installed",
		},
	}

	for _, command := range commands {
		if command.CheckCommand() {
			continue
		}

		log.Errorf("%s missing: %s", command.Name, command.FixMessage)

		return false
	}
	return true
}

func (b *ISOBuilder) modifyGrubForAutoinstall(content string) string {
	modified := strings.Replace(content, "---", "autoinstall ds=nocloud;s=/cdrom/nocloud/ ---", -1)

	return strings.Replace(modified, "set timeout=30", "set timeout=5", -1)
}

func (b *ISOBuilder) downloadIso() bool {
	stat, err := os.Stat(b.sourceIsoPath())
	if err == nil && stat.Size() > 0 {
		size := stat.Size()
		log.Debugf("Size: %d GB", size)
		return true
	}
	log.Infof("📥 Downloading Ubuntu %s %s ISO...", b.version, b.osType)
	log.Infof("   URL: %s", b.isoUrl())
	log.Infof("   This may take a while (typically 2-3 GB)...")

	if err = os.MkdirAll(b.extractDir(), 0755); err != nil {
		log.Fatalf("error creating directory %s: %v", b.extractDir(), err)
	}

	if err = utils.DownloadWithProgress(b.isoUrl(), b.sourceIsoPath()); err != nil {
		_ = os.RemoveAll(b.sourceIsoPath())
		log.Errorf("error downloading ISO: %v", err)
		return false
	}

	return true
}
func (b *ISOBuilder) extractIso() bool {
	fmt.Println("📦 Extracting ISO contents...")
	extractDir := b.extractDir()
	// Remove existing directory if it exists
	if _, err := os.Stat(extractDir); err == nil {
		if err := os.RemoveAll(extractDir); err != nil {
			fmt.Printf("❌ Failed to remove existing directory: %v\n", err)
			return false
		}
	}

	// Create the extraction directory
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		fmt.Printf("❌ Failed to create directory: %v\n", err)
		return false
	}

	// Run 7z extraction
	cmd := exec.Command("7z", "x", b.sourceIsoPath(), fmt.Sprintf("-o%s", extractDir), "-y")

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Extraction failed: %v\n", err)
		if len(output) > 0 {
			fmt.Printf("Output: %s\n", string(output))
		}
		return false
	}

	fmt.Println("✅ Extraction complete")
	return true
}
func (b *ISOBuilder) createAutoinstallConfigs() bool {
	var config map[string]interface{}
	if err := yaml.Unmarshal([]byte(b.cloudConfig), &b.cloudConfig); err != nil {
		log.Errorf("error parsing cloud-config: %v", err)
		return false
	}
	dump := strings.Join([]string{"#cloud-config", b.cloudConfig}, "\n")
	hostname := ((config["autoinstall"].(map[string]any))["user-data"].(map[string]any))["hostname"].(string)

	autoinstallFile := filepath.Join(b.extractDir(), "autoinstall")
	if err := os.WriteFile(autoinstallFile, []byte(dump), 0644); err == nil {
		log.Errorf("error creating autoinstall config file: %v", err)
		return false
	}

	// Method 2: Create nocloud datasource files for compatibility
	/*nocloud_dir = self.extract_dir / "nocloud"
	nocloud_dir.mkdir(exist_ok=True)*/
	noCloudDir := filepath.Join(b.extractDir(), "nocloud")
	if err := os.MkdirAll(noCloudDir, 0755); err != nil {
		log.Errorf("error creating nocloud directory: %v", err)
		return false
	}

	// Write user-data (preserve original formatting)
	userDataFile := filepath.Join(noCloudDir, "user-data")
	if err := os.WriteFile(userDataFile, []byte(dump), 0644); err != nil {
		log.Errorf("error writing user-data file: %v", err)
		return false
	}

	// Write meta-data
	metaDataFile := filepath.Join(noCloudDir, "meta-data")
	metaData := fmt.Sprintf("instance-id: %s\n", hostname)
	metaData += fmt.Sprintf("local-hostname: %s\n", hostname)
	if err := os.WriteFile(metaDataFile, []byte(metaData), 0644); err != nil {
		log.Errorf("error writing meta-data file: %v", err)
		return false
	}

	/*# Write vendor-data (required by nocloud)
	vendor_data_file = nocloud_dir / "vendor-data"
	with open(vendor_data_file, 'w') as f:
	f.write("#cloud-config\n{}\n")*/
	// Write vendor-data (required by nocloud)
	vendorDataFile := filepath.Join(noCloudDir, "vendor-data")
	if err := os.WriteFile(vendorDataFile, []byte("#cloud-config\n{}\n"), 0644); err != nil {
		log.Errorf("error writing vendor-data file: %v", err)
		return false
	}

	log.Infoln("✅ Configuration created")

	return true
}
func (b *ISOBuilder) modifyGrubConfig() bool {
	log.Infof("⚙️  Modifying boot configuration...")

	grubCfg := filepath.Join(b.extractDir(), "boot", "grub", "grub.cfg")
	grubCfgBackup := filepath.Join(b.extractDir(), "boot", "grub", "grub.cfg.backup")
	content, err := os.ReadFile(grubCfg)
	if err != nil {
		log.Errorf("error reading grub config file: %v", err)
		return false
	}

	data := string(content)

	if err = os.WriteFile(grubCfgBackup, content, 0644); err != nil {
		log.Errorf("error backing up grub config file: %v", err)
		return false
	}

	data = strings.Replace(data, "---", "autoinstall ---", -1)
	data = strings.Replace(data, "set timeout=30", "set timeout=5", -1)

	if err = os.WriteFile(grubCfg, []byte(data), 0644); err != nil {
		log.Errorf("error writing modified grub config file: %v", err)
		return false
	}

	log.Infoln("✅ Boot configuration modified")
	return true
}

func (b *ISOBuilder) buildIso() bool {
	log.Infoln("🔨 Building ISO image...")

	mbrFile := filepath.Join(b.extractDir(), "isohdpfx.bin")
	if file, err := os.ReadFile(b.sourceIsoPath()); err != nil {
		log.Errorf("error reading source ISO file: %v", err)
	} else {
		if err = os.WriteFile(mbrFile, file, 0644); err != nil {
			log.Warnf("error writing MBR file: %v", err)
		}
	}

	efiImg := filepath.Join(b.extractDir(), "boot", "grub", "efi.img")
	if err := os.MkdirAll(filepath.Dir(efiImg), 0755); err != nil {
		log.Errorf("error creating EFI directory: %v", err)
		return false
	}

	cmd := exec.Command("xorisso", "-indev", b.sourceIsoPath(), "-report_el_torito", "plain")
	_, stderr := cmd.CombinedOutput()

	var lba *string
	var blocks *int

	for _, line := range strings.Split(stderr.Error(), "\n") {
		if strings.Contains(line, "EFI image start and size:") {
			match := lbaRegex.FindStringSubmatch(line)
			match1 := match[1]
			lba = &match1                          // Assuming "lba" corresponds to the first captured group
			blocks512, _ := strconv.Atoi(match[2]) // Assuming "block_count" corresponds to the second captured group
			blocks128 := blocks512 / 4
			blocks = &blocks128
		}
	}

	if lba != nil && blocks != nil {
		// Extract the EFI boot image using dd
		cmd := exec.Command(
			"dd",
			fmt.Sprintf("if=%s", b.sourceIsoPath()),
			"bs=2048",
			fmt.Sprintf("skip=%d", lba),
			fmt.Sprintf("count=%d", blocks),
			fmt.Sprintf("of=%s", efiImg),
		)
		err := cmd.Run()
		if err != nil {
			log.Errorf("❌ Failed to extract EFI boot image: %v", err)
			return false
		}
		blockSize := *blocks * 2048
		log.Infof("✅ EFI boot image extracted (LBA: %d, size: %dKB)\n", lba, blockSize/1024)
	} else {
		log.Errorf("⚠️  Warning: Could not detect EFI boot image location")
		return false
	}

	mkisofsCmdArgs := []string{
		"xorriso", "-as", "mkisofs",
		"-r", "-V", "Ubuntu Autoinstall",
		"-J", "-joliet-long",
		"-o", b.destIsoPath(),
		"-b", "boot/grub/i386-pc/eltorito.img",
		"-c", "boot.catalog",
		"-no-emul-boot",
		"-boot-load-size", "4",
		"-boot-info-table",
		"-eltorito-alt-boot",
		"-e", "boot/grub/efi.img",
		"-no-emul-boot",
		"-isohybrid-gpt-basdat",
	}

	if _, err := os.Stat(mbrFile); err == nil {
		mkisofsCmdArgs = append(mkisofsCmdArgs, "-isohybrid-mbr", mbrFile)
	}

	mkisofsCmdArgs = append(mkisofsCmdArgs, b.extractDir())

	cmd = exec.Command("xorisso", mkisofsCmdArgs...)
	if _, err := cmd.CombinedOutput(); err != nil {
		log.Errorf("error building ISO: %v\n", err)
		return false
	} else {
		log.Infof("✅ ISO created: %s", b.destIsoPath())
	}

	return true
}

func (b *ISOBuilder) Build() bool {
	log.Infof(strings.Repeat("=", 60))
	log.Infoln("Ubuntu Cloud-Init Autoinstall ISO Builder")
	log.Infof(strings.Repeat("=", 60))
	log.Infof("\n")

	steps := []struct {
		name string
		fn   func() bool
	}{
		{
			name: "Checking dependencies",
			fn:   b.checkDependencies,
		},
		{
			name: "Downloading ISO",
			fn:   b.downloadIso,
		},
		{
			name: "Extracting ISO",
			fn:   b.extractIso,
		},
		{
			name: "Creating autoinstall config",
			fn:   b.createAutoinstallConfigs,
		},
		{
			name: "Modifying boot config",
			fn:   b.modifyGrubConfig,
		},
		{
			name: "Building final ISO",
			fn:   b.buildIso,
		},
	}

	for _, step := range steps {
		log.Infof("\n📍 Step: %s", step.name)
		if !step.fn() {
			log.Errorf("\\n❌ Build failed at: %s", step.name)
			return false
		}
	}

	log.Infof(strings.Repeat("=", 60))
	log.Infoln("✅ Build complete!")
	log.Infof(strings.Repeat("=", 60))
	log.Infof("\n📀 Output ISO: %s", b.destIsoPath())
	log.Infof("\n💾 Write to USB with:")
	log.Infof("   sudo dd if={self.output_iso} of=/dev/sdX bs=4M status=progress && sync")
	log.Infof("\n")

	return true
}

func NewISOBuilder(cloudConfig, osType, version, outputPath string) *ISOBuilder {
	return &ISOBuilder{
		cloudConfig: cloudConfig,
		osType:      osType,
		version:     version,
		outputPath:  outputPath,
	}
}
