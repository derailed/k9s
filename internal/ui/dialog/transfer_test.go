// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dialog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testPodPath   = "default/pod:"
	testLocalPath = "/tmp/downloads"
)

func TestTransferArgsKeepConfiguredPathOnLocalSide(t *testing.T) {
	opts := TransferDialogOpts{
		Pod:       testPodPath,
		LocalPath: testLocalPath,
		Retries:   2,
	}

	args := newTransferArgs(&opts)
	assert.Equal(t, TransferArgs{
		From:     testPodPath,
		To:       testLocalPath,
		Download: true,
		Retries:  2,
	}, args)

	args.setDownload(false)
	assert.Equal(t, TransferArgs{
		From:    testLocalPath,
		To:      testPodPath,
		Retries: 2,
	}, args)

	args.setDownload(true)
	assert.Equal(t, TransferArgs{
		From:     testPodPath,
		To:       testLocalPath,
		Download: true,
		Retries:  2,
	}, args)
}

func TestTransferArgsPreserveEmptyLocalPath(t *testing.T) {
	args := newTransferArgs(&TransferDialogOpts{Pod: testPodPath})

	assert.Empty(t, args.To)
}
