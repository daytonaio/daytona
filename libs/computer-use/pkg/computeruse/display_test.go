// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"bytes"
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScaleImageKeepsTinyPositiveImagesNonZero(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))

	scaled := scaleImage(img, 0.1)

	assert.Equal(t, 1, scaled.Bounds().Dx())
	assert.Equal(t, 1, scaled.Bounds().Dy())
}

func TestEncodeImageWithCompressionSupportsPngAndJpeg(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))

	defaultData, err := encodeImageWithCompression(img, ImageCompressionParams{})
	assert.NoError(t, err)
	assert.True(t, bytes.HasPrefix(defaultData, []byte{0x89, 'P', 'N', 'G'}))

	pngData, err := encodeImageWithCompression(img, ImageCompressionParams{Format: "png"})
	assert.NoError(t, err)
	assert.True(t, bytes.HasPrefix(pngData, []byte{0x89, 'P', 'N', 'G'}))

	jpegData, err := encodeImageWithCompression(img, ImageCompressionParams{Format: "jpeg", Quality: 90})
	assert.NoError(t, err)
	assert.True(t, bytes.HasPrefix(jpegData, []byte{0xff, 0xd8}))
}

func TestEncodeImageWithCompressionRejectsUnsupportedFormat(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))

	unsupportedFormats := []string{"webp", "gif"}
	for _, format := range unsupportedFormats {
		_, err := encodeImageWithCompression(img, ImageCompressionParams{Format: format})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "unsupported image format")
	}
}
