# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import asyncio
import json
import os
import time
from pathlib import Path

import requests
from openai import AsyncOpenAI

client = AsyncOpenAI(
    base_url=f"{os.environ['ENDPOINT']}/v1",
    api_key="EMPTY",  # SGLang doesn't check it; auth is the preview-token header
    default_headers={"x-daytona-preview-token": os.environ["TOKEN"]},
)
MODEL = "gpt-oss-20b"
PASSAGES = 273
PASSAGE_CHARS = 12_000  # ~3k tokens

BOOKS = {
    "Austen": 1342,  # Pride and Prejudice
    "Bronte": 1260,  # Jane Eyre
    "Dickens": 98,  # A Tale of Two Cities
    "Doyle": 1661,  # The Adventures of Sherlock Holmes
    "Eliot": 145,  # Middlemarch
    "Hawthorne": 33,  # The Scarlet Letter
    "Melville": 2701,  # Moby Dick
    "Poe": 2148,  # The Works of Edgar Allan Poe, Vol. 2
    "Shelley": 84,  # Frankenstein
    "Stoker": 345,  # Dracula
    "Twain": 76,  # Adventures of Huckleberry Finn
    "Wells": 36,  # The War of the Worlds
    "Wilde": 174,  # The Picture of Dorian Gray
}
AUTHORS = sorted(BOOKS)

# The passage leads and the question trails, so the ~3k-token passage is the
# cached prefix: a second question over the same passages reuses it, and only
# the short trailing question is recomputed.
AUTHOR_QUESTION = f"Which of these authors wrote this passage: {', '.join(AUTHORS)}?"
SETTING_QUESTION = "Is this scene set indoors or outdoors?"

AUTHOR_SCHEMA = {
    "type": "object",
    "properties": {"author": {"type": "string", "enum": AUTHORS}},
    "required": ["author"],
}
SETTING_SCHEMA = {
    "type": "object",
    "properties": {"setting": {"type": "string", "enum": ["indoors", "outdoors"]}},
    "required": ["setting"],
}


def passages_from(author, book_id, count):
    cached = Path("gutenberg_cache") / f"pg{book_id}.txt"
    if cached.exists():
        text = cached.read_text()
    else:
        response = requests.get(f"https://www.gutenberg.org/cache/epub/{book_id}/pg{book_id}.txt", timeout=60)
        response.raise_for_status()
        text = response.text
        cached.parent.mkdir(exist_ok=True)
        cached.write_text(text)
    # keep only the body: drop the Gutenberg boilerplate, title page, and license
    body = text.split("*** START")[1].split("*** END")[0]
    body = body[len(body) // 10 : -len(body) // 10]
    step = (len(body) - PASSAGE_CHARS) // count
    return [(author, body[i * step : i * step + PASSAGE_CHARS]) for i in range(count)]


async def classify(passage, question, schema):
    resp = await client.chat.completions.create(
        model=MODEL,
        messages=[{"role": "user", "content": f"{passage}\n\n{question} Reply as JSON."}],
        response_format={"type": "json_schema", "json_schema": {"name": "answer", "schema": schema}},
        reasoning_effort="low",
        max_tokens=2048,
    )
    content = resp.choices[0].message.content  # None if thinking hit max_tokens
    return (json.loads(content) if content else None), resp.usage


def print_tokens(usages, dt, cached=False):
    in_tok = sum(u.prompt_tokens for u in usages)
    out_tok = sum(u.completion_tokens for u in usages)
    rate = f"{in_tok / dt:,.0f} tok/s, {in_tok / dt * 3.6 / 1e3:.1f}M/hour"
    print(f"in:  {in_tok:,} tok ({rate}{' including cache hits' if cached else ''})")
    print(f"out: {out_tok:,} tok ({out_tok / dt:,.0f} tok/s)")
    if cached:
        hits = sum((u.prompt_tokens_details.cached_tokens if u.prompt_tokens_details else 0) for u in usages)
        print(f"cached: {hits:,}/{in_tok:,} prompt tokens from cache")


def load_passages():
    dataset = []
    for author, book_id in BOOKS.items():
        dataset.extend(passages_from(author, book_id, PASSAGES // len(BOOKS)))
    return dataset


async def main(passages):
    # pass 1: attribute each passage to an author (cold cache, the passages are new)
    t0 = time.perf_counter()
    by_author = await asyncio.gather(*(classify(p, AUTHOR_QUESTION, AUTHOR_SCHEMA) for p in passages))
    dt1 = time.perf_counter() - t0
    # pass 2: a different question over the same passages, now served from cache
    t0 = time.perf_counter()
    by_setting = await asyncio.gather(*(classify(p, SETTING_QUESTION, SETTING_SCHEMA) for p in passages))
    dt2 = time.perf_counter() - t0
    return by_author, dt1, by_setting, dt2


def report():
    print(f"loading {len(BOOKS)} books (downloaded from Project Gutenberg on first run) ...")
    dataset = load_passages()
    passages = [p for _, p in dataset]
    truths = [a for a, _ in dataset]

    print(f"classifying {len(passages)} passages, two questions each ...")
    by_author, dt1, by_setting, dt2 = asyncio.run(main(passages))

    authors = [r["author"] if r else None for r, _ in by_author]
    correct = sum(g == t for g, t in zip(authors, truths))
    print(f"\npass 1 - author: {len(passages)} passages in {dt1:.1f}s")
    print(f"accuracy: {correct}/{len(passages)} ({100 * correct / len(passages):.0f}%)")
    per_author = {a: [0, 0] for a in AUTHORS}
    for g, t in zip(authors, truths):
        per_author[t][0] += g == t
        per_author[t][1] += 1
    print("per author: " + ", ".join(f"{a} {ok}/{n}" for a, (ok, n) in per_author.items()))
    print_tokens([u for _, u in by_author], dt1)

    settings = [r["setting"] if r else None for r, _ in by_setting]
    print(f"\npass 2 - setting: {len(passages)} passages in {dt2:.1f}s ({dt1 / dt2:.1f}x faster)")
    # group books by the model's predominant call (not graded, just its read)
    lean = {}
    for a in AUTHORS:
        calls = [s for t, s in zip(truths, settings) if t == a and s is not None]
        if not calls:
            continue
        top = max(set(calls), key=calls.count)
        lean[a] = (top, calls.count(top), len(calls))
    for side in ("indoors", "outdoors"):
        books = sorted((a for a, v in lean.items() if v[0] == side), key=lambda a: -lean[a][1])
        print(f"predominantly {side + ':':<9} " + ", ".join(f"{a} {lean[a][1]}/{lean[a][2]}" for a in books))
    print_tokens([u for _, u in by_setting], dt2, cached=True)


if __name__ == "__main__":
    report()
