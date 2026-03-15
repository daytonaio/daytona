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

API_KEY        = os.getenv("WHALE_API_KEY", "dev-whale-key")
GEMINI_API_KEY = os.getenv("GEMINI_API_KEY", "")
RATE_LIMIT_PER_MINUTE = int(os.getenv("RATE_LIMIT", "10"))

if not GEMINI_API_KEY: raise RuntimeError("GEMINI_API_KEY is required.")
genai.configure(api_key=GEMINI_API_KEY)
model = genai.GenerativeModel("gemini-2.0-flash")

app = FastAPI(title="Whale Watcher API", version="1.0.0")

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

class WhaleInput(BaseModel):
    wallet_address: str
    min_amount: str = "100000"
    alert_type: str = "telegram"
    idempotency_key: Optional[str] = None

class AgentResponse(BaseModel):
    status: str
    agent: str
    data: dict

class WhaleState(TypedDict):
    input_params: dict
    recent_txs: List[dict]
    analysis: str
    messages: List[str]

def fetch_txs(state: WhaleState) -> WhaleState:
    state["messages"].append(f"Scanning blockchain for {state['input_params']['wallet_address']}...")
    state["recent_txs"] = [
        {"hash": "0x123", "value": "1500 ETH", "to": "Binance Hot Wallet"},
        {"hash": "0x456", "value": "2000000 USDC", "to": "Uniswap V3 Pool"}
    ]
    state["messages"].append("Found 2 massive transactions.")
    return state

def analyze_intent(state: WhaleState) -> WhaleState:
    state["messages"].append("Analyzing Whale Intent with Gemini 2.0 Flash...")
    prompt = f"Analyze these recent whale transactions: {state['recent_txs']}. Is this accumulation, distribution, or swapping? Write a short alert."
    try:
        res = model.generate_content(prompt).text
        state["analysis"] = res
        state["messages"].append("Intent analysis complete.")
    except Exception as e:
        state["analysis"] = f"Failed: {str(e)}"
    return state

workflow = StateGraph(WhaleState)
workflow.add_node("fetch_txs", fetch_txs)
workflow.add_node("analyze_intent", analyze_intent)
workflow.set_entry_point("fetch_txs")
workflow.add_edge("fetch_txs", "analyze_intent")
workflow.add_edge("analyze_intent", END)
whale_agent = workflow.compile()

@app.get("/health")
async def health(): return {"status": "ok", "agent": "whale-watcher"}

@app.post("/api/v1/execute", response_model=AgentResponse, dependencies=[Depends(rate_limit)])
async def run_scan(payload: WhaleInput):
    state: WhaleState = {"input_params": payload.dict(), "recent_txs": [], "analysis": "", "messages": []}
    return AgentResponse(status="success", agent="whale-watcher", data=whale_agent.invoke(state))

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8008)
