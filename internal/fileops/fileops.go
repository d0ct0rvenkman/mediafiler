package fileops

import (
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
)

/*
	os.Rename() doesn't work well when renaming files across block devices. This takes a
	two-tiered approach.

	Thanks!
	- https://gist.github.com/var23rav/23ae5d0d4d830aff886c3c970b8f6c6b?permalink_comment_id=4431960#gistcomment-4431960
*/

func Move(source, destination string) error {
	err := os.Rename(source, destination)
	if err != nil && strings.Contains(err.Error(), "invalid cross-device link") {
		return moveCrossDevice(source, destination)
	}
	return err
}

func moveCrossDevice(source, destination string) error {
	src, err := os.Open(source)
	if err != nil {
		return errors.Wrap(err, "Open(source)")
	}
	dst, err := os.Create(destination)
	if err != nil {
		src.Close()
		return errors.Wrap(err, "Create(destination)")
	}
	_, err = io.Copy(dst, src)
	src.Close()
	dst.Close()
	if err != nil {
		return errors.Wrap(err, "Copy")
	}
	fi, err := os.Stat(source)
	if err != nil {
		os.Remove(destination)
		return errors.Wrap(err, "Stat")
	}
	err = os.Chmod(destination, fi.Mode())
	if err != nil {
		os.Remove(destination)
		return errors.Wrap(err, "Stat")
	}
	os.Remove(source)
	return nil
}
