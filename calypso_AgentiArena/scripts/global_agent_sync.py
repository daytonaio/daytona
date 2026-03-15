import os
import glob
import re

agent_dirs = glob.glob('*-agent')

def patch_file(file_path):
    print(f"Patching {file_path}...")
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()

    # 1. Add CORSMiddleware import if missing
    if 'from fastapi.middleware.cors import CORSMiddleware' not in content:
        content = re.sub(r'(from fastapi import FastAPI.*)', r'\1\nfrom fastapi.middleware.cors import CORSMiddleware', content)

    # 2. Add middleware setup if missing
    if 'app.add_middleware' not in content:
        middleware_code = """
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)
"""
        # Inject after app = FastAPI(...)
        content = re.sub(r'(app\s*=\s*FastAPI\([\s\S]*?\))', r'\1' + middleware_code, content)

    # 3. Standardize and strip API Key from the main execution route
    matches = re.findall(r'@app\.post\("(/api/v1/[^"]+)"', content)
    for route in matches:
        if "/health" in route: continue
        if route != "/api/v1/execute":
             print(f"  Changing route {route} -> /api/v1/execute")
             content = content.replace(f'@app.post("{route}"', '@app.post("/api/v1/execute"')
             
    # 4. Remove Depends(verify_api_key) from the POST routes
    content = content.replace("Depends(verify_api_key), ", "")
    content = content.replace(", Depends(verify_api_key)", "")
    content = content.replace("dependencies=[Depends(verify_api_key)]", "dependencies=[]")

    with open(file_path, 'w', encoding='utf-8') as f:
        f.write(content)

for agent_dir in agent_dirs:
    main_py = os.path.join(agent_dir, 'main.py')
    if os.path.exists(main_py):
        patch_file(main_py)

print("Global patch complete.")
