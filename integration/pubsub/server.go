/*
Copyright: 2025, Deep Codify Limited

intercart
*/

package pubsub

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/rs/zerolog/log"
)

var _pubSubTestServer pubSubTestServer

type pubSubTestServer struct {
	cmd *exec.Cmd
}

var (
	_isTestPubSubEmulatorReady bool = true

	_readyLock sync.Mutex
)

func StartTestServer(port int) error {
	if _pubSubTestServer.cmd != nil {
		log.Error().Msg("pubsub-emulator already started")
		return errors.New("pubsub-emulator already started")
	}

	setReady(false)

	downloadPath := "./dist/pubsub-emulator"
	binPath := filepath.Join(downloadPath, "bin", "cloud-pubsub-emulator")
	libPath := filepath.Join(downloadPath, "lib")

	needDownload := false
	_, errBinPath := os.Stat(binPath)
	_, errLibPath := os.Stat(libPath)
	if errBinPath != nil || errLibPath != nil {
		log.Warn().Msg("emulator binary not found, downloading...")
		needDownload = true
	}

	var err error
	if needDownload {
		binPath, err = DownloadEmulator(downloadPath)
		if err != nil {
			log.Err(err).Msg("failed to download emulator")
			return err
		}
	}

	log.Info().Msg("starting pub/sub emulator")
	var args []string
	args = append(args, "--host=0.0.0.0")
	args = append(args, fmt.Sprintf("--port=%d", port))
	cmd := exec.Command(binPath, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		log.Err(err).Msg("failed to start emulator")
		return err
	}

	setReady(true)
	_pubSubTestServer.cmd = cmd

	return nil
}

func CloseTestServer() error {
	if _pubSubTestServer.cmd == nil {
		return nil
	}

	log.Info().Msg("stopping emulator")
	err := _pubSubTestServer.cmd.Process.Kill()
	if err != nil {
		log.Err(err).Msg("failed to stop emulator")
		return err
	}

	_pubSubTestServer.cmd = nil

	return nil
}

func IsReady() bool {
	_readyLock.Lock()
	defer _readyLock.Unlock()

	return _isTestPubSubEmulatorReady
}

func setReady(ready bool) {
	_readyLock.Lock()
	defer _readyLock.Unlock()

	_isTestPubSubEmulatorReady = ready
}
