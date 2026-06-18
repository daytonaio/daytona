# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import os

import litellm

resp = litellm.completion(
    model="openai/gpt-oss-20b",  # generic OpenAI-compatible provider; SGLang speaks that API
    api_base=f"{os.environ['ENDPOINT']}/v1",
    api_key="EMPTY",
    extra_headers={"x-daytona-preview-token": os.environ["TOKEN"]},
    messages=[{"role": "user", "content": "Write a haiku about calling a model that runs in the cloud."}],
    max_tokens=4096,  # covers reasoning plus answer
)
print(resp.choices[0].message.content)
