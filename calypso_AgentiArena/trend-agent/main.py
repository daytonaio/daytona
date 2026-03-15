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

# ---- Config ----
API_KEY        = os.getenv("TREND_API_KEY", "dev-trend-key-unsafe")
GEMINI_API_KEY = os.getenv("GEMINI_API_KEY", "")
RATE_LIMIT_PER_MINUTE = int(os.getenv("RATE_LIMIT", "10"))

if not GEMINI_API_KEY:
    raise RuntimeError("GEMINI_API_KEY environment variable is required.")

genai.configure(api_key=GEMINI_API_KEY)
model = genai.GenerativeModel("gemini-2.0-flash")

app = FastAPI(
    title="Alpha Trend Spotter API",
    description="Analyzes live CoinGecko trending search data and generates AI market sentiment reports.",
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
        raise HTTPException(status_code=401, detail={"error": "UNAUTHORIZED"})
    return api_key

request_log: Dict[str, list] = defaultdict(list)

def rate_limit(request: Request):
    client_ip = request.client.host
    now = time.time()
    request_log[client_ip] = [t for t in request_log[client_ip] if t > now - 60]
    if len(request_log[client_ip]) >= RATE_LIMIT_PER_MINUTE:
        raise HTTPException(status_code=429, detail={"error": "RATE_LIMIT_EXCEEDED"})
    request_log[client_ip].append(now)

# ---- Models ----
class TrendInput(BaseModel):
    query: Optional[str] = "Crypto Market"
    timeframe: Optional[str] = "24h"
    sources: Optional[str] = "CoinGecko, Twitter Sentiment"
    idempotency_key: Optional[str] = None

class AgentResponse(BaseModel):
    status: str
    agent: str
    data: dict

class TrendState(TypedDict):
    input_params: dict
    trending_coins: List[dict]
    trending_nfts: List[dict]
    trending_categories: List[dict]
    raw_json_str: str
    analysis_report: str
    sentiment_score: int
    key_narratives: List[str]
    messages: List[str]

# ---- LangGraph Nodes ----

def fetch_trending_data(state: TrendState) -> TrendState:
    """Fetch global trending coins, NFTs, and categories from CoinGecko."""
    state["messages"].append("Connecting to CoinGecko Search Trending API...")
    url = "https://api.coingecko.com/api/v3/search/trending"
    headers = {"accept": "application/json"}
    
    try:
        req = http_client.get(url, headers=headers, timeout=10)
        req.raise_for_status()
        data = req.json()
        
        coins = []
        for c in data.get("coins", [])[:5]:
            item = c.get("item", {})
            coins.append({
                "name": item.get("name"),
                "symbol": item.get("symbol"),
                "rank": item.get("market_cap_rank"),
                "price_btc": item.get("price_btc")
            })
            
        nfts = []
        for n in data.get("nfts", [])[:3]:
            nfts.append({
                "name": n.get("name"),
                "symbol": n.get("symbol"),
                "volume_24h": n.get("data", {}).get("h24_volume", "N/A")
            })
            
        categories = []
        for cat in data.get("categories", [])[:3]:
            categories.append({
                "name": cat.get("name"),
                "market_cap": cat.get("data", {}).get("market_cap", "N/A")
            })
            
        state["trending_coins"] = coins
        state["trending_nfts"] = nfts
        state["trending_categories"] = categories
        
        state["raw_json_str"] = f"Top Coins: {coins}\nTop NFTs: {nfts}\nTop Categories: {categories}"
        
        state["messages"].append(f"Successfully pulled {len(coins)} trending coins and {len(categories)} hot categories.")
        for c in coins:
            state["messages"].append(f"  🔥 Trending: {c['name']} ({c['symbol']}) - Rank #{c['rank']}")
            
    except Exception as e:
        state["messages"].append(f"API Error fetching trends: {str(e)}")
        state["raw_json_str"] = "Data unavailable due to API limits."
        
    return state

def generate_trend_analysis(state: TrendState) -> TrendState:
    """Use Gemini 2.0 Flash to analyze the raw trending arrays into an Alpha report."""
    state["messages"].append("Running AI Sentiment Analysis (Gemini 2.0 Flash)...")
    
    context = state["raw_json_str"]
    user_query = state["input_params"].get("query", "General Crypto Market")
    
    prompt = f"""You are Alpha Trend Spotter, an elite crypto data analyst.
Analyze the following raw trending data from CoinGecko.
Topic/Query focus: {user_query}

Data:
{context}

Respond EXACTLY in this format:

---REPORT---
[Write a punchy 2-paragraph market overview explaining WHAT is trending and WHY you think the market is focusing on these specific assets right now.]

---SCORE---
[Provide a single integer from 1 to 100 representing overall market bullish sentiment (1=extreme fear, 100=extreme greed) based strictly on these trending items]

---NARRATIVES---
• [Narrative 1: e.g. "Layer 1s making a comeback"]
• [Narrative 2: e.g. "NFT volume drying up"]
• [Narrative 3]"""

    try:
        response = model.generate_content(prompt)
        raw = response.text.strip()
        
        report = ""
        score = 50
        narratives = []
        
        if "---REPORT---" in raw:
            report_part = raw.split("---REPORT---")[1].split("---SCORE---")[0].strip()
            score_part = raw.split("---SCORE---")[1].split("---NARRATIVES---")[0].strip()
            narr_part = raw.split("---NARRATIVES---")[1].strip()
            
            report = report_part
            try:
                score = int(''.join(filter(str.isdigit, score_part)))
            except:
                score = 50
                
            for line in narr_part.split('\n'):
                line = line.strip()
                if line.startswith('•'):
                    narratives.append(line.lstrip('•').strip())
        else:
            report = raw
            
        state["analysis_report"] = report
        state["sentiment_score"] = score
        state["key_narratives"] = narratives
        
        state["messages"].append(f"Analysis complete. Sentiment Score: {score}/100.")
        state["messages"].append(f"Identified {len(narratives)} core market narratives.")
        
    except Exception as e:
        state["messages"].append(f"AI Analysis failed: {str(e)}")
        state["analysis_report"] = "Failed to generate report."
        state["sentiment_score"] = 50
        state["key_narratives"] = []

    return state

# ---- LangGraph Compilation ----
from langgraph.graph import StateGraph, END

workflow = StateGraph(TrendState)
workflow.add_node("fetch_trending_data", fetch_trending_data)
workflow.add_node("generate_trend_analysis", generate_trend_analysis)
workflow.set_entry_point("fetch_trending_data")
workflow.add_edge("fetch_trending_data", "generate_trend_analysis")
workflow.add_edge("generate_trend_analysis", END)
trend_agent = workflow.compile()

executed_keys = {}

# ---- FastAPI Routes ----
@app.get("/health", tags=["System"])
async def health_check():
    return {"status": "ok", "agent": "trend-spotter", "version": "1.0.0", "model": "gemini-2.0-flash"}

@app.post("/api/v1/execute", response_model=AgentResponse, tags=["Agent Execution"],
          dependencies=[Depends(rate_limit)])
async def run_trend_spotter(payload: TrendInput):
    """Fetches trending Web3 data and uses Gemini to synthesize narratives."""
    if payload.idempotency_key and payload.idempotency_key in executed_keys:
        return AgentResponse(status="success", agent="trend-spotter", data={**executed_keys[payload.idempotency_key], "cached": True})

    try:
        initial_state: TrendState = {
            "input_params": payload.dict(),
            "trending_coins": [], "trending_nfts": [], "trending_categories": [],
            "raw_json_str": "", "analysis_report": "", "sentiment_score": 50,
            "key_narratives": [], "messages": []
        }
        final_state = trend_agent.invoke(initial_state)
        
        result = {
            "top_coins": final_state["trending_coins"],
            "top_categories": final_state["trending_categories"],
            "analysis": final_state["analysis_report"],
            "score": final_state["sentiment_score"],
            "narratives": final_state["key_narratives"],
            "logs": final_state["messages"]
        }
        
        if payload.idempotency_key:
            executed_keys[payload.idempotency_key] = result
            
        return AgentResponse(status="success", agent="trend-spotter", data=result)
    except Exception as e:
        raise HTTPException(status_code=500, detail={"error": "AGENT_EXECUTION_FAILED", "message": str(e)})

@app.get("/", include_in_schema=False)
async def serve_dashboard():
    with open("index.html", "r") as f:
        return HTMLResponse(content=f.read())

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8006)
