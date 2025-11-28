from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from routes import execute, autocomplete
import os
from dotenv import load_dotenv

load_dotenv("../.env")

app = FastAPI(
    title="Pair Programming Execution API",
    description="Code execution and AI autocomplete service",
    version="1.0.0"
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:5173", "http://localhost:3000"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(execute.router, tags=["execution"])
app.include_router(autocomplete.router, tags=["autocomplete"])

@app.get("/")
async def root():
    return {
        "message": "Pair Programming Execution API",
        "version": "1.0.0",
        "endpoints": {
            "execute": "/execute",
            "autocomplete": "/autocomplete"
        }
    }

@app.get("/health")
async def health():
    return {"status": "healthy"}

if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PYTHON_PORT", "8001"))
    print(f"Python execution service starting on port {port}")
    print(f"Execution endpoint: http://localhost:{port}/execute")
    print(f"Autocomplete endpoint: http://localhost:{port}/autocomplete")
    uvicorn.run(app, host="0.0.0.0", port=port)
