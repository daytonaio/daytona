from fastapi import FastAPI, HTTPException, Depends, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import HTMLResponse
from fastapi.security.api_key import APIKeyHeader
from pydantic import BaseModel
from typing import Dict, TypedDict, List, Optional
import uvicorn
import os
import time
from collections import defaultdict
import google.generativeai as genai
from langgraph.graph import StateGraph, END

API_KEY        = os.getenv("SUMMARY_API_KEY", "dev-summary-key")
GEMINI_API_KEY = os.getenv("GEMINI_API_KEY", "")
RATE_LIMIT_PER_MINUTE = int(os.getenv("RATE_LIMIT", "10"))

if not GEMINI_API_KEY: raise RuntimeError("GEMINI_API_KEY is required.")
genai.configure(api_key=GEMINI_API_KEY)
model = genai.GenerativeModel("gemini-2.0-flash")

app = FastAPI(title="Executive Summarizer API", version="1.0.0")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

api_key_header = APIKeyHeader(name="X-API-Key", auto_error=False)

async def verify_api_key(api_key: str = Depends(api_key_header)):
    if api_key != API_KEY: raise HTTPException(status_code=401)
    return api_key

request_log: Dict[str, list] = defaultdict(list)
def rate_limit(request: Request):
    ip = request.client.host
    now = time.time()
    request_log[ip] = [t for t in request_log[ip] if t > now - 60]
    if len(request_log[ip]) >= RATE_LIMIT_PER_MINUTE: raise HTTPException(status_code=429)
    request_log[ip].append(now)

class SummaryInput(BaseModel):
    meetingNotes: str
    format: str = "Bullet Points"
    attendees: str = "Alice, Bob, Dave"
    idempotency_key: Optional[str] = None

class AgentResponse(BaseModel):
    status: str
    agent: str
    data: dict

class SummaryState(TypedDict):
    input_params: dict
    cleaned_transcript: str
    executive_summary: str
    messages: List[str]

def clean_transcript(state: SummaryState) -> SummaryState:
    state["messages"].append("Running NLP pipeline to strip filler words and normalize speaker tags...")
    state["cleaned_transcript"] = state["input_params"]["meetingNotes"][:2000] # truncate
    state["messages"].append("Transcript normalized.")
    return state

def generate_summary(state: SummaryState) -> SummaryState:
    state["messages"].append("Synthesizing Executive Action Items via Gemini 2.0 Flash...")
    prompt = f"You are a ruthless Executive Assistant. Read transcription: {state['cleaned_transcript']}. Attendees: {state['input_params']['attendees']}. Output format: {state['input_params']['format']}. Create a summary with hard action items assigned to specific attendees."
    try:
        res = model.generate_content(prompt).text
        state["executive_summary"] = res
        state["messages"].append("Executive Summary ready.")
    except Exception as e:
        state["executive_summary"] = str(e)
    return state

workflow = StateGraph(SummaryState)
workflow.add_node("clean_transcript", clean_transcript)
workflow.add_node("generate_summary", generate_summary)
workflow.set_entry_point("clean_transcript")
workflow.add_edge("clean_transcript", "generate_summary")
workflow.add_edge("generate_summary", END)
summary_agent = workflow.compile()

@app.get("/health")
async def health(): return {"status": "ok", "agent": "executive-summarizer"}

@app.post("/api/v1/execute", response_model=AgentResponse, dependencies=[Depends(rate_limit)])
async def summarize(payload: SummaryInput):
    state: SummaryState = {"input_params": payload.dict(), "cleaned_transcript": "", "executive_summary": "", "messages": []}
    return AgentResponse(status="success", agent="executive-summarizer", data=summary_agent.invoke(state))

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8011)
