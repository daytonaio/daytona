import os
import glob

agent_dirs = glob.glob('*-agent')

for agent_dir in agent_dirs:
    main_py = os.path.join(agent_dir, 'main.py')
    if not os.path.exists(main_py):
        continue

    with open(main_py, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    # Detect the corruption pattern:
    # Line N:   "from fastapi import FastAPI\n"
    # Line N+1: "from fastapi.middleware.cors import CORSMiddleware\n"
    # Line N+2: ", HTTPException, Depends, Request\n"
    #
    # Fix: merge line N and N+2 into one import, keep N+1 as separate import
    
    fixed_lines = []
    i = 0
    fixed = False
    while i < len(lines):
        line = lines[i].rstrip('\r\n')
        
        # Check if current line is "from fastapi import FastAPI" (without the rest of the imports)
        if line.strip() == 'from fastapi import FastAPI' and i + 2 < len(lines):
            next1 = lines[i+1].rstrip('\r\n').strip()
            next2 = lines[i+2].rstrip('\r\n').strip()
            
            if next1 == 'from fastapi.middleware.cors import CORSMiddleware' and next2.startswith(', '):
                # Fix: merge line i and i+2, then add the CORS import
                rest_imports = next2[2:]  # Remove leading ", "
                fixed_lines.append(f'from fastapi import FastAPI, {rest_imports}\n')
                fixed_lines.append('from fastapi.middleware.cors import CORSMiddleware\n')
                print(f"  FIXED: {main_py} - merged broken import lines")
                i += 3
                fixed = True
                continue
        
        fixed_lines.append(lines[i])
        i += 1

    if fixed:
        with open(main_py, 'w', encoding='utf-8') as f:
            f.writelines(fixed_lines)
        print(f"  Wrote fixed file: {main_py}")
    else:
        print(f"  OK (no corruption found): {main_py}")

print("\nRepair complete!")
