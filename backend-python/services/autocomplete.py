from typing import List, Dict
import re

class AutocompleteService:
    def __init__(self):
        self.python_keywords = [
            'def', 'class', 'if', 'elif', 'else', 'for', 'while', 'try', 'except',
            'finally', 'with', 'import', 'from', 'return', 'yield', 'break', 'continue',
            'pass', 'raise', 'assert', 'lambda', 'global', 'nonlocal', 'async', 'await'
        ]
        self.python_builtins = [
            'print', 'len', 'range', 'str', 'int', 'float', 'list', 'dict', 'set',
            'tuple', 'bool', 'type', 'isinstance', 'hasattr', 'getattr', 'setattr',
            'open', 'input', 'map', 'filter', 'zip', 'enumerate', 'sorted', 'sum',
            'min', 'max', 'abs', 'round', 'all', 'any'
        ]
        self.js_keywords = [
            'function', 'const', 'let', 'var', 'if', 'else', 'for', 'while', 'do',
            'switch', 'case', 'break', 'continue', 'return', 'try', 'catch', 'finally',
            'throw', 'new', 'class', 'extends', 'import', 'export', 'async', 'await'
        ]
        self.js_builtins = [
            'console.log', 'console.error', 'console.warn', 'Array', 'Object', 'String',
            'Number', 'Boolean', 'Date', 'Math', 'JSON', 'Promise', 'setTimeout',
            'setInterval', 'fetch', 'parseInt', 'parseFloat', 'isNaN', 'isFinite'
        ]
        self.go_keywords = [
            'package', 'import', 'func', 'var', 'const', 'type', 'struct', 'interface',
            'go', 'select', 'chan', 'map', 'range', 'if', 'else', 'for', 'switch',
            'case', 'default', 'break', 'continue', 'return', 'defer'
        ]

        self.go_builtins = [
            'fmt.Println', 'fmt.Printf', 'len', 'append', 'make', 'new', 'close',
            'panic', 'recover', 'copy', 'delete'
        ]
        self.cpp_keywords = [
            'int', 'float', 'double', 'char', 'bool', 'void', 'class', 'struct',
            'namespace', 'public', 'private', 'protected', 'template', 'typename',
            'if', 'else', 'for', 'while', 'switch', 'case', 'default',
            'return', 'break', 'continue', 'new', 'delete', 'try', 'catch'
        ]

        self.cpp_builtins = [
            'std::cout', 'std::cin', 'std::string', 'std::vector', 'std::map',
            'std::unordered_map', 'std::set', 'std::sort', 'std::endl'
        ]

    def get_suggestions(self, code: str, language: str, cursor_position: int) -> List[Dict]:
        lines = code[:cursor_position].split('\n')
        current_line = lines[-1] if lines else ''

        match = re.search(r'(\w+)$', current_line)
        partial = match.group(1) if match else ''

        if language.lower() in ['python', 'py']:
            return self._get_python_suggestions(code, partial)
        elif language.lower() in ['javascript', 'js']:
            return self._get_js_suggestions(code, partial)
        elif language.lower() in ['go', 'golang']:
            return self._get_go_suggestions(code, partial)
        elif language.lower() in ['cpp', 'c++']:
            return self._get_cpp_suggestions(code, partial)

        return []
    def _get_python_suggestions(self, code: str, partial: str) -> List[Dict]:
        suggestions = []
        
        # Only provide suggestions if there's partial text to match
        if not partial:
            return suggestions

        for kw in self.python_keywords:
            if kw.startswith(partial):
                suggestions.append({'label': kw, 'kind': 'keyword', 'insertText': kw})

        for b in self.python_builtins:
            if b.startswith(partial):
                suggestions.append({'label': b, 'kind': 'function', 'insertText': b + '()'})

        for m in re.finditer(r'def\s+(\w+)', code):
            if m.group(1).startswith(partial):
                suggestions.append({'label': m.group(1), 'kind': 'function', 'insertText': m.group(1) + '()'})

        return suggestions[:20]
    def _get_js_suggestions(self, code: str, partial: str) -> List[Dict]:
        suggestions = []
        
        # Only provide suggestions if there's partial text to match
        if not partial:
            return suggestions

        for kw in self.js_keywords:
            if kw.startswith(partial):
                suggestions.append({'label': kw, 'kind': 'keyword', 'insertText': kw})

        for b in self.js_builtins:
            if b.startswith(partial):
                suggestions.append({'label': b, 'kind': 'function', 'insertText': b})

        return suggestions[:20]
    def _get_go_suggestions(self, code: str, partial: str) -> List[Dict]:
        suggestions = []
        
        # Only provide suggestions if there's partial text to match
        if not partial:
            return suggestions

        for kw in self.go_keywords:
            if kw.startswith(partial):
                suggestions.append({'label': kw, 'kind': 'keyword', 'insertText': kw})

        for b in self.go_builtins:
            if b.startswith(partial):
                suggestions.append({'label': b, 'kind': 'function', 'insertText': b})
        for m in re.finditer(r'func\s+(\w+)', code):
            name = m.group(1)
            if name.startswith(partial):
                suggestions.append({'label': name, 'kind': 'function', 'insertText': name + '()'})
        for m in re.finditer(r'type\s+(\w+)\s+struct', code):
            name = m.group(1)
            if name.startswith(partial):
                suggestions.append({'label': name, 'kind': 'struct', 'insertText': name})

        return suggestions[:20]
    def _get_cpp_suggestions(self, code: str, partial: str) -> List[Dict]:
        suggestions = []
        
        # Only provide suggestions if there's partial text to match
        if not partial:
            return suggestions

        for kw in self.cpp_keywords:
            if kw.startswith(partial):
                suggestions.append({'label': kw, 'kind': 'keyword', 'insertText': kw})

        for b in self.cpp_builtins:
            if b.startswith(partial):
                suggestions.append({'label': b, 'kind': 'function', 'insertText': b})
        for m in re.finditer(r'class\s+(\w+)', code):
            if m.group(1).startswith(partial):
                suggestions.append({'label': m.group(1), 'kind': 'class', 'insertText': m.group(1)})

        for m in re.finditer(r'[a-zA-Z_]\w*\s+(\w+)\s*\(', code):
            name = m.group(1)
            if name.startswith(partial):
                suggestions.append({'label': name, 'kind': 'function', 'insertText': name + '()'})

        return suggestions[:20]
