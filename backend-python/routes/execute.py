from fastapi import APIRouter
from pydantic import BaseModel
from services.executor import CodeExecutor

router = APIRouter()
executor = CodeExecutor()

class ExecuteRequest(BaseModel):
    code: str
    language: str = "python"
    stdin: str = ""

@router.post("/execute")
async def execute_code(request: ExecuteRequest):
    result = executor.execute(request.code, request.language, request.stdin)
    return result
