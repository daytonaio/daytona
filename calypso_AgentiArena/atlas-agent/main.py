from fastapi import FastAPI, HTTPException, Depends, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import HTMLResponse
from fastapi.security.api_key import APIKeyHeader
from pydantic import BaseModel
from typing import Dict, TypedDict, List, Optional
import uvicorn
import os
import time
import requests
from collections import defaultdict

# ---- Configuration ----
API_KEY = os.getenv("ATLAS_API_KEY", "dev-atlas-key-unsafe")
RATE_LIMIT_PER_MINUTE = int(os.getenv("RATE_LIMIT", "30"))

COINGECKO_PRICE_URL = "https://api.coingecko.com/api/v3/simple/price"

app = FastAPI(
    title="Atlas Portfolio Rebalancer API",
    description="Autonomous rebalancing agent. Uses live CoinGecko prices to detect portfolio drift and generate precise DEX swap payloads.",
    version="2.0.0",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


api_key_header = APIKeyHeader(name="X-API-Key", auto_error=False)

async def verify_api_key(api_key: str = Depends(api_key_header)):
    if api_key != API_KEY:
        raise HTTPException(status_code=401, detail={"error": "UNAUTHORIZED", "message": "Invalid or missing API Key."})
    return api_key

request_log: Dict[str, list] = defaultdict(list)

def rate_limit(request: Request):
    client_ip = request.client.host
    now = time.time()
    window_start = now - 60
    request_log[client_ip] = [t for t in request_log[client_ip] if t > window_start]
    if len(request_log[client_ip]) >= RATE_LIMIT_PER_MINUTE:
        raise HTTPException(status_code=429, detail={"error": "RATE_LIMIT_EXCEEDED"})
    request_log[client_ip].append(now)

class BalancesInput(BaseModel):
    ETH: float
    BTC: float
    USDC: float
    idempotency_key: Optional[str] = None

    class Config:
        json_schema_extra = {"example": {"ETH": 2.5, "BTC": 0.5, "USDC": 10000}}

class AgentResponse(BaseModel):
    status: str
    agent: str
    data: dict
    error: Optional[dict] = None

class PortfolioState(TypedDict):
    current_balances: Dict[str, float]
    target_allocation: Dict[str, float]
    live_prices: Dict[str, float]
    drift_percentage: float
    required_trades: List[Dict]
    messages: List[str]

# ---- LangGraph Nodes ----
def fetch_prices(state: PortfolioState) -> PortfolioState:
    """Fetch live ETH, BTC, USDC prices from CoinGecko API."""
    state["messages"].append("Fetching live prices from CoinGecko API...")
    try:
        resp = requests.get(
            COINGECKO_PRICE_URL,
            params={"ids": "ethereum,bitcoin,usd-coin", "vs_currencies": "usd"},
            timeout=8
        )
        resp.raise_for_status()
        data = resp.json()
        state["live_prices"] = {
            "ETH": data["ethereum"]["usd"],
            "BTC": data["bitcoin"]["usd"],
            "USDC": data["usd-coin"]["usd"]
        }
        state["messages"].append(
            f"Live prices — ETH: ${data['ethereum']['usd']:,.2f} | "
            f"BTC: ${data['bitcoin']['usd']:,.2f} | "
            f"USDC: ${data['usd-coin']['usd']:.4f}"
        )
    except Exception as e:
        state["messages"].append(f"ERROR: CoinGecko API failed — {str(e)}. Using fallback prices.")
        state["live_prices"] = {"ETH": 3500.0, "BTC": 65000.0, "USDC": 1.0}
    return state

def calculate_drift(state: PortfolioState) -> PortfolioState:
    prices = state["live_prices"]
    balances = state["current_balances"]
    target = state["target_allocation"]
    total_value = sum(balances[k] * prices[k] for k in balances)
    if total_value == 0:
        state["drift_percentage"] = 0.0
        state["messages"].append("Portfolio value is $0. Nothing to rebalance.")
        return state
    current_alloc = {k: (balances[k] * prices[k]) / total_value for k in balances}
    max_drift = max(abs(current_alloc[k] - target[k]) for k in target)
    state["drift_percentage"] = max_drift
    state["messages"].append(f"Total Portfolio Value: ${total_value:,.2f}")
    for k in target:
        state["messages"].append(
            f"  {k}: ${balances[k] * prices[k]:,.2f} = {current_alloc[k]*100:.1f}% (Target {target[k]*100:.1f}%) Δ{abs(current_alloc[k]-target[k])*100:.1f}%"
        )
    state["messages"].append(f"Max Drift: {max_drift*100:.2f}% {'⚠ REBALANCE NEEDED' if max_drift > 0.05 else '✓ WITHIN RANGE'}")
    return state

def generate_trade_payload(state: PortfolioState) -> PortfolioState:
    if state["drift_percentage"] <= 0.05:
        state["required_trades"] = []
        state["messages"].append("Portfolio balanced. No trades required.")
        return state
        
    prices = state["live_prices"]
    balances = state["current_balances"]
    target = state["target_allocation"]
    total_value = sum(balances[k] * prices[k] for k in balances)
    current_alloc = {k: (balances[k] * prices[k]) / total_value for k in balances}
    
    # Store token excesses/deficits as strictly typed floats explicitly to prevent mixed str/Literal inference
    over = {k: float((current_alloc[k] - target[k]) * total_value) for k in target if (current_alloc[k] - target[k]) * total_value > 5}
    under = {k: float(abs((current_alloc[k] - target[k]) * total_value)) for k in target if (current_alloc[k] - target[k]) * total_value < -5}
    
    trades = []
    for o_token, excess in list(over.items()):
        for u_token, deficit in list(under.items()):
            if over.get(o_token, 0.0) <= 0 or under.get(u_token, 0.0) <= 0:
                continue
            
            amount = float(min(over[o_token], under[u_token]))
            trades.append({
                "action": "SWAP", "route": "1inch Aggregator",
                "fromToken": o_token, "toToken": u_token,
                "amountUSD": float(round(amount, 2)),
                "amountFrom": float(round(amount / prices[o_token], 6)),
                "fromPrice": float(prices[o_token]),
                "estimatedSlippage": "0.1%"
            })
            
            over[o_token] -= amount
            under[u_token] -= amount
            
    state["required_trades"] = trades  # type: ignore
    for t in trades:
        state["messages"].append(f"SWAP {t['amountFrom']} {t['fromToken']} → {t['toToken']} (${t['amountUSD']:,.2f})")
    
    return state

from langgraph.graph import StateGraph, END

workflow = StateGraph(PortfolioState)
workflow.add_node("fetch_prices", fetch_prices)
workflow.add_node("calculate_drift", calculate_drift)
workflow.add_node("generate_trade_payload", generate_trade_payload)
workflow.set_entry_point("fetch_prices")
workflow.add_edge("fetch_prices", "calculate_drift")
workflow.add_edge("calculate_drift", "generate_trade_payload")
workflow.add_edge("generate_trade_payload", END)
atlas_agent = workflow.compile()

executed_keys = {}

@app.get("/health", tags=["System"])
async def health_check():
    return {"status": "ok", "agent": "atlas", "version": "2.0.0", "data_source": "CoinGecko live API"}

@app.post("/api/v1/execute", response_model=AgentResponse, tags=["Agent Execution"],
          summary="Rebalance a portfolio using live CoinGecko price data",
          dependencies=[Depends(rate_limit)])
async def run_rebalance(payload: BalancesInput):
    """Calculates real-time portfolio drift using live ETH/BTC prices and generates rebalancing trades."""
    if payload.idempotency_key and payload.idempotency_key in executed_keys:
        # Pass payload as dict to satisfy strict intellisense when BaseModel isn't fully loaded
        return AgentResponse(**{"status": "success", "agent": "atlas", "data": {**executed_keys[payload.idempotency_key], "cached": True}})
    try:
        initial_state: PortfolioState = {
            "current_balances": {"ETH": payload.ETH, "BTC": payload.BTC, "USDC": payload.USDC},
            "target_allocation": {"ETH": 0.40, "BTC": 0.40, "USDC": 0.20},
            "live_prices": {}, "drift_percentage": 0.0, "required_trades": [], "messages": []
        }
        final_state = atlas_agent.invoke(initial_state)
        result = {
            "live_prices": final_state["live_prices"],
            "drift_percentage": final_state["drift_percentage"],
            "required_trades": final_state["required_trades"],
            "messages": final_state["messages"]
        }
        if payload.idempotency_key:
            executed_keys[payload.idempotency_key] = result
        return AgentResponse(**{"status": "success", "agent": "atlas", "data": result})
    except Exception as e:
        raise HTTPException(status_code=500, detail={"error": "AGENT_EXECUTION_FAILED", "message": str(e)})

@app.get("/", include_in_schema=False)
async def serve_dashboard():
    with open("index.html", "r") as f:
        return HTMLResponse(content=f.read())

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
