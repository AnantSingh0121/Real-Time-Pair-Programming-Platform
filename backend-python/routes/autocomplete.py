from fastapi import APIRouter
from pydantic import BaseModel
from services.autocomplete import AutocompleteService

router = APIRouter()
autocomplete_service = AutocompleteService()

class AutocompleteRequest(BaseModel):
    code: str
    language: str
    cursorPosition: int

@router.post("/autocomplete")
async def get_autocomplete(request: AutocompleteRequest):
    suggestions = autocomplete_service.get_suggestions(
        request.code,
        request.language,
        request.cursorPosition
    )
    return {"suggestions": suggestions}
