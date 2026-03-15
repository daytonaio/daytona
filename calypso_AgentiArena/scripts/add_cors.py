import os
import glob
import re

agent_dirs = glob.glob('*-agent')

for agent_dir in agent_dirs:
    main_py_path = os.path.join(agent_dir, 'main.py')
    if not os.path.exists(main_py_path):
        continue
        
    with open(main_py_path, 'r', encoding='utf-8') as f:
        content = f.read()
        
    if 'CORSMiddleware' in content:
        print(f'Already has CORS: {main_py_path}')
        continue
        
    print(f'Adding CORS to: {main_py_path}')
    
    # 1. Add import
    import_statement = "from fastapi.middleware.cors import CORSMiddleware\n"
    if 'from fastapi import FastAPI' in content:
        content = content.replace('from fastapi import FastAPI', 'from fastapi import FastAPI\n' + import_statement, 1)
    else:
        content = import_statement + content
        
    # 2. Add middleware after app = FastAPI(...)
    middleware_code = """

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)
"""
    # Regex to find the end of app = FastAPI(...)
    # Note: app = FastAPI(...) could span multiple lines.
    
    pattern = re.compile(r'(app\s*=\s*FastAPI\([^)]*\))')
    match = pattern.search(content)
    if match:
        content = content[:match.end()] + middleware_code + content[match.end():]
    else:
        # Fallback if no exact match (e.g., if there's no closing paren on same line or we missed something)
        # Just find app = FastAPI
        pattern2 = re.compile(r'(app\s*=\s*FastAPI\([^\n]*)')
        match2 = pattern2.search(content)
        if match2:
            content = content[:match2.end()] + middleware_code + content[match2.end():]
        else:
            print(f'COULD NOT FIND app = FastAPI in {main_py_path}')
            continue
            
    with open(main_py_path, 'w', encoding='utf-8') as f:
        f.write(content)
        
print("Done!")
