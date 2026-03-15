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
import yfinance as yf
from datetime import datetime

# ---- Config ----
API_KEY        = os.getenv("CONSIGLIERE_API_KEY", "dev-consigliere-key-unsafe")
GEMINI_API_KEY = os.getenv("GEMINI_API_KEY", "")
RATE_LIMIT_PER_MINUTE = int(os.getenv("RATE_LIMIT", "10"))

if not GEMINI_API_KEY:
    raise RuntimeError("GEMINI_API_KEY environment variable is required.")

genai.configure(api_key=GEMINI_API_KEY)
model = genai.GenerativeModel("gemini-2.0-flash")

app = FastAPI(
    title="Consigliere Business Intelligence API",
    description=(
        "An autonomous AI business advisor. Fetches live financial data from Yahoo Finance "
        "and uses Gemini AI to generate an executive briefing, SWOT analysis, and strategic recommendations."
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
class ConsigliereInput(BaseModel):
    ticker: Optional[str] = None          # e.g. "AAPL", "TSLA", "BTC-USD"
    business_question: Optional[str] = None  # Free-form business intelligence question
    company_name: Optional[str] = None    # Used when no ticker available
    industry: Optional[str] = None        # e.g. "DeFi", "SaaS", "Electric Vehicles"
    idempotency_key: Optional[str] = None

    class Config:
        json_schema_extra = {
            "example": {
                "ticker": "AAPL",
                "business_question": "Should we build a competitive product in this space?",
                "industry": "Consumer Electronics"
            }
        }

class AgentResponse(BaseModel):
    status: str
    agent: str
    data: dict
    error: Optional[dict] = None

class ConsigliereState(TypedDict):
    ticker: Optional[str]
    business_question: Optional[str]
    company_name: Optional[str]
    industry: Optional[str]
    financial_data: Dict
    market_context: str
    executive_briefing: str
    swot_analysis: Dict
    strategic_recommendations: List[str]
    risk_flags: List[str]
    messages: List[str]

# ---- LangGraph Nodes ----

def gather_intelligence(state: ConsigliereState) -> ConsigliereState:
    """Fetch real financial data from Yahoo Finance for the given ticker."""
    ticker = state.get("ticker")
    state["financial_data"] = {}

    if ticker:
        state["messages"].append(f"Fetching live financial data for {ticker} from Yahoo Finance...")
        try:
            t = yf.Ticker(ticker)
            info = t.info

            # Pull key financial metrics
            fin = {
                "name": info.get("longName") or info.get("shortName", ticker),
                "ticker": ticker.upper(),
                "sector": info.get("sector", "N/A"),
                "industry": info.get("industry", state.get("industry", "N/A")),
                "country": info.get("country", "N/A"),
                "employees": info.get("fullTimeEmployees", "N/A"),
                "market_cap_usd": info.get("marketCap", 0),
                "revenue_ttm": info.get("totalRevenue", 0),
                "gross_margin": info.get("grossMargins", 0),
                "profit_margin": info.get("profitMargins", 0),
                "pe_ratio": info.get("trailingPE", "N/A"),
                "forward_pe": info.get("forwardPE", "N/A"),
                "debt_to_equity": info.get("debtToEquity", "N/A"),
                "free_cashflow": info.get("freeCashflow", 0),
                "beta": info.get("beta", "N/A"),
                "52w_high": info.get("fiftyTwoWeekHigh", "N/A"),
                "52w_low": info.get("fiftyTwoWeekLow", "N/A"),
                "current_price": info.get("currentPrice") or info.get("regularMarketPrice", "N/A"),
                "analyst_target": info.get("targetMeanPrice", "N/A"),
                "recommendation": info.get("recommendationKey", "N/A"),
                "description": (info.get("longBusinessSummary") or "")[:800]
            }
            state["financial_data"] = fin

            mc = fin["market_cap_usd"]
            mc_str = f"${mc/1e9:.1f}B" if mc > 1e9 else f"${mc/1e6:.1f}M" if mc else "N/A"
            state["messages"].append(f"  Company: {fin['name']} | Sector: {fin['sector']}")
            profit_str = f"{fin['profit_margin']*100:.1f}%" if fin.get("profit_margin") else "N/A"
            state["messages"].append(f"  Market Cap: {mc_str} | P/E: {fin['pe_ratio']} | Margin: {profit_str}")
            state["messages"].append(f"  Analyst Rating: {fin['recommendation'].upper()} | Target: ${fin['analyst_target']}")
        except Exception as e:
            state["messages"].append(f"Yahoo Finance error for {ticker}: {str(e)} — proceeding with general analysis.")
    else:
        company = state.get("company_name", "the target company")
        state["messages"].append(f"No ticker provided. Running general intelligence for: {company}")
        state["financial_data"] = {
            "name": company,
            "industry": state.get("industry", "General"),
            "description": ""
        }

    # Build a concise market context string for Gemini
    fd = state["financial_data"]
    ctx_parts = []
    if fd.get("name"): ctx_parts.append(f"Company: {fd['name']} ({fd.get('ticker','')})")
    if fd.get("sector"): ctx_parts.append(f"Sector: {fd['sector']} | Industry: {fd.get('industry','N/A')}")
    if fd.get("market_cap_usd"): ctx_parts.append(f"Market Cap: ${fd['market_cap_usd']/1e9:.2f}B")
    if fd.get("revenue_ttm"): ctx_parts.append(f"Revenue (TTM): ${fd['revenue_ttm']/1e9:.2f}B")
    if fd.get("profit_margin"): ctx_parts.append(f"Net Profit Margin: {fd['profit_margin']*100:.1f}%")
    if fd.get("pe_ratio") and fd.get("pe_ratio") != "N/A": ctx_parts.append(f"P/E Ratio: {fd['pe_ratio']:.1f}")
    if fd.get("beta") and fd.get("beta") != "N/A": ctx_parts.append(f"Beta: {fd['beta']}")
    if fd.get("recommendation"): ctx_parts.append(f"Analyst Consensus: {fd['recommendation'].upper()}")
    if fd.get("description"): ctx_parts.append(f"Business: {fd['description']}")
    if state.get("business_question"): ctx_parts.append(f"Business Question: {state['business_question']}")

    state["market_context"] = "\n".join(ctx_parts)
    state["messages"].append("Intelligence gathered. Passing to Gemini for analysis...")
    return state

def analyze_context(state: ConsigliereState) -> ConsigliereState:
    """Use Gemini to generate an executive briefing and SWOT analysis."""
    context = state["market_context"]
    if not context.strip():
        state["executive_briefing"] = "Insufficient data."
        state["swot_analysis"] = {}
        state["messages"].append("No data to analyze.")
        return state

    state["messages"].append("Running Gemini AI analysis (Executive Briefing + SWOT)...")

    prompt = f"""You are Consigliere, a world-class business strategist and M&A advisor.
Analyze the following company/market data and respond in this EXACT format:

---BRIEFING---
[Write a 3-paragraph executive briefing. Para 1: business overview. Para 2: financial health. Para 3: market position.]

---SWOT---
STRENGTHS:
• [strength 1]
• [strength 2]
• [strength 3]
WEAKNESSES:
• [weakness 1]
• [weakness 2]
OPPORTUNITIES:
• [opportunity 1]
• [opportunity 2]
THREATS:
• [threat 1]
• [threat 2]

Data:
{context}

Be specific, data-driven, and ruthlessly honest. No fluff."""

    try:
        response = model.generate_content(prompt)
        raw = response.text.strip()

        briefing = ""
        swot = {"strengths": [], "weaknesses": [], "opportunities": [], "threats": []}

        if "---BRIEFING---" in raw and "---SWOT---" in raw:
            briefing = raw.split("---BRIEFING---")[1].split("---SWOT---")[0].strip()
            swot_raw = raw.split("---SWOT---")[1].strip()

            current = None
            for line in swot_raw.split("\n"):
                line = line.strip()
                if line.startswith("STRENGTHS"): current = "strengths"
                elif line.startswith("WEAKNESSES"): current = "weaknesses"
                elif line.startswith("OPPORTUNITIES"): current = "opportunities"
                elif line.startswith("THREATS"): current = "threats"
                elif line.startswith("•") and current:
                    swot[current].append(line.lstrip("•").strip())
        else:
            briefing = raw

        state["executive_briefing"] = briefing
        state["swot_analysis"] = swot
        state["messages"].append("Executive briefing and SWOT analysis complete.")
    except Exception as e:
        state["messages"].append(f"Analysis failed: {str(e)}")
        state["executive_briefing"] = ""
        state["swot_analysis"] = {}

    return state

def generate_strategy(state: ConsigliereState) -> ConsigliereState:
    """Use Gemini to produce strategic recommendations and risk flags."""
    context = state["market_context"]
    question = state.get("business_question", "")

    state["messages"].append("Generating strategic recommendations and risk flags...")

    prompt = f"""You are Consigliere, a top-tier business advisor.
Based on this company data{' and the question: "' + question + '"' if question else ''}, provide:

---RECOMMENDATIONS---
1. [Specific, actionable recommendation with brief rationale]
2. [Specific, actionable recommendation with brief rationale]
3. [Specific, actionable recommendation with brief rationale]

---RISKS---
⚠ [Risk 1 — be specific and quantified where possible]
⚠ [Risk 2]
⚠ [Risk 3]

Data:
{context}

Be a trusted advisor who tells hard truths. No generic advice."""

    try:
        response = model.generate_content(prompt)
        raw = response.text.strip()

        recs = []
        risks = []

        if "---RECOMMENDATIONS---" in raw and "---RISKS---" in raw:
            rec_raw = raw.split("---RECOMMENDATIONS---")[1].split("---RISKS---")[0].strip()
            risk_raw = raw.split("---RISKS---")[1].strip()

            for line in rec_raw.split("\n"):
                line = line.strip()
                if line and line[0].isdigit():
                    recs.append(line[2:].strip() if line[1] in ".)" else line)

            for line in risk_raw.split("\n"):
                line = line.strip()
                if line.startswith("⚠"):
                    risks.append(line.lstrip("⚠").strip())
        else:
            recs = [raw]

        state["strategic_recommendations"] = recs
        state["risk_flags"] = risks
        state["messages"].append(f"Generated {len(recs)} recommendations and {len(risks)} risk flags.")
    except Exception as e:
        state["messages"].append(f"Strategy generation failed: {str(e)}")
        state["strategic_recommendations"] = []
        state["risk_flags"] = []

    return state

# ---- LangGraph Compilation ----
from langgraph.graph import StateGraph, END

workflow = StateGraph(ConsigliereState)
workflow.add_node("gather_intelligence", gather_intelligence)
workflow.add_node("analyze_context", analyze_context)
workflow.add_node("generate_strategy", generate_strategy)
workflow.set_entry_point("gather_intelligence")
workflow.add_edge("gather_intelligence", "analyze_context")
workflow.add_edge("analyze_context", "generate_strategy")
workflow.add_edge("generate_strategy", END)
consigliere_agent = workflow.compile()

executed_keys = {}

# ---- FastAPI Routes ----
@app.get("/health", tags=["System"])
async def health_check():
    return {"status": "ok", "agent": "consigliere", "version": "1.0.0", "model": "gemini-2.0-flash"}

@app.post("/api/v1/execute", response_model=AgentResponse, tags=["Agent Execution"],
          summary="Get a full business intelligence report powered by Yahoo Finance + Gemini AI",
          dependencies=[Depends(rate_limit)])
async def run_consigliere(payload: ConsigliereInput):
    """
    Fetches live financial data from Yahoo Finance for the given ticker and uses Gemini AI to
    generate an executive briefing, full SWOT analysis, strategic recommendations, and risk flags.
    """
    if not payload.ticker and not payload.business_question and not payload.company_name:
        raise HTTPException(status_code=400, detail={"error": "BAD_REQUEST", "message": "Provide at least ticker, company_name, or business_question."})

    if payload.idempotency_key and payload.idempotency_key in executed_keys:
        return AgentResponse(status="success", agent="consigliere", data={**executed_keys[payload.idempotency_key], "cached": True})

    try:
        initial_state: ConsigliereState = {
            "ticker": payload.ticker,
            "business_question": payload.business_question,
            "company_name": payload.company_name,
            "industry": payload.industry,
            "financial_data": {},
            "market_context": "",
            "executive_briefing": "",
            "swot_analysis": {},
            "strategic_recommendations": [],
            "risk_flags": [],
            "messages": []
        }
        final_state = consigliere_agent.invoke(initial_state)
        result = {
            "financial_data": final_state["financial_data"],
            "executive_briefing": final_state["executive_briefing"],
            "swot_analysis": final_state["swot_analysis"],
            "strategic_recommendations": final_state["strategic_recommendations"],
            "risk_flags": final_state["risk_flags"],
            "messages": final_state["messages"]
        }
        if payload.idempotency_key:
            executed_keys[payload.idempotency_key] = result
        return AgentResponse(status="success", agent="consigliere", data=result)
    except Exception as e:
        raise HTTPException(status_code=500, detail={"error": "AGENT_EXECUTION_FAILED", "message": str(e)})

@app.get("/", include_in_schema=False)
async def serve_dashboard():
    with open("index.html", "r") as f:
        return HTMLResponse(content=f.read())

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8005)
