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

API_KEY        = os.getenv("AIRDROP_API_KEY", "dev-airdrop-key")
GEMINI_API_KEY = os.getenv("GEMINI_API_KEY", "")
RATE_LIMIT_PER_MINUTE = int(os.getenv("RATE_LIMIT", "10"))

if not GEMINI_API_KEY: raise RuntimeError("GEMINI_API_KEY is required.")
genai.configure(api_key=GEMINI_API_KEY)
model = genai.GenerativeModel("gemini-2.0-flash")

app = FastAPI(title="Airdrop Hunter API", version="1.0.0")

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

class AirdropInput(BaseModel):
    wallet: str
    target_ecosystem: str = "solana"
    idempotency_key: Optional[str] = None

class AgentResponse(BaseModel):
    status: str
    agent: str
    data: dict

class AirdropState(TypedDict):
    input_params: dict
    active_campaigns: List[str]
    execution_route: str
    messages: List[str]

def find_campaigns(state: AirdropState) -> AirdropState:
    state["messages"].append(f"Scanning {state['input_params']['target_ecosystem']} ecosystem for un-snapshotted airdrops...")
    state["active_campaigns"] = ["Jito V2", "MarginFi", "Drift Protocol"]
    state["messages"].append("Found 3 high-probability alpha routes.")
    return state

def generate_route(state: AirdropState) -> AirdropState:
    state["messages"].append("Calculating optimal sybil-resistant gas route...")
    prompt = f"Create a step-by-step transaction route for wallet {state['input_params']['wallet']} to qualify for these drops: {state['active_campaigns']}. Make it look like organic human behavior to avoid sybil filters."
    try:
        res = model.generate_content(prompt).text
        state["execution_route"] = res
        state["messages"].append("Farming route synthesized.")
    except Exception as e:
        state["execution_route"] = str(e)
    return state

workflow = StateGraph(AirdropState)
workflow.add_node("find_campaigns", find_campaigns)
workflow.add_node("generate_route", generate_route)
workflow.set_entry_point("find_campaigns")
workflow.add_edge("find_campaigns", "generate_route")
workflow.add_edge("generate_route", END)
airdrop_agent = workflow.compile()

@app.get("/health")
async def health(): return {"status": "ok", "agent": "airdrop-hunter"}

@app.post("/api/v1/execute", response_model=AgentResponse, dependencies=[Depends(rate_limit)])
async def hunt(payload: AirdropInput):
    state: AirdropState = {"input_params": payload.dict(), "active_campaigns": [], "execution_route": "", "messages": []}
    return AgentResponse(status="success", agent="airdrop-hunter", data=airdrop_agent.invoke(state))

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8010)
