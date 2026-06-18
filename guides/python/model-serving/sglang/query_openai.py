# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import json
import os
import random
import time

from openai import OpenAI

client = OpenAI(
    base_url=f"{os.environ['ENDPOINT']}/v1",
    api_key="EMPTY",  # SGLang doesn't check it; auth is the preview-token header
    default_headers={"x-daytona-preview-token": os.environ["TOKEN"]},
)
MODEL = "gpt-oss-20b"

# plain chat. max_tokens covers reasoning plus answer; gpt-oss thinks before
# it speaks, so budget for both.
resp = client.chat.completions.create(
    model=MODEL,
    messages=[{"role": "user", "content": "Write a haiku about a sandbox that vanishes when the work is done."}],
    max_tokens=4096,
)
print("chat:", resp.choices[0].message.content)


# streaming: gpt-oss streams its reasoning first, then the answer, so print both
stream = client.chat.completions.create(
    model=MODEL,
    messages=[{"role": "user", "content": "Write ten haikus about tokens arriving one at a time."}],
    max_tokens=8192,
    stream=True,
)
print("stream:")
for chunk in stream:
    delta = chunk.choices[0].delta
    print(delta.reasoning_content or delta.content or "", end="", flush=True)
print()


# structured output: response_format constrains decoding to the schema, so the
# reply is guaranteed to parse. The prompt still describes the shape: a model
# that tries to answer in prose can stall on whitespace (the only thing the
# grammar allows it) until max_tokens cuts the JSON short.
schema = {
    "type": "object",
    "properties": {
        "title": {"type": "string"},
        "lines": {"type": "array", "items": {"type": "string"}, "minItems": 3, "maxItems": 3},
        "season": {"type": "string"},
    },
    "required": ["title", "lines", "season"],
}
resp = client.chat.completions.create(
    model=MODEL,
    messages=[
        {
            "role": "user",
            "content": "Compose a haiku about GPU sandboxes, as JSON with title, lines, and season.",
        }
    ],
    response_format={"type": "json_schema", "json_schema": {"name": "haiku", "schema": schema}},
    max_tokens=4096,
)
content = resp.choices[0].message.content
if content:
    haiku = json.loads(content)
    print(f"\nstructured: {haiku['title']} ({haiku['season']})")
    for line in haiku["lines"]:
        print(f"  {line}")
else:
    print("\nstructured: response truncated (raise max_tokens)")


# reasoning: gpt-oss thinks by default at medium effort; reasoning_effort
# turns it up or down, and the parsed trace comes back in reasoning_content
resp = client.chat.completions.create(
    model=MODEL,
    messages=[{"role": "user", "content": "Write a haiku about thinking before speaking."}],
    reasoning_effort="high",
    max_tokens=8192,
)
print("\nreasoning:")
print(resp.choices[0].message.reasoning_content)
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

messages = [{"role": "user", "content": "Write a haiku about the current weather in Lisbon."}]
resp = client.chat.completions.create(model=MODEL, messages=messages, tools=tools, max_tokens=4096)
msg = resp.choices[0].message

if msg.tool_calls:
    messages.append(msg.model_dump(exclude_none=True))
    for call in msg.tool_calls:
        args = json.loads(call.function.arguments)
        result = get_weather(**args)
        print(f"\ntool call: {call.function.name}({args})")
        print(f"result:    {result}")
        messages.append({"role": "tool", "tool_call_id": call.id, "content": result})
    resp = client.chat.completions.create(model=MODEL, messages=messages, max_tokens=4096)
    print("final:")
    print(resp.choices[0].message.content)
else:
    print("no tool call:", msg.content)


# prefix caching: RadixAttention reuses the KV cache of any shared prompt
# prefix; the server's --enable-cache-report flag exposes the hit count in
# usage.prompt_tokens_details
context = "The Daytona platform provides isolated sandboxes for AI agents to safely execute code. " * 60
print("\nprefix cache:")
for attempt in (1, 2):
    t0 = time.perf_counter()
    resp = client.chat.completions.create(
        model=MODEL,
        messages=[{"role": "user", "content": context + "Summarize the above in one sentence."}],
        max_tokens=32,
    )
    dt = time.perf_counter() - t0
    details = resp.usage.prompt_tokens_details  # omitted entirely on a cold cache
    cached = details.cached_tokens if details else 0
    print(f"  attempt {attempt}: {dt:.2f}s, {cached}/{resp.usage.prompt_tokens} prompt tokens from cache")
