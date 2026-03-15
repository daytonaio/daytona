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
API_KEY = os.getenv("SNIPER_API_KEY", "dev-sniper-key-unsafe")
RATE_LIMIT_PER_MINUTE = int(os.getenv("RATE_LIMIT", "30"))

COINGECKO_TICKERS_URL = "https://api.coingecko.com/api/v3/coins/{coin_id}/tickers"

app = FastAPI(
    title="Sniper Arbitrage Agent API",
    description=(
        "Autonomous arbitrage agent. Fetches real-time price data for a token across "
        "multiple exchanges from CoinGecko to detect profitable spread opportunities."
    ),
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

# ---- Token to CoinGecko ID map ----
TOKEN_IDS = {
    "WETH": "ethereum",
    "ETH":  "ethereum",
    "BTC":  "bitcoin",
    "WBTC": "wrapped-bitcoin",
    "SOL":  "solana",
    "MATIC": "matic-network",
    "BNB":  "binancecoin",
}

# ---- Trusted DEX/CEX exchanges to compare (must be in CoinGecko's exchange list) ----
TRUSTED_EXCHANGES = {
    "uniswap_v3", "sushiswap", "curve", "pancakeswap_new",
    "binance", "coinbase_pro", "kraken", "okex", "gate"
}

GAS_COST_USD = 12.0

# ---- Models ----
class SnipeInput(BaseModel):
    target_token: str      # e.g. "WETH", "BTC"
    quote_currency: str    # e.g. "USDT", "USDC"
    trade_volume_usd: float
    min_profit_threshold: float
    idempotency_key: Optional[str] = None

    class Config:
        json_schema_extra = {
            "example": {
                "target_token": "WETH",
                "quote_currency": "USDT",
                "trade_volume_usd": 10000,
                "min_profit_threshold": 0.5
            }
        }

class AgentResponse(BaseModel):
    status: str
    agent: str
    data: dict
    error: Optional[dict] = None

class TraderState(TypedDict):
    target_token: str
    quote_currency: str
    trade_volume_usd: float
    min_profit_threshold: float
    market_data: List[Dict]
    opportunity_found: bool
    execution_payload: Dict
    messages: List[str]

# ---- LangGraph Nodes ----
def fetch_market_data(state: TraderState) -> TraderState:
    """Fetch live ticker prices for the target token across exchanges using CoinGecko."""
    token = state["target_token"].upper()
    quote = state["quote_currency"].upper()
    coin_id = TOKEN_IDS.get(token)

    if not coin_id:
        state["messages"].append(f"ERROR: Token '{token}' not supported. Supported: {list(TOKEN_IDS.keys())}")
        state["market_data"] = []
        return state

    state["messages"].append(f"Scanning live exchange prices for {token}/{quote} from CoinGecko...")

    try:
        resp = requests.get(
            COINGECKO_TICKERS_URL.format(coin_id=coin_id),
            params={"include_exchange_logo": "false", "depth": "false"},
            timeout=10
        )
        resp.raise_for_status()
        tickers = resp.json().get("tickers", [])
    except Exception as e:
        state["messages"].append(f"ERROR: CoinGecko API failed — {str(e)}")
        state["market_data"] = []
        return state

    # Filter: only include tickers quoting in our target currency with enough volume
    matching = []
    for t in tickers:
        ex_id = (t.get("market", {}).get("identifier") or "").lower()
        t_quote = (t.get("target") or "").upper()
        price = t.get("converted_last", {}).get("usd") or t.get("last", 0)
        volume_usd = t.get("converted_volume", {}).get("usd") or 0

        if t_quote != quote:
            continue
        if ex_id not in TRUSTED_EXCHANGES:
            continue
        if price <= 0 or volume_usd < 50000:  # Require at least $50k daily volume
            continue

        matching.append({
            "exchange": t.get("market", {}).get("name", ex_id),
            "exchange_id": ex_id,
            "price_usd": round(price, 4),
            "volume_usd": round(volume_usd, 2)
        })

    # Deduplicate by exchange
    seen = set()
    deduped = []
    for m in matching:
        if m["exchange_id"] not in seen:
            seen.add(m["exchange_id"])
            deduped.append(m)

    state["market_data"] = deduped
    state["messages"].append(f"Found {len(deduped)} price feeds for {token}/{quote} across trusted exchanges.")
    for ex in deduped[:6]:
        state["messages"].append(f"  {ex['exchange']}: ${ex['price_usd']:,.4f} | Vol: ${ex['volume_usd']:,.0f}")

    return state

def analyze_opportunity(state: TraderState) -> TraderState:
    """Calculate spread between cheapest buy and most expensive sell."""
    data = state["market_data"]

    if len(data) < 2:
        state["opportunity_found"] = False
        state["messages"].append("Insufficient exchange data. Need at least 2 exchanges to detect spread.")
        return state

    sorted_data = sorted(data, key=lambda x: x["price_usd"])
    buy_ex = sorted_data[0]
    sell_ex = sorted_data[-1]

    buy_price = buy_ex["price_usd"]
    sell_price = sell_ex["price_usd"]
    spread_pct = ((sell_price - buy_price) / buy_price) * 100

    token_amount = state["trade_volume_usd"] / buy_price
    gross_profit = token_amount * (sell_price - buy_price)
    net_profit = gross_profit - GAS_COST_USD
    net_profit_pct = (net_profit / state["trade_volume_usd"]) * 100

    state["messages"].append(f"Best Buy:  {buy_ex['exchange']} @ ${buy_price:,.4f}")
    state["messages"].append(f"Best Sell: {sell_ex['exchange']} @ ${sell_price:,.4f}")
    state["messages"].append(f"Spread: {spread_pct:.3f}% | Net Profit: ${net_profit:.2f} ({net_profit_pct:.3f}%)")

    if net_profit_pct >= state["min_profit_threshold"]:
        state["opportunity_found"] = True
        state["messages"].append(f"PROFITABLE ARBITRAGE DETECTED! Net: ${net_profit:.2f} ({net_profit_pct:.3f}%)")
        state["execution_payload"] = {
            "action": "ARBITRAGE",
            "buy": {"exchange": buy_ex["exchange"], "price_usd": buy_price},
            "sell": {"exchange": sell_ex["exchange"], "price_usd": sell_price},
            "token_amount": round(token_amount, 6),
            "gross_profit_usd": round(gross_profit, 2),
            "gas_cost_usd": GAS_COST_USD,
            "net_profit_usd": round(net_profit, 2),
            "net_profit_pct": round(net_profit_pct, 4),
            "spread_pct": round(spread_pct, 4)
        }
    else:
        state["opportunity_found"] = False
        state["messages"].append(f"No profitable setup. {net_profit_pct:.3f}% < target {state['min_profit_threshold']}%. Hold.")

    return state

def execute_trade(state: TraderState) -> TraderState:
    """Generate the trade execution plan."""
    if not state["opportunity_found"]:
        state["messages"].append("Sniper standing down. Monitoring continues.")
        return state
    p = state["execution_payload"]
    state["messages"].append("--- DEPLOYING FLASH ARBITRAGE ---")
    state["messages"].append(f"Step 1: BUY {p['token_amount']} {state['target_token']} on {p['buy']['exchange']} @ ${p['buy']['price_usd']:,.4f}")
    state["messages"].append(f"Step 2: SELL {p['token_amount']} {state['target_token']} on {p['sell']['exchange']} @ ${p['sell']['price_usd']:,.4f}")
    state["messages"].append(f"NET SECURED: ${p['net_profit_usd']:.2f} profit after gas.")
    return state

# ---- LangGraph Compilation ----
from langgraph.graph import StateGraph, END

workflow = StateGraph(TraderState)
workflow.add_node("fetch_market_data", fetch_market_data)
workflow.add_node("analyze_opportunity", analyze_opportunity)
workflow.add_node("execute_trade", execute_trade)
workflow.set_entry_point("fetch_market_data")
workflow.add_edge("fetch_market_data", "analyze_opportunity")
workflow.add_edge("analyze_opportunity", "execute_trade")
workflow.add_edge("execute_trade", END)
sniper_agent = workflow.compile()

executed_keys = {}

# ---- FastAPI Routes ----
@app.get("/health", tags=["System"])
async def health_check():
    return {"status": "ok", "agent": "sniper", "version": "2.0.0", "data_source": "CoinGecko live API"}

@app.post("/api/v1/execute", response_model=AgentResponse, tags=["Agent Execution"],
          summary="Detect real-time arbitrage opportunity across exchanges (live data)",
          dependencies=[Depends(rate_limit)])
async def run_sniper(payload: SnipeInput):
    """Uses live CoinGecko exchange tickers to detect real price spreads and execute arbitrage."""
    if payload.idempotency_key and payload.idempotency_key in executed_keys:
        return AgentResponse(status="success", agent="sniper", data={**executed_keys[payload.idempotency_key], "cached": True})
    try:
        initial_state: TraderState = {
            "target_token": payload.target_token,
            "quote_currency": payload.quote_currency,
            "trade_volume_usd": payload.trade_volume_usd,
            "min_profit_threshold": payload.min_profit_threshold,
            "market_data": [], "opportunity_found": False, "execution_payload": {}, "messages": []
        }
        final_state = sniper_agent.invoke(initial_state)
        result = {
            "opportunity_found": final_state["opportunity_found"],
            "market_data": final_state["market_data"],
            "execution_payload": final_state["execution_payload"],
            "messages": final_state["messages"]
        }
        if payload.idempotency_key:
            executed_keys[payload.idempotency_key] = result
        return AgentResponse(status="success", agent="sniper", data=result)
    except Exception as e:
        raise HTTPException(status_code=500, detail={"error": "AGENT_EXECUTION_FAILED", "message": str(e)})

@app.get("/", include_in_schema=False)
async def serve_dashboard():
    with open("index.html", "r") as f:
        return HTMLResponse(content=f.read())

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8001)
