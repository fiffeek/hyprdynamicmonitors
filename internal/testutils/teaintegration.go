package testutils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/charmbracelet/x/exp/golden"
	"github.com/charmbracelet/x/vt"
	"github.com/charmbracelet/x/xpty"
	"golang.org/x/sys/unix"
)

var errReadTimeout = errors.New("timeout while reading")

const (
	maxInt32 = 2147483647
	minInt32 = -2147483648
)

// safeInt32 safely converts an int to int32, returning an error if out of range
func safeInt32(val int, name string) (int32, error) {
	if val < minInt32 || val > maxInt32 {
		return 0, fmt.Errorf("%s %d out of valid range (%d-%d)", name, val, minInt32, maxInt32)
	}
	return int32(val), nil // #nosec G115 -- validated above
}

func findExitSequence(data []byte, exitSequences []string) int {
	minPos := -1

	for _, seq := range exitSequences {
		if pos := bytes.Index(data, []byte(seq)); pos != -1 {
			if minPos == -1 || pos < minPos {
				minPos = pos
			}
		}
	}

	return minPos
}

type WaitingForContext struct {
	Duration      time.Duration
	CheckInterval time.Duration
}

type WaitForOption func(*WaitingForContext)

func WithCheckInterval(d time.Duration) WaitForOption {
	return func(wf *WaitingForContext) {
		wf.CheckInterval = d
	}
}

func WithDuration(d time.Duration) WaitForOption {
	return func(wf *WaitingForContext) {
		wf.Duration = d
	}
}

type TestModelOptions struct {
	width         int
	height        int
	cmd           *exec.Cmd
	exitSequences []string
}

type TestOption func(opts *TestModelOptions)

func WithExitSequences(seq []string) TestOption {
	return func(opts *TestModelOptions) {
		opts.exitSequences = seq
	}
}

func WithInitialTermSize(width, height int) TestOption {
	return func(opts *TestModelOptions) {
		opts.width = width
		opts.height = height
	}
}

func WithCommand(cmd *exec.Cmd) TestOption {
	return func(opts *TestModelOptions) {
		opts.cmd = cmd
	}
}

type TestModel struct {
	ptmx          xpty.Pty
	vt            *vt.Emulator
	done          sync.Once
	exitSequences []string
	doneCh        chan error
	cmd           *exec.Cmd
	readTimeout   time.Duration
	outputBuffer  bytes.Buffer
	mu            sync.Mutex
}

func NewTestModel(options ...TestOption) (*TestModel, error) {
	opts := TestModelOptions{
		width:  100,
		height: 50,
		exitSequences: []string{
			"\x1b[?1049l", // exit alternate screen
			"\x1b[?25h",   // show cursor
		},
	}
	for _, opt := range options {
		opt(&opts)
	}

	if opts.cmd == nil {
		return nil, errors.New("cmd is required")
	}

	ptmx, err := xpty.NewPty(opts.width, opts.height)
	if err != nil {
		return nil, fmt.Errorf("cant get pty: %w", err)
	}

	if err := ptmx.Start(opts.cmd); err != nil {
		return nil, fmt.Errorf("cant start pty: %w", err)
	}

	vt := vt.NewEmulator(opts.width, opts.height)

	t := &TestModel{
		ptmx:          ptmx,
		vt:            vt,
		done:          sync.Once{},
		doneCh:        make(chan error),
		cmd:           opts.cmd,
		readTimeout:   200 * time.Millisecond,
		outputBuffer:  bytes.Buffer{},
		mu:            sync.Mutex{},
		exitSequences: opts.exitSequences,
	}

	fd := int(t.ptmx.Fd())
	err = syscall.SetNonblock(fd, true)
	if err != nil {
		return nil, fmt.Errorf("cant set nonblocking on pty fd: %w", err)
	}

	go func() {
		defer func() {
			_ = t.ptmx.Close()
		}()
		t.doneCh <- t.cmd.Wait()
	}()

	return t, nil
}

func (t *TestModel) Send(sequence []byte) error {
	err := writeAll(t.ptmx, sequence)
	if err != nil {
		return fmt.Errorf("cant write the input sequence: %w", err)
	}

	return nil
}

func (t *TestModel) Type(s string) error {
	for _, c := range s {
		if err := t.Send([]byte(string(c))); err != nil {
			return fmt.Errorf("cant write %s: %w", string(c), err)
		}
	}
	return nil
}

func (t *TestModel) readWithInterrupt(buf []byte, timeout time.Duration) (int, error) {
	deadline := time.Now().Add(timeout)
	fd := int(t.ptmx.Fd())

	fd32, err := safeInt32(fd, "file descriptor")
	if err != nil {
		return 0, err
	}

	for {
		if time.Now().After(deadline) {
			return 0, errReadTimeout
		}

		pollFds := []unix.PollFd{
			{
				Fd:     fd32,
				Events: unix.POLLIN,
			},
		}

		remainingTimeout := time.Until(deadline)
		if remainingTimeout < 0 {
			return 0, errReadTimeout
		}

		n, err := unix.Poll(pollFds, int(remainingTimeout.Milliseconds()))
		if err != nil {
			return 0, fmt.Errorf("cant poll: %w", err)
		}

		if n == 0 {
			return 0, errReadTimeout
		}

		if pollFds[0].Revents&(unix.POLLERR|unix.POLLHUP|unix.POLLNVAL) != 0 {
			return 0, fmt.Errorf("poll error: revents=%d", pollFds[0].Revents)
		}

		if pollFds[0].Revents&unix.POLLIN != 0 {
			bytesRead, readErr := syscall.Read(fd, buf)

			if readErr == syscall.EAGAIN || readErr == syscall.EWOULDBLOCK {
				// Spurious wakeup, retry
				continue
			}

			if readErr != nil {
				return bytesRead, fmt.Errorf("cant read data: %w", readErr)
			}

			return bytesRead, nil
		}
	}
}

func (t *TestModel) setReadTimeout(timeout time.Duration) {
	t.readTimeout = timeout
}

func (t *TestModel) Read(p []byte) (n int, err error) {
	read, err := t.readWithInterrupt(p, t.readTimeout)
	if err != nil {
		return read, fmt.Errorf("cant read from ptmx: %w", err)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	writeErr := writeAll(&t.outputBuffer, p[:read])
	if writeErr != nil {
		return read, fmt.Errorf("cant write to internal buffer: %w", writeErr)
	}

	// Check if we've seen exit sequences - if so, stop reading
	if findExitSequence(t.outputBuffer.Bytes(), t.exitSequences) != -1 {
		return read, io.EOF
	}

	return read, nil
}

func (t *TestModel) WaitFor(condition func(bts []byte) bool, options ...WaitForOption) error {
	wf := WaitingForContext{
		Duration:      time.Second,
		CheckInterval: 50 * time.Millisecond,
	}

	for _, opt := range options {
		opt(&wf)
	}

	var b bytes.Buffer
	start := time.Now()
	t.setReadTimeout(wf.CheckInterval)

	for time.Since(start) <= wf.Duration {
		buf := make([]byte, 4096)
		n, err := t.Read(buf)

		if err != nil && !errors.Is(err, errReadTimeout) {
			return fmt.Errorf("WaitFor: %w", err)
		}

		if n > 0 {
			b.Write(buf[:n])
		}

		if condition(b.Bytes()) {
			return nil
		}

		time.Sleep(wf.CheckInterval)
	}
	return fmt.Errorf("WaitFor: condition not met after %s. Last frames:\n %q", wf.Duration, b.String())
}

type FinalOpts struct {
	timeout time.Duration
}

type FinalOpt func(opts *FinalOpts)

func WithFinalTimeout(d time.Duration) FinalOpt {
	return func(opts *FinalOpts) {
		opts.timeout = d
	}
}

func (t *TestModel) waitDone(opts []FinalOpt) error {
	var err error
	t.done.Do(func() {
		fopts := FinalOpts{}
		for _, opt := range opts {
			opt(&fopts)
		}
		if fopts.timeout > 0 {
			select {
			case <-time.After(fopts.timeout):
				err = errors.New("timeout while waiting for done")
				return
			case doneErr, ok := <-t.doneCh:
				if !ok {
					err = errors.New("done channel closed")
					return
				}
				err = doneErr
				return
			}
		} else {
			doneErr, ok := <-t.doneCh
			if !ok {
				err = errors.New("done channel closed")
				return
			}
			err = doneErr
		}
	})
	return err
}

func (t *TestModel) WaitFinished(opts ...FinalOpt) error {
	readDone := make(chan error)
	defer close(readDone)

	go func() {
		_, err := io.ReadAll(t)
		readDone <- err
	}()

	doneErr := t.waitDone(opts)
	readErr, ok := <-readDone
	if !ok {
		return errors.New("read channel closed")
	}
	if readErr != nil && !errors.Is(readErr, syscall.EIO) && !errors.Is(readErr, syscall.EINTR) {
		return fmt.Errorf("cant read full output: %w", readErr)
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	bufferedData := t.outputBuffer.Bytes()

	exitPos := findExitSequence(bufferedData, t.exitSequences)
	var dataToWrite []byte
	if exitPos != -1 {
		dataToWrite = bufferedData[:exitPos]
	} else {
		dataToWrite = bufferedData
	}

	if err := writeAll(t.vt, dataToWrite); err != nil {
		return fmt.Errorf("cant hydrate virtual terminal: %w", err)
	}

	if doneErr != nil {
		return fmt.Errorf("program exited: %w", doneErr)
	}

	return nil
}

func (t *TestModel) FinalScreen(opts ...FinalOpt) (string, error) {
	err := t.WaitFinished(opts...)
	return t.vt.String(), err
}

func (t *TestModel) FinalOutput(opts ...FinalOpt) (string, error) {
	err := t.WaitFinished(opts...)
	return t.outputBuffer.String(), err
}

func writeAll(w io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := w.Write(data)
		if err != nil {
			return fmt.Errorf("cant write data chunk: %w", err)
		}
		data = data[n:]
	}
	return nil
}

func RequireEqualOutput(tb testing.TB, out []byte) {
	tb.Helper()
	golden.RequireEqualEscape(tb, out, true)
}
