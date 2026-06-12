# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import json
import os
import random

from openai import OpenAI

client = OpenAI(
    base_url=f"{os.environ['ENDPOINT']}/v1",
    api_key="EMPTY",  # vLLM doesn't check it; auth is the preview-token header
    default_headers={"x-daytona-preview-token": os.environ["TOKEN"]},
)
MODEL = "gemma-4-moe"

# plain chat
resp = client.chat.completions.create(
    model=MODEL,
    messages=[{"role": "user", "content": "Write a haiku about ephemeral sandboxes."}],
    max_tokens=64,
)
print("chat:", resp.choices[0].message.content)


# streaming
stream = client.chat.completions.create(
    model=MODEL,
    messages=[{"role": "user", "content": "Write ten haikus about tokens streaming from a sandbox."}],
    max_tokens=512,
    stream=True,
)
print("stream:")
for chunk in stream:
    print(chunk.choices[0].delta.content or "", end="", flush=True)
print()


# reasoning: gemma-4 generates reasoning tokens only when asked;
# reasoning_effort turns thinking mode on
resp = client.chat.completions.create(
    model=MODEL,
    messages=[{"role": "user", "content": "Write a haiku about GPU sandboxes."}],
    reasoning_effort="low",
    max_tokens=2048,
)
print("\nreasoning:")
print(resp.choices[0].message.reasoning)
print("answer:")
print(resp.choices[0].message.content)


# tool calling: the model emits a call, we run it, feed the result back, the
# model answers. Only the function body would change to run elsewhere (e.g. in
# a Daytona sandbox); the loop stays the same.
def get_weather(city):
    rng = random.Random(city.lower())  # same city, same weather
    temp = rng.randint(-5, 35)
    sky = rng.choice(["sunny", "cloudy", "rainy", "foggy", "windy"])
    return f"{temp}°C and {sky} in {city}"


tools = [
    {
        "type": "function",
        "function": {
            "name": "get_weather",
            "description": "Get the current weather for a city.",
            "parameters": {
                "type": "object",
                "properties": {"city": {"type": "string"}},
                "required": ["city"],
            },
        },
    }
]

messages = [{"role": "user", "content": "Write a haiku about the current weather in Paris."}]
resp = client.chat.completions.create(model=MODEL, messages=messages, tools=tools, max_tokens=256)
msg = resp.choices[0].message

if msg.tool_calls:
    messages.append(msg.model_dump(exclude_none=True))
    for call in msg.tool_calls:
        args = json.loads(call.function.arguments)
        result = get_weather(**args)
        print(f"\ntool call: {call.function.name}({args})")
        print(f"result:    {result}")
        messages.append({"role": "tool", "tool_call_id": call.id, "content": result})
    resp = client.chat.completions.create(model=MODEL, messages=messages, max_tokens=256)
    print("final:")
    print(resp.choices[0].message.content)
else:
    print("no tool call:", msg.content)
