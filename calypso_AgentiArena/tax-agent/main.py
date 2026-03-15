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

API_KEY        = os.getenv("TAX_API_KEY", "dev-tax-key")
GEMINI_API_KEY = os.getenv("GEMINI_API_KEY", "")
RATE_LIMIT_PER_MINUTE = int(os.getenv("RATE_LIMIT", "10"))

if not GEMINI_API_KEY: raise RuntimeError("GEMINI_API_KEY is required.")
genai.configure(api_key=GEMINI_API_KEY)
model = genai.GenerativeModel("gemini-2.0-flash")

app = FastAPI(title="Tax Reporter API", version="1.0.0")

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

class TaxInput(BaseModel):
    wallet: str
    taxYear: str = "2025"
    jurisdiction: str = "US"
    idempotency_key: Optional[str] = None

class AgentResponse(BaseModel):
    status: str
    agent: str
    data: dict

class TaxState(TypedDict):
    input_params: dict
    raw_data: str
    tax_report: str
    messages: List[str]

def gather_ledger(state: TaxState) -> TaxState:
    state["messages"].append(f"Pulling ledger for {state['input_params']['wallet']} in year {state['input_params']['taxYear']}...")
    state["raw_data"] = "Mocked Ledger: 450 Swaps, 12 Bridge Txs, Net Profit $14,500. Airdrop received: $1,200."
    return state

def generate_report(state: TaxState) -> TaxState:
    state["messages"].append("Running Gemini 2.0 Tax Categorizer...")
    prompt = f"Act as a Crypto CPA for {state['input_params']['jurisdiction']}. Analyze this ledger: {state['raw_data']}. Provide estimated Short Term Capital Gains vs Income, and a legally compliant summary paragraph."
    try:
        res = model.generate_content(prompt).text
        state["tax_report"] = res
        state["messages"].append("Tax Document Generated.")
    except Exception as e:
        state["tax_report"] = str(e)
    return state

workflow = StateGraph(TaxState)
workflow.add_node("gather_ledger", gather_ledger)
workflow.add_node("generate_report", generate_report)
workflow.set_entry_point("gather_ledger")
workflow.add_edge("gather_ledger", "generate_report")
workflow.add_edge("generate_report", END)
tax_agent = workflow.compile()

@app.get("/health")
async def health(): return {"status": "ok", "agent": "tax-reporter"}

@app.post("/api/v1/execute", response_model=AgentResponse, dependencies=[Depends(rate_limit)])
async def generate(payload: TaxInput):
    state: TaxState = {"input_params": payload.dict(), "raw_data": "", "tax_report": "", "messages": []}
    return AgentResponse(status="success", agent="tax-reporter", data=tax_agent.invoke(state))

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8009)
