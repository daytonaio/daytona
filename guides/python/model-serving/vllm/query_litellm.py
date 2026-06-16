# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import os

import litellm

resp = litellm.completion(
    model="hosted_vllm/gemma-4-moe",  # OpenAI-compatible vLLM server
    api_base=f"{os.environ['ENDPOINT']}/v1",
    api_key="EMPTY",
    extra_headers={"x-daytona-preview-token": os.environ["TOKEN"]},
    messages=[{"role": "user", "content": "Write a haiku about agents running code in the cloud."}],
    max_tokens=64,
)
print(resp.choices[0].message.content)
