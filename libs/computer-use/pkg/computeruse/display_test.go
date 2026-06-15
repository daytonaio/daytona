// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"bytes"
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScaleImageKeepsTinyPositiveImagesNonZero(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))

	scaled := scaleImage(img, 0.1)

	assert.Equal(t, 1, scaled.Bounds().Dx())
	assert.Equal(t, 1, scaled.Bounds().Dy())
}

func TestScaleImageReturnsDegenerateImagesUnchanged(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 0, 0))

	scaled := scaleImage(img, 0.5)

	assert.Same(t, image.Image(img), scaled)
}

func TestValidateScreenshotRegion(t *testing.T) {
	assert.NoError(t, validateScreenshotRegion(1, 1))
	assert.Error(t, validateScreenshotRegion(0, 1))
	assert.Error(t, validateScreenshotRegion(1, 0))
	assert.Error(t, validateScreenshotRegion(-1, -1))
}

// Scaling must apply for every format value — including unrecognized ones
// that fall back to PNG — because callers report cursor coordinates
// multiplied by Scale and would desync from an unscaled image.
func TestEncodeImageWithCompressionScalesAllFormats(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	for _, format := range []string{"png", "jpeg", "webp", ""} {
		params := ImageCompressionParams{Format: format, Quality: 80, Scale: 0.5}

		data, err := encodeImageWithCompression(img, params)
		require.NoError(t, err, "format %q", format)

		decoded, _, err := image.Decode(bytes.NewReader(data))
		require.NoError(t, err, "format %q", format)
		assert.Equal(t, 5, decoded.Bounds().Dx(), "format %q", format)
		assert.Equal(t, 5, decoded.Bounds().Dy(), "format %q", format)
	}
}
