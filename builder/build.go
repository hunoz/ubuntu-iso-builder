package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	generate_cloud_config "github.com/hunoz/ubuntu-iso-builder/generate-cloud-config"
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
	log.Infof("⚙️ Checking dependencies")
	commands := []Dependency{
		{
			Name:         "xorriso",
			CheckCommand: func() bool { return commandIsInPath("xorriso") },
			FixMessage:   "xorriso needs to be installed",
		},
		{
			Name:         "7z",
			CheckCommand: func() bool { return commandIsInPath("7z") },
			FixMessage:   "7z needs to be installed",
		},
		{
			Name:         "dd",
			CheckCommand: func() bool { return commandIsInPath("dd") },
			FixMessage:   "dd needs to be installed",
		},
	}

	for _, command := range commands {
		if command.CheckCommand() {
			continue
		}

		log.Errorf("%s missing: %s", command.Name, command.FixMessage)

		return false
	}
	log.Infoln("✅ All dependencies are installed")
	return true
}

func (b *ISOBuilder) modifyGrubForAutoinstall(content string) string {
	modified := strings.Replace(content, "---", "autoinstall ds=nocloud;s=/cdrom/nocloud/ ---", -1)

	return strings.Replace(modified, "set timeout=30", "set timeout=5", -1)
}

func (b *ISOBuilder) downloadIso() bool {
	log.Infof("📀 Checking if ISO is already present")
	stat, err := os.Stat(b.sourceIsoPath())
	if err == nil && stat.Size() > 0 {
		log.Infof("📀 ISO already present")
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
	log.Infoln("📦 Extracting ISO contents...")
	extractDir := b.extractDir()
	// Remove existing directory if it exists
	if _, err := os.Stat(extractDir); err == nil {
		if err := os.RemoveAll(extractDir); err != nil {
			log.Infoln("❌ Failed to remove existing directory: %v\n", err)
			return false
		}
	}

	// Create the extraction directory
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		log.Infoln("❌ Failed to create directory: %v\n", err)
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

	log.Infoln("✅ Extraction complete")
	return true
}
func (b *ISOBuilder) createAutoinstallConfigs() bool {
	var config generate_cloud_config.CloudConfig
	cfg := b.cloudConfig
	if strings.HasPrefix("#cloud-config", cfg) {
		cfg = strings.TrimPrefix("#cloud-config\n", cfg)
	}
	if err := yaml.Unmarshal([]byte(b.cloudConfig), &config); err != nil {
		log.Errorf("error parsing cloud-config: %v", err)
		return false
	}
	dump := strings.Join([]string{"#cloud-config", b.cloudConfig}, "\n")
	hostname := config.AutoInstall.UserData.Hostname

	autoinstallFile := filepath.Join(b.extractDir(), "autoinstall.yaml")
	if err := os.WriteFile(autoinstallFile, []byte(dump), 0644); err != nil {
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

	// Extract MBR template from the system area of the source ISO
	mbrFile := filepath.Join(b.extractDir(), "isohdpfx.bin")
	if err := b.extractMBRTemplate(mbrFile); err != nil {
		log.Warnf("⚠️  Could not extract MBR template: %v. ISO may not boot on legacy BIOS", err)
		mbrFile = "" // Clear it so we don't try to use it
	} else {
		log.Infof("✅ MBR template extracted (%d bytes)", 432)
	}

	efiImg := filepath.Join(b.extractDir(), "boot", "grub", "efi.img")

	// Check if efi.img already exists in extracted files
	if stat, err := os.Stat(efiImg); err != nil || stat.Size() == 0 {
		log.Infof("📀 Extracting EFI boot image from source ISO...")

		if !b.extractEfiImage(efiImg) {
			return false
		}
	} else {
		log.Infof("✅ EFI image found in extracted files (%d KB)", stat.Size()/1024)
	}

	mkisofsCmdArgs := []string{
		"-as", "mkisofs",
		"-r", "-V", "Ubuntu-Autoinstall",
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

	if mbrFile != "" {
		mkisofsCmdArgs = append(mkisofsCmdArgs, "-isohybrid-mbr", mbrFile)
	}

	mkisofsCmdArgs = append(mkisofsCmdArgs, b.extractDir())

	cmd := exec.Command("xorriso", mkisofsCmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("xorriso output:\n%s", string(out))
		log.Errorf("error building ISO: %v", err)
		return false
	}

	log.Infof("✅ ISO created: %s", b.destIsoPath())
	return true
}

// extractMBRTemplate extracts the MBR template from the system area of the ISO
func (b *ISOBuilder) extractMBRTemplate(outputPath string) error {
	// The MBR template is in the first 432 bytes of the ISO
	// (some sources say 446 bytes, but xorriso typically uses 432)
	cmd := exec.Command(
		"dd",
		fmt.Sprintf("if=%s", b.sourceIsoPath()),
		fmt.Sprintf("of=%s", outputPath),
		"bs=1",
		"count=432",
		"skip=0",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("dd failed: %v, output: %s", err, string(output))
	}

	// Verify the file was created and has the right size
	stat, err := os.Stat(outputPath)
	if err != nil {
		return fmt.Errorf("could not verify MBR file: %v", err)
	}

	if stat.Size() != 432 {
		return fmt.Errorf("MBR file has unexpected size: %d bytes (expected 432)", stat.Size())
	}

	return nil
}

func (b *ISOBuilder) extractEfiImage(efiImg string) bool {
	cmd := exec.Command("xorriso", "-indev", b.sourceIsoPath(), "-report_el_torito", "plain")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("error running xorriso: %v", err)
		return false
	}

	var lba string
	var blocks int

	for _, line := range strings.Split(string(stdout), "\n") {
		if strings.Contains(line, "EFI image start and size:") {
			match := lbaRegex.FindStringSubmatch(line)
			if len(match) >= 4 {
				lba = match[1]
				blockSize, _ := strconv.Atoi(match[2])
				blockCount, _ := strconv.Atoi(match[3])
				blocks = (blockSize * blockCount) / 2048
			}
		}
	}

	if lba == "" || blocks == 0 {
		log.Errorf("❌ Could not detect EFI boot image location")
		return false
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(efiImg), 0755); err != nil {
		log.Errorf("error creating EFI directory: %v", err)
		return false
	}

	// Extract the EFI boot image using dd
	cmd = exec.Command(
		"dd",
		fmt.Sprintf("if=%s", b.sourceIsoPath()),
		"bs=2048",
		fmt.Sprintf("skip=%s", lba),
		fmt.Sprintf("count=%d", blocks),
		fmt.Sprintf("of=%s", efiImg),
	)
	if err := cmd.Run(); err != nil {
		log.Errorf("❌ Failed to extract EFI boot image: %v", err)
		return false
	}

	log.Infof("✅ EFI boot image extracted (LBA: %s, size: %d KB)", lba, blocks*2)
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
		log.Infof("📍 Step: %s", step.name)
		if !step.fn() {
			log.Errorf("❌ Build failed at: %s", step.name)
			return false
		}
	}

	log.Infof(strings.Repeat("=", 60))
	log.Infoln("✅ Build complete!")
	log.Infof(strings.Repeat("=", 60))
	log.Infof("📀 Output ISO: %s", b.destIsoPath())
	log.Infof("💾 Write to USB with:")
	log.Infof("sudo dd if=%s of=/dev/sdX bs=4M status=progress && sync", b.destIsoPath())
	log.Infof("")

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
