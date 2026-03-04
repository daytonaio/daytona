// Copyright 2019-2025 Daytona Labs Inc.
// SPDX-License-Identifier: Apache-2.0

use crate::error::DaytonaError;
use crate::services::ServiceClient;
use crate::types::{
    DisplayInfo, Recording, ScreenshotOptions, ScreenshotRegion, ScreenshotResponse, WindowInfo,
};
use serde_json::{json, Value};

#[derive(Clone)]
pub struct ComputerUseService {
    client: ServiceClient,
}

impl ComputerUseService {
    pub fn new(client: ServiceClient) -> Self {
        Self { client }
    }

    pub async fn start(&self) -> Result<Value, DaytonaError> {
        self.client
            .post_json_value("/computer-use/start", &json!({}))
            .await
    }

    pub async fn stop(&self) -> Result<Value, DaytonaError> {
        self.client
            .post_json_value("/computer-use/stop", &json!({}))
            .await
    }

    /// Take a screenshot of the current display.
    ///
    /// # Arguments
    /// * `options` - Optional screenshot settings (show_cursor, format, quality, scale)
    ///
    /// # Returns
    /// A `ScreenshotResponse` containing the base64-encoded image and metadata.
    pub async fn screenshot(
        &self,
        options: Option<&ScreenshotOptions>,
    ) -> Result<ScreenshotResponse, DaytonaError> {
        let mut url = "/computer-use/screenshot".to_string();

        if let Some(opts) = options {
            let mut params = Vec::new();
            if let Some(show_cursor) = opts.show_cursor {
                params.push(format!("show_cursor={}", show_cursor));
            }
            if let Some(ref format) = opts.format {
                params.push(format!("format={}", format));
            }
            if let Some(quality) = opts.quality {
                params.push(format!("quality={}", quality));
            }
            if let Some(scale) = opts.scale {
                params.push(format!("scale={}", scale));
            }
            if !params.is_empty() {
                url.push('?');
                url.push_str(&params.join("&"));
            }
        }

        self.client.get_json(&url).await
    }

    /// Take a screenshot of a specific region of the display.
    ///
    /// # Arguments
    /// * `region` - The region to capture (x, y, width, height)
    /// * `options` - Optional screenshot settings
    ///
    /// # Returns
    /// A `ScreenshotResponse` containing the base64-encoded image and metadata.
    pub async fn screenshot_region(
        &self,
        region: &ScreenshotRegion,
        options: Option<&ScreenshotOptions>,
    ) -> Result<ScreenshotResponse, DaytonaError> {
        let mut params = vec![
            format!("x={}", region.x),
            format!("y={}", region.y),
            format!("width={}", region.width),
            format!("height={}", region.height),
        ];

        if let Some(opts) = options {
            if let Some(show_cursor) = opts.show_cursor {
                params.push(format!("show_cursor={}", show_cursor));
            }
            if let Some(ref format) = opts.format {
                params.push(format!("format={}", format));
            }
            if let Some(quality) = opts.quality {
                params.push(format!("quality={}", quality));
            }
            if let Some(scale) = opts.scale {
                params.push(format!("scale={}", scale));
            }
        }

        let url = format!("/computer-use/screenshot?{}", params.join("&"));
        self.client.get_json(&url).await
    }

    pub async fn mouse_click(
        &self,
        x: i32,
        y: i32,
        button: Option<&str>,
        double: Option<bool>,
    ) -> Result<Value, DaytonaError> {
        self.client
            .post_json(
                "/computer-use/mouse/click",
                &json!({ "x": x, "y": y, "button": button.unwrap_or("left"), "double": double.unwrap_or(false) }),
            )
            .await
    }

    pub async fn mouse_move(&self, x: i32, y: i32) -> Result<Value, DaytonaError> {
        self.client
            .post_json("/computer-use/mouse/move", &json!({ "x": x, "y": y }))
            .await
    }

    pub async fn mouse_drag(
        &self,
        start_x: i32,
        start_y: i32,
        end_x: i32,
        end_y: i32,
    ) -> Result<Value, DaytonaError> {
        self.client
            .post_json(
                "/computer-use/mouse/drag",
                &json!({ "startX": start_x, "startY": start_y, "endX": end_x, "endY": end_y }),
            )
            .await
    }

    pub async fn mouse_scroll(
        &self,
        x: i32,
        y: i32,
        direction: &str,
        amount: Option<i32>,
    ) -> Result<Value, DaytonaError> {
        self.client
            .post_json(
                "/computer-use/mouse/scroll",
                &json!({ "x": x, "y": y, "direction": direction, "amount": amount.unwrap_or(1) }),
            )
            .await
    }

    pub async fn keyboard_type(&self, text: &str) -> Result<(), DaytonaError> {
        self.client
            .post_empty("/computer-use/keyboard/type", &json!({ "text": text }))
            .await
    }

    pub async fn keyboard_press(&self, key: &str) -> Result<(), DaytonaError> {
        self.client
            .post_empty("/computer-use/keyboard/press", &json!({ "key": key }))
            .await
    }

    pub async fn keyboard_hotkey(&self, keys: &str) -> Result<(), DaytonaError> {
        self.client
            .post_empty("/computer-use/keyboard/hotkey", &json!({ "keys": keys }))
            .await
    }

    pub async fn display_info(&self) -> Result<DisplayInfo, DaytonaError> {
        self.client.get_json("/computer-use/display/info").await
    }

    pub async fn display_windows(&self) -> Result<Vec<WindowInfo>, DaytonaError> {
        let value: Value = self
            .client
            .get_json("/computer-use/display/windows")
            .await?;
        Ok(value
            .get("windows")
            .and_then(Value::as_array)
            .cloned()
            .unwrap_or_default()
            .into_iter()
            .filter_map(|v| serde_json::from_value(v).ok())
            .collect())
    }

    pub async fn recording_start(&self, label: Option<&str>) -> Result<Recording, DaytonaError> {
        self.client
            .post_json("/computer-use/recording/start", &json!({ "label": label }))
            .await
    }

    pub async fn recording_stop(&self, id: &str) -> Result<Recording, DaytonaError> {
        self.client
            .post_json("/computer-use/recording/stop", &json!({ "id": id }))
            .await
    }

    pub async fn recording_list(&self) -> Result<Vec<Recording>, DaytonaError> {
        let value: Value = self.client.get_json("/computer-use/recording/list").await?;
        Ok(value
            .get("recordings")
            .and_then(Value::as_array)
            .cloned()
            .unwrap_or_default()
            .into_iter()
            .filter_map(|v| serde_json::from_value(v).ok())
            .collect())
    }

    // CamelCase aliases
    #[allow(non_snake_case)]
    pub async fn mouseClick(
        &self,
        x: i32,
        y: i32,
        button: Option<&str>,
        double: Option<bool>,
    ) -> Result<Value, DaytonaError> {
        self.mouse_click(x, y, button, double).await
    }

    #[allow(non_snake_case)]
    pub async fn mouseMove(&self, x: i32, y: i32) -> Result<Value, DaytonaError> {
        self.mouse_move(x, y).await
    }

    #[allow(non_snake_case)]
    pub async fn mouseDrag(
        &self,
        start_x: i32,
        start_y: i32,
        end_x: i32,
        end_y: i32,
    ) -> Result<Value, DaytonaError> {
        self.mouse_drag(start_x, start_y, end_x, end_y).await
    }

    #[allow(non_snake_case)]
    pub async fn mouseScroll(
        &self,
        x: i32,
        y: i32,
        direction: &str,
        amount: Option<i32>,
    ) -> Result<Value, DaytonaError> {
        self.mouse_scroll(x, y, direction, amount).await
    }

    #[allow(non_snake_case)]
    pub async fn keyboardType(&self, text: &str) -> Result<(), DaytonaError> {
        self.keyboard_type(text).await
    }

    #[allow(non_snake_case)]
    pub async fn keyboardPress(&self, key: &str) -> Result<(), DaytonaError> {
        self.keyboard_press(key).await
    }

    #[allow(non_snake_case)]
    pub async fn keyboardHotkey(&self, keys: &str) -> Result<(), DaytonaError> {
        self.keyboard_hotkey(keys).await
    }

    #[allow(non_snake_case)]
    pub async fn displayInfo(&self) -> Result<DisplayInfo, DaytonaError> {
        self.display_info().await
    }

    #[allow(non_snake_case)]
    pub async fn displayWindows(&self) -> Result<Vec<WindowInfo>, DaytonaError> {
        self.display_windows().await
    }

    #[allow(non_snake_case)]
    pub async fn recordingStart(&self, label: Option<&str>) -> Result<Recording, DaytonaError> {
        self.recording_start(label).await
    }

    #[allow(non_snake_case)]
    pub async fn recordingStop(&self, id: &str) -> Result<Recording, DaytonaError> {
        self.recording_stop(id).await
    }

    #[allow(non_snake_case)]
    pub async fn recordingList(&self) -> Result<Vec<Recording>, DaytonaError> {
        self.recording_list().await
    }

    #[allow(non_snake_case)]
    pub async fn screenshotRegion(
        &self,
        region: &ScreenshotRegion,
        options: Option<&ScreenshotOptions>,
    ) -> Result<ScreenshotResponse, DaytonaError> {
        self.screenshot_region(region, options).await
    }
}
