package testutils

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/stretchr/testify/assert"
)

func AssertFileExists(t *testing.T, path string) {
	_, err := os.Stat(path)
	assert.NoError(t, err, "file should exist")
}

func AssertFileDoesNotExist(t *testing.T, path string) {
	stat, err := os.Stat(path)
	assert.Error(t, err, "file should not exist")
	assert.Nil(t, stat, "file should not exist")
}

func AssertIsSymlink(t *testing.T, path string) {
	fi, err := os.Lstat(path)
	assert.NoError(t, err, "file should exist")
	isSymlink := fi.Mode()&fs.ModeSymlink != 0
	assert.True(t, isSymlink, "file should be a symlink")
}

func ContentSameAsFixture(t *testing.T, targetFile, fixtureFile string) error {
	// nolint:gosec
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		return fmt.Errorf("target content cant be read: %w", err)
	}

	// nolint:gosec
	fixtureContent, err := os.ReadFile(fixtureFile)
	if err != nil {
		return fmt.Errorf("fixture content cant be read: %w", err)
	}

	if !reflect.DeepEqual(string(targetContent), string(fixtureContent)) {
		return errors.New("contents differ")
	}

	return nil
}

func AssertFixture(t *testing.T, target, fixture string, regenerate bool) {
	if regenerate {
		UpdateFixture(t, target, fixture)
		return
	}
	AssertContentsSameAsFixture(t, target, fixture)
}

func AssertContentsSameAsFixture(t *testing.T, targetFile, fixtureFile string) {
	// nolint:gosec
	targetContent, err := os.ReadFile(targetFile)
	assert.NoError(t, err, "should be able to read the target file")
	// nolint:gosec
	fixtureContent, err := os.ReadFile(fixtureFile)
	assert.NoError(t, err, "should be able to read the fixture file")
	assert.Equal(t, string(fixtureContent), string(targetContent),
		"target content should be the same as in the figture %s", fixtureContent)
}

func UpdateFixture(t *testing.T, targetFile, fixtureFile string) {
	// nolint:gosec
	targetContent, err := os.ReadFile(targetFile)
	assert.NoError(t, err, "should be able to read the target file")
	// nolint:gosec
	_, err = os.ReadFile(fixtureFile)
	assert.NoError(t, err, "should be able to read the fixture file")
	assert.NoError(t, utils.WriteAtomic(fixtureFile, targetContent), "cant write to file")
}

func SetupFakeConfigUpdater(t *testing.T, updates []*TestConfig,
	initialSleep, sleepBetweenEvents time.Duration, binaryStarting <-chan struct{}, configPath string,
) chan struct{} {
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		t.Log("Waiting for the server to start")
		<-binaryStarting
		t.Log("Starting fake config writer")
		time.Sleep(initialSleep)

		for _, update := range updates {
			// ensure it materializes in the same config
			_ = update.WithConfigPath(configPath).Get()
			t.Log("Wrote configuration update")
			time.Sleep(sleepBetweenEvents)
		}
	}()
	return serverDone
}
