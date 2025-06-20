// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-vgo/robotgo"
)

func (u *ComputerUse) TypeText(c *gin.Context) {
	var input struct {
		Text  string `json:"text"`
		Delay int    `json:"delay"` // milliseconds between keystrokes
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input",
		})
		return
	}

	if input.Delay > 0 {
		robotgo.TypeStr(input.Text, input.Delay)
	} else {
		robotgo.TypeStr(input.Text)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"typed":   input.Text,
	})
}

func (u *ComputerUse) PressKey(c *gin.Context) {
	var key struct {
		Key       string   `json:"key"`
		Modifiers []string `json:"modifiers"` // ctrl, alt, shift, cmd
	}

	if err := c.ShouldBindJSON(&key); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid key",
		})
		return
	}

	if len(key.Modifiers) > 0 {
		robotgo.KeyTap(key.Key, key.Modifiers)
	} else {
		robotgo.KeyTap(key.Key)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"key":       key.Key,
		"modifiers": key.Modifiers,
	})
}

func (u *ComputerUse) PressHotkey(c *gin.Context) {
	var hotkey struct {
		Keys string `json:"keys"` // e.g., "ctrl+c", "cmd+v"
	}

	if err := c.ShouldBindJSON(&hotkey); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid hotkey",
		})
		return
	}

	keys := strings.Split(hotkey.Keys, "+")
	if len(keys) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid hotkey format",
		})
		return
	}

	mainKey := keys[len(keys)-1]
	modifiers := keys[:len(keys)-1]

	robotgo.KeyTap(mainKey, modifiers)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"hotkey":  hotkey.Keys,
	})
}
