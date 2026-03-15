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
from langgraph.graph import StateGraph, END

# ---- Config ----
API_KEY        = os.getenv("GUARDIAN_API_KEY", "dev-guardian-key")
GEMINI_API_KEY = os.getenv("GEMINI_API_KEY", "")
RATE_LIMIT_PER_MINUTE = int(os.getenv("RATE_LIMIT", "10"))

if not GEMINI_API_KEY:
    raise RuntimeError("GEMINI_API_KEY is required.")

genai.configure(api_key=GEMINI_API_KEY)
model = genai.GenerativeModel("gemini-2.0-flash")

app = FastAPI(title="Guardian Auditor API", version="1.0.0")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

api_key_header = APIKeyHeader(name="X-API-Key", auto_error=False)

async def verify_api_key(api_key: str = Depends(api_key_header)):
    if api_key != API_KEY: raise HTTPException(status_code=401, detail={"error": "UNAUTHORIZED"})
    return api_key

request_log: Dict[str, list] = defaultdict(list)
def rate_limit(request: Request):
    ip = request.client.host
    now = time.time()
    request_log[ip] = [t for t in request_log[ip] if t > now - 60]
    if len(request_log[ip]) >= RATE_LIMIT_PER_MINUTE: raise HTTPException(status_code=429)
    request_log[ip].append(now)

# ---- Models ----
class AuditInput(BaseModel):
    contract_address: str
    network: Optional[str] = "ethereum"
    idempotency_key: Optional[str] = None

class AgentResponse(BaseModel):
    status: str
    agent: str
    data: dict

class AuditState(TypedDict):
    input_params: dict
    source_code: str
    vulnerabilities: List[str]
    risk_score: int
    audit_report: str
    messages: List[str]

# ---- LangGraph Nodes ----
def fetch_code(state: AuditState) -> AuditState:
    state["messages"].append(f"Fetching verified source code for {state['input_params']['contract_address']}...")
    # Mocked fetch for demo
    state["source_code"] = "contract Token { mapping(address => uint) balances; function transfer(address to, uint val) { balances[msg.sender] -= val; balances[to] += val; } }"
    state["messages"].append("Source code retrieved (Mock implementation).")
    return state

def analyze_vulnerabilities(state: AuditState) -> AuditState:
    state["messages"].append("Running Gemini 2.0 Flash Security Analysis...")
    prompt = f"Analyze this Solidity snippet for common vulnerabilities (Reentrancy, Integer underflow, etc.). Return a score 1-100 (100=safest) and a brief report.\nCode:\n{state['source_code']}"
    
    try:
        res = model.generate_content(prompt).text
        state["audit_report"] = res
        state["risk_score"] = 45 # Mock parsed score based on the underflow
        state["vulnerabilities"] = ["Integer Underflow/Overflow", "Missing visibility specifier"]
        state["messages"].append("Audit complete. Vulnerabilities found.")
    except Exception as e:
        state["audit_report"] = f"Failed to analyze: {str(e)}"
        
    return state

workflow = StateGraph(AuditState)
workflow.add_node("fetch_code", fetch_code)
workflow.add_node("analyze_vulnerabilities", analyze_vulnerabilities)
workflow.set_entry_point("fetch_code")
workflow.add_edge("fetch_code", "analyze_vulnerabilities")
workflow.add_edge("analyze_vulnerabilities", END)
guardian_agent = workflow.compile()

# ---- Routes ----
@app.get("/health")
async def health(): return {"status": "ok", "agent": "guardian", "model": "gemini-2.0-flash"}

@app.post("/api/v1/execute", response_model=AgentResponse, dependencies=[Depends(rate_limit)])
async def run_audit(payload: AuditInput):
    state: AuditState = {"input_params": payload.dict(), "source_code": "", "vulnerabilities": [], "risk_score": 100, "audit_report": "", "messages": []}
    final = guardian_agent.invoke(state)
    return AgentResponse(status="success", agent="guardian", data=final)

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8007)
