from fastapi import FastAPI, HTTPException, Depends, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import HTMLResponse
from fastapi.security.api_key import APIKeyHeader
from pydantic import BaseModel
from typing import Dict, TypedDict, List, Optional
import uvicorn
import os
import time
import requests as http_client
from collections import defaultdict
import google.generativeai as genai
from bs4 import BeautifulSoup

# ---- Config ----
API_KEY       = os.getenv("SCRIBE_API_KEY", "dev-scribe-key-unsafe")
GEMINI_API_KEY = os.getenv("GEMINI_API_KEY", "")
RATE_LIMIT_PER_MINUTE = int(os.getenv("RATE_LIMIT", "10"))

if not GEMINI_API_KEY:
    raise RuntimeError("GEMINI_API_KEY environment variable is required for Scribe agent.")

genai.configure(api_key=GEMINI_API_KEY)
model = genai.GenerativeModel("gemini-2.0-flash")

app = FastAPI(
    title="Scribe Content Agent API",
    description=(
        "An autonomous AI content generation agent. Given a topic, URL, or text, "
        "Scribe ingests the content, generates a punchy Twitter/X thread, a blog post draft, "
        "and key bullet points — ready to publish."
    ),
    version="1.0.0",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# ---- Security ----
api_key_header = APIKeyHeader(name="X-API-Key", auto_error=False)

async def verify_api_key(api_key: str = Depends(api_key_header)):
    if api_key != API_KEY:
        raise HTTPException(status_code=401, detail={"error": "UNAUTHORIZED", "message": "Invalid API Key."})
    return api_key

request_log: Dict[str, list] = defaultdict(list)

def rate_limit(request: Request):
    client_ip = request.client.host
    now = time.time()
    request_log[client_ip] = [t for t in request_log[client_ip] if t > now - 60]
    if len(request_log[client_ip]) >= RATE_LIMIT_PER_MINUTE:
        raise HTTPException(status_code=429, detail={"error": "RATE_LIMIT_EXCEEDED", "message": "Max 10 requests/minute (Gemini quota protection)."})
    request_log[client_ip].append(now)

# ---- Models ----
class ScribeInput(BaseModel):
    topic: Optional[str] = None           # A topic or headline to write about
    source_url: Optional[str] = None      # A URL to scrape and summarize
    raw_text: Optional[str] = None        # Raw text or whitepaper excerpt to process
    tone: Optional[str] = "professional"  # professional, casual, hype, academic
    audience: Optional[str] = "Web3 crypto Twitter"
    idempotency_key: Optional[str] = None

    class Config:
        json_schema_extra = {
            "example": {
                "topic": "Why autonomous AI agents will replace traditional DeFi bots",
                "tone": "hype",
                "audience": "Web3 crypto Twitter"
            }
        }

class AgentResponse(BaseModel):
    status: str
    agent: str
    data: dict
    error: Optional[dict] = None

class ScribeState(TypedDict):
    topic: Optional[str]
    source_url: Optional[str]
    raw_text: Optional[str]
    tone: str
    audience: str
    ingested_content: str
    twitter_thread: List[str]
    blog_post: str
    key_bullets: List[str]
    messages: List[str]

# ---- Helpers ----
def scrape_url(url: str) -> str:
    """Scrape and clean text from a URL."""
    resp = http_client.get(url, timeout=10, headers={"User-Agent": "Mozilla/5.0"})
    resp.raise_for_status()
    soup = BeautifulSoup(resp.text, "html.parser")
    # Remove scripts and styles
    for tag in soup(["script", "style", "nav", "footer", "header"]):
        tag.decompose()
    text = soup.get_text(separator=" ", strip=True)
    # Truncate to ~4000 chars to stay within Gemini's context
    return text[:4000]

# ---- LangGraph Nodes ----

def ingest_content(state: ScribeState) -> ScribeState:
    """Gather content from topic, URL, or raw text."""
    parts = []

    if state.get("source_url"):
        state["messages"].append(f"Scraping content from URL: {state['source_url']}")
        try:
            scraped = scrape_url(state["source_url"])
            parts.append(f"[Web Content]\n{scraped}")
            state["messages"].append(f"Scraped {len(scraped)} characters of content.")
        except Exception as e:
            state["messages"].append(f"URL scraping failed: {str(e)} — using topic fallback.")

    if state.get("raw_text"):
        parts.append(f"[Source Material]\n{state['raw_text'][:4000]}")
        state["messages"].append(f"Ingested {len(state['raw_text'])} chars of raw text.")

    if state.get("topic"):
        parts.append(f"[Core Topic]\n{state['topic']}")
        state["messages"].append(f"Topic: {state['topic']}")

    if not parts:
        state["ingested_content"] = ""
        state["messages"].append("ERROR: No content provided.")
        return state

    state["ingested_content"] = "\n\n".join(parts)
    state["messages"].append(f"Content ingested. Passing to Gemini for generation...")
    return state

def generate_thread(state: ScribeState) -> ScribeState:
    """Use Gemini to generate a punchy Twitter/X thread."""
    if not state["ingested_content"]:
        state["twitter_thread"] = []
        return state

    prompt = f"""You are Scribe, a world-class Web3 content strategist.
Based on the following content, write a punchy {state['tone']} 8-tweet Twitter/X thread for {state['audience']}.
Rules:
- Each tweet must be under 280 characters
- Tweet 1 must be a bold hook that stops the scroll  
- Use emojis sparingly but effectively
- End with a strong CTA (call to action)
- Number each tweet: 1/, 2/, etc.
- Do NOT include hashtags, they reduce reach

Content:
{state['ingested_content']}

Return ONLY the numbered tweets, one per line, no extra commentary."""

    try:
        response = model.generate_content(prompt)
        raw = response.text.strip()
        # Parse into individual tweets
        tweets = []
        for line in raw.split("\n"):
            line = line.strip()
            if line and (line[0].isdigit() or line.startswith("1/")):
                tweets.append(line)
        state["twitter_thread"] = tweets if tweets else raw.split("\n\n")
        state["messages"].append(f"Generated {len(state['twitter_thread'])}-tweet thread.")
    except Exception as e:
        state["twitter_thread"] = []
        state["messages"].append(f"Thread generation failed: {str(e)}")

    return state

def generate_blog_post(state: ScribeState) -> ScribeState:
    """Use Gemini to write a blog post and extract key bullets."""
    if not state["ingested_content"]:
        state["blog_post"] = ""
        state["key_bullets"] = []
        return state

    prompt = f"""You are Scribe, a {state['tone']} Web3 content writer.
Based on the following content, write:

1. A SHORT BLOG POST (300-400 words) with:
   - A compelling headline (H1)
   - 3 short paragraphs
   - A conclusion

2. 5 KEY BULLET POINTS (for LinkedIn/email summary)

Tone: {state['tone']}
Audience: {state['audience']}

Content:
{state['ingested_content']}

Format your response EXACTLY like this:
---BLOG---
[blog post here]
---BULLETS---
• [bullet 1]
• [bullet 2]
• [bullet 3]
• [bullet 4]
• [bullet 5]"""

    try:
        response = model.generate_content(prompt)
        raw = response.text.strip()

        blog_part = ""
        bullets_part = []

        if "---BLOG---" in raw and "---BULLETS---" in raw:
            blog_section = raw.split("---BLOG---")[1].split("---BULLETS---")[0].strip()
            bullets_section = raw.split("---BULLETS---")[1].strip()
            blog_part = blog_section
            bullets_part = [b.strip().lstrip("•").strip() for b in bullets_section.split("\n") if b.strip().startswith("•")]
        else:
            # Fallback if format not followed
            blog_part = raw
            bullets_part = []

        state["blog_post"] = blog_part
        state["key_bullets"] = bullets_part
        state["messages"].append(f"Blog post written ({len(blog_part)} chars). Extracted {len(bullets_part)} bullet points.")
    except Exception as e:
        state["blog_post"] = ""
        state["key_bullets"] = []
        state["messages"].append(f"Blog generation failed: {str(e)}")

    return state

# ---- LangGraph Compilation ----
from langgraph.graph import StateGraph, END

workflow = StateGraph(ScribeState)
workflow.add_node("ingest_content", ingest_content)
workflow.add_node("generate_thread", generate_thread)
workflow.add_node("generate_blog_post", generate_blog_post)
workflow.set_entry_point("ingest_content")
workflow.add_edge("ingest_content", "generate_thread")
workflow.add_edge("generate_thread", "generate_blog_post")
workflow.add_edge("generate_blog_post", END)
scribe_agent = workflow.compile()

executed_keys = {}

# ---- FastAPI Routes ----
@app.get("/health", tags=["System"])
async def health_check():
    return {"status": "ok", "agent": "scribe", "version": "1.0.0", "model": "gemini-2.0-flash"}

@app.post("/api/v1/execute", response_model=AgentResponse, tags=["Agent Execution"],
          summary="Generate a Twitter thread, blog post, and bullet points from any content",
          dependencies=[Depends(rate_limit)])
async def run_scribe(payload: ScribeInput):
    """
    Ingests a topic, URL, or raw text. Uses Gemini AI to generate a full content package:
    Twitter/X thread, blog post draft, and key bullet points.
    """
    if not payload.topic and not payload.source_url and not payload.raw_text:
        raise HTTPException(status_code=400, detail={"error": "BAD_REQUEST", "message": "Provide at least one of: topic, source_url, or raw_text."})

    if payload.idempotency_key and payload.idempotency_key in executed_keys:
        return AgentResponse(status="success", agent="scribe", data={**executed_keys[payload.idempotency_key], "cached": True})

    try:
        initial_state: ScribeState = {
            "topic": payload.topic,
            "source_url": payload.source_url,
            "raw_text": payload.raw_text,
            "tone": payload.tone or "professional",
            "audience": payload.audience or "Web3 crypto Twitter",
            "ingested_content": "",
            "twitter_thread": [],
            "blog_post": "",
            "key_bullets": [],
            "messages": []
        }
        final_state = scribe_agent.invoke(initial_state)
        result = {
            "twitter_thread": final_state["twitter_thread"],
            "blog_post": final_state["blog_post"],
            "key_bullets": final_state["key_bullets"],
            "messages": final_state["messages"]
        }
        if payload.idempotency_key:
            executed_keys[payload.idempotency_key] = result
        return AgentResponse(status="success", agent="scribe", data=result)
    except Exception as e:
        raise HTTPException(status_code=500, detail={"error": "AGENT_EXECUTION_FAILED", "message": str(e)})

@app.get("/", include_in_schema=False)
async def serve_dashboard():
    with open("index.html", "r") as f:
        return HTMLResponse(content=f.read())

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8004)
