// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"net/http"
	"strings"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/gin-gonic/gin"
	"github.com/go-vgo/robotgo"
)

func (u *ComputerUse) TypeText(req *computeruse.ComputerUseRequest) (*computeruse.Empty, error) {
	var input struct {
		Text  string `json:"text"`
		Delay int    `json:"delay"` // milliseconds between keystrokes
	}

	if err := req.RequestContext.ShouldBindJSON(&input); err != nil {
		req.RequestContext.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input",
		})
		return new(computeruse.Empty), nil
	}

	if input.Delay > 0 {
		robotgo.TypeStr(input.Text, input.Delay)
	} else {
		robotgo.TypeStr(input.Text)
	}

	req.RequestContext.JSON(http.StatusOK, gin.H{
		"success": true,
		"typed":   input.Text,
	})
	return new(computeruse.Empty), nil
}

func (u *ComputerUse) PressKey(req *computeruse.ComputerUseRequest) (*computeruse.Empty, error) {
	var key struct {
		Key       string   `json:"key"`
		Modifiers []string `json:"modifiers"` // ctrl, alt, shift, cmd
	}

	if err := req.RequestContext.ShouldBindJSON(&key); err != nil {
		req.RequestContext.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid key",
		})
		return new(computeruse.Empty), nil
	}

	if len(key.Modifiers) > 0 {
		robotgo.KeyTap(key.Key, key.Modifiers)
	} else {
		robotgo.KeyTap(key.Key)
	}

	req.RequestContext.JSON(http.StatusOK, gin.H{
		"success":   true,
		"key":       key.Key,
		"modifiers": key.Modifiers,
	})
	return new(computeruse.Empty), nil
}

func (u *ComputerUse) PressHotkey(req *computeruse.ComputerUseRequest) (*computeruse.Empty, error) {
	var hotkey struct {
		Keys string `json:"keys"` // e.g., "ctrl+c", "cmd+v"
	}

	if err := req.RequestContext.ShouldBindJSON(&hotkey); err != nil {
		req.RequestContext.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid hotkey",
		})
		return new(computeruse.Empty), nil
	}

	keys := strings.Split(hotkey.Keys, "+")
	if len(keys) < 2 {
		req.RequestContext.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid hotkey format",
		})
		return new(computeruse.Empty), nil
	}

	mainKey := keys[len(keys)-1]
	modifiers := keys[:len(keys)-1]

	robotgo.KeyTap(mainKey, modifiers)

	req.RequestContext.JSON(http.StatusOK, gin.H{
		"success": true,
		"hotkey":  hotkey.Keys,
	})
	return new(computeruse.Empty), nil
}
