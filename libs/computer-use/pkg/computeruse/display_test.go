// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
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
