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
API_KEY = os.getenv("HARVESTER_API_KEY", "dev-harvester-key-unsafe")
RATE_LIMIT_PER_MINUTE = int(os.getenv("RATE_LIMIT", "30"))

DEFILLAMA_URL = "https://yields.llama.fi/pools"

app = FastAPI(
    title="Harvester Yield Optimizer API",
    description="Autonomous yield farming agent. Fetches live APY data from DeFiLlama across Aave, Compound, Curve and Yearn to find the optimal pool for your deposit.",
    version="2.0.0",
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

# ---- Models ----
class HarvestInput(BaseModel):
    deposit_token: str
    deposit_amount_usd: float
    chain: Optional[str] = None  # e.g. "Ethereum", "Polygon" — None means all chains
    idempotency_key: Optional[str] = None

    class Config:
        json_schema_extra = {
            "example": {"deposit_token": "USDC", "deposit_amount_usd": 5000, "chain": "Ethereum"}
        }

class AgentResponse(BaseModel):
    status: str
    agent: str
    data: dict
    error: Optional[dict] = None

class HarvesterState(TypedDict):
    deposit_token: str
    deposit_amount_usd: float
    chain: Optional[str]
    available_pools: List[Dict]
    best_pool: Dict
    ranked_pools: List[Dict]
    compounding_plan: Dict
    messages: List[str]

# ---- TRUSTED protocols filter (product-grade, no scam pools) ----
TRUSTED_PROTOCOLS = {
    "aave-v3", "aave-v2", "compound-v3", "compound-v2",
    "curve-dex", "yearn-finance", "lido", "rocket-pool",
    "makerdao", "uniswap-v3", "convex-finance"
}

GAS_COST_USD = 12.0

# ---- LangGraph Nodes ----
def scan_protocols(state: HarvesterState) -> HarvesterState:
    """Fetch live pool data from DeFiLlama yields API."""
    token = state["deposit_token"].upper()
    chain_filter = state.get("chain")

    state["messages"].append(f"Fetching live pool data from DeFiLlama API...")

    try:
        resp = requests.get(DEFILLAMA_URL, timeout=10)
        resp.raise_for_status()
        all_pools = resp.json().get("data", [])
    except Exception as e:
        state["messages"].append(f"ERROR: DeFiLlama API failed — {str(e)}")
        state["available_pools"] = []
        return state

    state["messages"].append(f"Received {len(all_pools)} pools. Filtering for {token}...")

    matching = []
    for pool in all_pools:
        pool_symbol = (pool.get("symbol") or "").upper()
        protocol = (pool.get("project") or "").lower()
        chain = (pool.get("chain") or "").lower()
        apy = pool.get("apy") or 0

        # Filter: token must be in pool symbol, protocol must be trusted, APY must be positive
        if token not in pool_symbol:
            continue
        if protocol not in TRUSTED_PROTOCOLS:
            continue
        if apy <= 0 or apy > 200:  # Reject obvious scams / outliers
            continue
        if chain_filter and chain_filter.lower() not in chain:
            continue

        matching.append({
            "protocol": pool.get("project", "Unknown"),
            "pool": pool.get("symbol", "Unknown"),
            "chain": pool.get("chain", "Unknown"),
            "apy": round(apy, 2),
            "tvl_usd": pool.get("tvlUsd", 0),
            "risk": "Low" if apy < 10 else "Medium" if apy < 25 else "High",
            "pool_id": pool.get("pool", "")
        })

    # Sort by APY, take top 8
    matching = sorted(matching, key=lambda x: x["apy"], reverse=True)[:8]

    state["available_pools"] = matching
    state["messages"].append(f"Found {len(matching)} live pools matching {token} from trusted protocols.")
    for p in matching[:5]:
        state["messages"].append(f"  [{p['chain']}] {p['protocol'].upper()} — {p['pool']}: {p['apy']}% APY | TVL ${p['tvl_usd']:,.0f}")

    return state

def rank_opportunities(state: HarvesterState) -> HarvesterState:
    """Rank pools by net APY after gas deduction."""
    pools = state["available_pools"]
    amount = state["deposit_amount_usd"]

    if not pools:
        state["ranked_pools"] = []
        state["best_pool"] = {}
        state["messages"].append("No pools to rank. Try a different token or chain.")
        return state

    for pool in pools:
        gross_yield_30d = (pool["apy"] / 100 / 365) * 30 * amount
        net_yield_30d = gross_yield_30d - GAS_COST_USD
        pool["gross_yield_30d"] = round(gross_yield_30d, 2)
        pool["net_yield_30d"] = round(net_yield_30d, 2)
        pool["net_apy"] = round((net_yield_30d / amount) * (365 / 30) * 100, 2) if net_yield_30d > 0 else 0

    ranked = sorted(pools, key=lambda x: x["net_apy"], reverse=True)
    state["ranked_pools"] = ranked
    state["best_pool"] = ranked[0]

    state["messages"].append(f"WINNER: {ranked[0]['protocol'].upper()} — {ranked[0]['pool']} @ {ranked[0]['apy']}% gross / {ranked[0]['net_apy']}% net APY")
    return state

def generate_compound_plan(state: HarvesterState) -> HarvesterState:
    """Calculate compounding projections for the best pool."""
    best = state["best_pool"]
    amount = state["deposit_amount_usd"]

    if not best:
        state["compounding_plan"] = {}
        state["messages"].append("Cannot generate plan — no valid pool found.")
        return state

    apy = best["apy"] / 100
    state["compounding_plan"] = {
        "deposit_amount_usd": amount,
        "protocol": best["protocol"],
        "pool": best["pool"],
        "chain": best["chain"],
        "net_apy": best["net_apy"],
        "projections": {
            "week_1":  {"value": round(amount * (1 + apy/52), 2),   "gain": round(amount * apy/52, 2)},
            "month_1": {"value": round(amount * (1 + apy/12), 2),   "gain": round(amount * apy/12, 2)},
            "month_3": {"value": round(amount * (1 + apy/4), 2),    "gain": round(amount * apy/4, 2)},
            "year_1":  {"value": round(amount * (1 + apy), 2),      "gain": round(amount * apy, 2)},
        },
        "migration_payload": {
            "action": "DEPOSIT",
            "protocol": best["protocol"],
            "pool_id": best.get("pool_id", ""),
            "pool": best["pool"],
            "chain": best["chain"],
            "token": state["deposit_token"].upper(),
            "amount_usd": amount,
            "estimated_gas_usd": GAS_COST_USD,
            "source": "DeFiLlama"
        }
    }

    proj = state["compounding_plan"]["projections"]
    state["messages"].append(f"1W: ${proj['week_1']['value']:,.2f} | 1M: ${proj['month_1']['value']:,.2f} | 1Y: ${proj['year_1']['value']:,.2f}")
    return state

# ---- LangGraph Compilation ----
from langgraph.graph import StateGraph, END

workflow = StateGraph(HarvesterState)
workflow.add_node("scan_protocols", scan_protocols)
workflow.add_node("rank_opportunities", rank_opportunities)
workflow.add_node("generate_compound_plan", generate_compound_plan)
workflow.set_entry_point("scan_protocols")
workflow.add_edge("scan_protocols", "rank_opportunities")
workflow.add_edge("rank_opportunities", "generate_compound_plan")
workflow.add_edge("generate_compound_plan", END)
harvester_agent = workflow.compile()

executed_keys = {}

# ---- FastAPI Routes ----
@app.get("/health", tags=["System"])
async def health_check():
    return {"status": "ok", "agent": "harvester", "version": "2.0.0", "data_source": "DeFiLlama live API"}

@app.post("/api/v1/execute", response_model=AgentResponse, tags=["Agent Execution"],
          summary="Find the highest-yield DeFi pool for a given token (live data)",
          dependencies=[Depends(rate_limit)])
async def run_harvester(payload: HarvestInput):
    """Uses live DeFiLlama data to find and rank the best yield farming opportunity."""
    if payload.idempotency_key and payload.idempotency_key in executed_keys:
        return AgentResponse(status="success", agent="harvester", data={**executed_keys[payload.idempotency_key], "cached": True})
    try:
        initial_state: HarvesterState = {
            "deposit_token": payload.deposit_token,
            "deposit_amount_usd": payload.deposit_amount_usd,
            "chain": payload.chain,
            "available_pools": [], "best_pool": {}, "ranked_pools": [], "compounding_plan": {}, "messages": []
        }
        final_state = harvester_agent.invoke(initial_state)
        result = {
            "ranked_pools": final_state["ranked_pools"],
            "best_pool": final_state["best_pool"],
            "compounding_plan": final_state["compounding_plan"],
            "messages": final_state["messages"]
        }
        if payload.idempotency_key:
            executed_keys[payload.idempotency_key] = result
        return AgentResponse(status="success", agent="harvester", data=result)
    except Exception as e:
        raise HTTPException(status_code=500, detail={"error": "AGENT_EXECUTION_FAILED", "message": str(e)})

@app.get("/", include_in_schema=False)
async def serve_dashboard():
    with open("index.html", "r") as f:
        return HTMLResponse(content=f.read())

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8002)
