from fastapi import FastAPI, HTTPException, Request
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import List, Optional
import os
import uvicorn
import google.generativeai as genai
import json

GEMINI_API_KEY = os.getenv("GEMINI_API_KEY", "")
if not GEMINI_API_KEY:
    raise RuntimeError("GEMINI_API_KEY is required.")

genai.configure(api_key=GEMINI_API_KEY)
# Using Gemini 2.0 Flash for lightning-fast JSON generation
model = genai.GenerativeModel("gemini-2.0-flash")

app = FastAPI(title="Arena AI Bid Engine")

# Allow React app to call this directly
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

class AgentSpec(BaseModel):
    id: int
    name: str
    category: str
    pricePerCall: str
    description: str

class BidRequest(BaseModel):
    task: str
    category: str
    maxBudget: str
    available_agents: List[AgentSpec]

class BidResponse(BaseModel):
    agent_id: int
    bid_amount: float
    confidence_score: int
    rationale: str

class ArenaResponse(BaseModel):
    bids: List[BidResponse]

@app.post("/api/v1/arena/bids", response_model=ArenaResponse)
async def generate_bids(payload: BidRequest):
    """
    Takes the user's task prompt and the list of available agents.
    Uses Gemini AI to evaluate which 3 agents are best suited for the task,
    and generates competitive, logical bid amounts for them.
    """
    
    # Filter agents to the requested category if it's not "all"
    eligible_agents = [
        a for a in payload.available_agents 
        if payload.category == 'all' or a.category == payload.category
    ]
    
    if len(eligible_agents) == 0:
        return ArenaResponse(bids=[])

    # Construct the AI Prompt
    agents_json = json.dumps([a.dict() for a in eligible_agents], indent=2)
    
    prompt = f"""You are the central Arena intelligence for an AI Agent Marketplace.
A user has posted a task. You must evaluate the available agents and select the 3 MOST qualified agents to submit competitive bids.

User Task: "{payload.task}"
Max Budget (HLUSD): {payload.maxBudget}

Available Agents:
{agents_json}

INSTRUCTIONS:
1. Select exactly up to 3 agents that are best suited to complete this task based on their 'description' and 'category'.
2. For each selected agent, generate a 'bid_amount'. This should be based on their 'pricePerCall' base rate, but adjusted dynamically based on the complexity of the task and the user's max budget. Do NOT exceed the max budget.
3. Provide a 'confidence_score' (1-100) representing how likely this agent can successfully complete this specific task.
4. Provide a super short 'rationale' (1 sentence) on why this agent bid this amount.

Respond ONLY with a raw, valid JSON array of objects. No markdown formatting, no code blocks, just the JSON string starting with [ and ending with ].

Format required:
[
  {{ "agent_id": 1, "bid_amount": 0.045, "confidence_score": 92, "rationale": "I am a LangGraph expert explicitly built for standard token swapping." }}
]
"""

    try:
        response = model.generate_content(prompt)
        text = response.text.strip()
        
        # Clean up potential markdown formatting from Gemini
        if text.startswith("```json"):
            text = text[7:]
        if text.startswith("```"):
            text = text[3:]
        if text.endswith("```"):
            text = text[:-3]
            
        parsed_bids = json.loads(text.strip())
        
        # Format explicitly
        final_bids = []
        for b in parsed_bids:
            try:
                final_bids.append(BidResponse(
                    agent_id=int(b["agent_id"]),
                    bid_amount=float(b["bid_amount"]),
                    confidence_score=int(b["confidence_score"]),
                    rationale=str(b["rationale"])
                ))
            except Exception as e:
                print(f"Skipping malformed bid: {b} - {e}")
                
        return ArenaResponse(bids=final_bids)

    except Exception as e:
        print(f"Bid Engine Error: {str(e)}")
        # Fallback empty or simple default bid if AI fails
        if len(eligible_agents) > 0:
            fallback = eligible_agents[0]
            bd = min(float(fallback.pricePerCall), float(payload.maxBudget))
            return ArenaResponse(bids=[BidResponse(
                agent_id=fallback.id,
                bid_amount=bd,
                confidence_score=50,
                rationale="Fallback deterministic bid generated due to overloaded AI API."
            )])
        return ArenaResponse(bids=[])

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8012)
