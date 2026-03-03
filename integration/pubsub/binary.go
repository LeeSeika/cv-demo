package pubsub

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

//go:embed script/install_google_cloud_sdk.sh
var script []byte

func DownloadEmulator(downloadPath string) (string, error) {
	baseDir := filepath.Join(os.TempDir(), "google")
	err := os.RemoveAll(baseDir)
	if err != nil {
		log.Err(err).Msgf("failed to clean directory %s", baseDir)
		return "", err
	}

	log.Info().Msg("downloading gcloud...")
	binPath, err := downloadGCloud(baseDir)
	if err != nil {
		return "", fmt.Errorf("failed to install gcloud: %w", err)
	}
	log.Info().Msg("gcloud is successfully downloaded")

	log.Info().Msg("downloading pub/sub emulator...")
	emulatorPath, err := downloadEmulator(binPath, baseDir)
	if err != nil {
		return "", fmt.Errorf("failed to install pub/sub emulator: %w", err)
	}
	log.Info().Msg("pub/sub emulator is successfully downloaded")

	log.Info().Msg("copying pub/sub emulator binary...")
	err = copyEmulatorBinary(emulatorPath, downloadPath)
	if err != nil {
		return "", fmt.Errorf("failed to copy pub/sub emulator binary: %w", err)
	}
	log.Info().Msg("pub/sub emulator binary is successfully copied")

	cloudPubSubEmulatorBinPath := filepath.Join(downloadPath, "bin", "cloud-pubsub-emulator")
	log.Info().Msgf("pub/sub emulator binary is at: %s", cloudPubSubEmulatorBinPath)

	return cloudPubSubEmulatorBinPath, nil
}

func downloadGCloud(baseDir string) (string, error) {
	cmd := exec.Command("bash", "-s", "--",
		"--disable-prompts",
		fmt.Sprintf("--install-dir=%s", baseDir))

	cmd.Stdin = bytes.NewReader(script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Err(err).Msg("failed to run install_google_cloud_sdk.sh")
		return "", err
	}

	binPath := filepath.Join(baseDir, "google-cloud-sdk", "bin", "gcloud")
	return binPath, nil
}

func downloadEmulator(binPath, baseDir string) (string, error) {
	cmd := exec.Command(binPath, "components", "install", "pubsub-emulator", "--quiet")

	cmd.Env = append(os.Environ(),
		"CLOUDSDK_CORE_DISABLE_PROMPTS=1",
		"CLOUDSDK_COMPONENT_MANAGER_DISABLE_UPDATE_CHECK=1",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Err(err).Msg("failed to run pub/sub emulator installation")
		return "", err
	}

	pubsubEmulatorDir := filepath.Join(baseDir, "google-cloud-sdk", "platform", "pubsub-emulator")
	return pubsubEmulatorDir, nil
}

func copyEmulatorBinary(source, dest string) error {
	err := os.MkdirAll(dest, 0o755)
	if err != nil {
		log.Err(err).Msg("failed to create dir")
		return err
	}

	err = os.RemoveAll(dest)
	if err != nil {
		log.Err(err).Msgf("failed to remove old files in %s", dest)
		return err
	}

	err = os.Rename(source, dest)
	if err != nil {
		log.Err(err).Msg("failed to move cloud-pubsub-emulator binary")
		return err
	}

	return nil
}
